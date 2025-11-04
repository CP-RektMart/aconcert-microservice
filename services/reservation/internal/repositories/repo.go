package repositories

import ("github.com/redis/go-redis/v9"
	db "github.com/cp-rektmart/aconcert-microservice/rervation/db/codegen"
)

type ReservationRepository interface{}

type ReservationImpl struct {
	db          *db.Queries
	redisClient *redis.Client
}

func NewReservationRepository(db *db.Queries, redisClient *redis.Client) *ReservationImpl {
	return &ReservationImpl{
		db:          db,
		redisClient: redisClient,
	}
}
