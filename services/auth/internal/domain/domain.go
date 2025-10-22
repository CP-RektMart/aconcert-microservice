package domain

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/cp-rektmart/aconcert-microservice/auth/config"
	"github.com/cp-rektmart/aconcert-microservice/auth/internal/entities"
	"github.com/cp-rektmart/aconcert-microservice/auth/internal/errs"
	"github.com/cp-rektmart/aconcert-microservice/auth/internal/repositories"
	"github.com/cp-rektmart/aconcert-microservice/pkg/jwt"
	"github.com/google/uuid"
	"google.golang.org/api/idtoken"
)

type AuthDomain interface {
	LoginWithGoogle(ctx context.Context, idToken string) (entities.Token, entities.User, bool, error)
	RefreshToken(ctx context.Context, token string) (entities.Token, error)
	Logout(ctx context.Context, userID uuid.UUID) error
	GetUser(ctx context.Context, userID uuid.UUID) (entities.User, error)
}

type AuthDomainImpl struct {
	repo         repositories.AuthRepository
	jwtConfig    *jwt.Config
	googleConfig *config.GoogleConfig
}

func New(repo repositories.AuthRepository, jwtConfig *jwt.Config, googleConfig *config.GoogleConfig) AuthDomain {
	return &AuthDomainImpl{
		repo:         repo,
		jwtConfig:    jwtConfig,
		googleConfig: googleConfig,
	}
}

func (d *AuthDomainImpl) generateAuthToken(ctx context.Context, user entities.User) (accessToken, refreshToken string, exp int64, err error) {
	cachedToken, accessToken, refreshToken, exp, err := jwt.GenerateTokenPair(user.ID, d.jwtConfig.AccessTokenSecret, d.jwtConfig.RefreshTokenSecret, d.jwtConfig.AccessTokenExpire, d.jwtConfig.RefreshTokenExpire)
	if err != nil {
		return "", "", 0, errors.Wrap(err, "failed to generate token pair")
	}

	if err := d.repo.SetUserAuthToken(ctx, user.ID, *cachedToken); err != nil {
		return "", "", 0, errors.Wrap(err, "failed to set user auth token")
	}

	return accessToken, refreshToken, exp, nil
}

func (d *AuthDomainImpl) LoginWithGoogle(ctx context.Context, idToken string) (entities.Token, entities.User, bool, error) {

	var user entities.User
	var isNewUser bool

	payload, err := idtoken.Validate(ctx, idToken, d.googleConfig.ClientID)
	if err != nil {
		return entities.Token{}, entities.User{}, false, errors.Wrap(err, "failed to validate id token")
	}
	email, ok := payload.Claims["email"].(string)
	if !ok {
		return entities.Token{}, entities.User{}, false, errors.New("email not found in id token")
	}
	firstname, ok := payload.Claims["given_name"].(string)
	if !ok {
		return entities.Token{}, entities.User{}, false, errors.New("firstname not found in id token")
	}
	lastname, ok := payload.Claims["family_name"].(string)
	if !ok {
		return entities.Token{}, entities.User{}, false, errors.New("lastname not found in id token")
	}
	profileImage, ok := payload.Claims["picture"].(string)
	if !ok {
		return entities.Token{}, entities.User{}, false, errors.New("lastname not found in id token")
	}

	user, err = d.repo.GetUserByProviderEmail(ctx, entities.ProviderGoogle, email)
	if errors.Is(err, errs.ErrNotFound) {
		user, err = d.repo.CreateUser(ctx, entities.CreateUserInput{
			Provider:     entities.ProviderGoogle,
			Email:        email,
			Firstname:    firstname,
			Lastname:     lastname,
			ProfileImage: profileImage,
			Role:         entities.UserRoleUser,
		})
		if err != nil {
			return entities.Token{}, entities.User{}, false, errors.Wrap(err, "failed to create user")
		}

		isNewUser = true
		err = nil
	}
	if err != nil {
		return entities.Token{}, entities.User{}, false, errors.Wrap(err, "failed to get user by provider email")
	}

	accessToken, refreshToken, exp, err := d.generateAuthToken(ctx, user)
	if err != nil {
		return entities.Token{}, entities.User{}, false, errors.Wrap(err, "failed to generate auth token")
	}

	token := entities.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Exp:          exp,
	}

	return token, user, isNewUser, nil
}

func (d *AuthDomainImpl) RefreshToken(ctx context.Context, token string) (entities.Token, error) {
	claims, err := jwt.ParseToken(token, d.jwtConfig.RefreshTokenSecret)
	if err != nil {
		return entities.Token{}, errors.Wrap(err, "failed to parse refresh token")
	}

	cachedToken, err := d.repo.GetUserAuthToken(ctx, claims.ID)
	if err != nil {
		return entities.Token{}, errors.Wrap(err, "failed to get user auth token")
	}

	if err := jwt.ValidateToken(cachedToken, claims, true); err != nil {
		return entities.Token{}, errors.Wrap(err, "failed to validate refresh token")
	}

	user, err := d.repo.GetUser(ctx, claims.ID)
	if err != nil {
		return entities.Token{}, errors.Wrap(err, "user not found")
	}

	accessToken, refreshToken, exp, err := d.generateAuthToken(ctx, user)
	if err != nil {
		return entities.Token{}, errors.Wrap(err, "failed to generate auth token")
	}

	return entities.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Exp:          exp,
	}, nil
}

func (d *AuthDomainImpl) Logout(ctx context.Context, userID uuid.UUID) error {
	err := d.repo.DeleteUserAuthToken(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "failed to delete user auth token")
	}
	return nil
}

func (d *AuthDomainImpl) GetUser(ctx context.Context, userID uuid.UUID) (entities.User, error) {
	user, err := d.repo.GetUser(ctx, userID)
	if err != nil {
		return entities.User{}, errors.Wrap(err, "failed to get user by id")
	}

	return user, nil
}
