package repository

import (
	"context"
	"errors"

	"github.com/cp-rektmart/aconcert-microservice/location/internal/entity"
	locationproto "github.com/cp-rektmart/aconcert-microservice/location/proto/location"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type LocationRepository struct {
	DB        *mongo.Database
	CollName  string
}

func NewLocationRepository(db *mongo.Database, collName string) *LocationRepository {
	return &LocationRepository{DB: db, CollName: collName}
}

func (r *LocationRepository) Insert(ctx context.Context, loc *locationproto.Location) (primitive.ObjectID, error) {
	doc := bson.M{
		"venue_name":     loc.VenueName,
		"city":           loc.City,
		"state_province": loc.StateProvince,
		"country":        loc.Country,
		"latitude":       loc.Latitude,
		"longitude":      loc.Longitude,
		"zones":          loc.Zones,
	}
	res, err := r.DB.Collection(r.CollName).InsertOne(ctx, doc)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return res.InsertedID.(primitive.ObjectID), nil
}

func (r *LocationRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*entity.LocationEntity, error) {
	var loc entity.LocationEntity
	err := r.DB.Collection(r.CollName).FindOne(ctx, bson.M{"_id": id}).Decode(&loc)
	if err != nil {
		return nil, err
	}
	return &loc, nil
}

func (r *LocationRepository) List(ctx context.Context) ([]*entity.LocationEntity, error) {
	cursor, err := r.DB.Collection(r.CollName).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var locs []*entity.LocationEntity
	for cursor.Next(ctx) {
		var loc entity.LocationEntity
		if err := cursor.Decode(&loc); err == nil {
			locs = append(locs, &loc)
		}
	}
	return locs, nil
}

func (r *LocationRepository) Update(ctx context.Context, id primitive.ObjectID, fields bson.M) error {
	if len(fields) == 0 {
		return errors.New("no fields to update")
	}
	_, err := r.DB.Collection(r.CollName).UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": fields})
	return err
}

func (r *LocationRepository) Delete(ctx context.Context, id primitive.ObjectID) (bool, error) {
	res, err := r.DB.Collection(r.CollName).DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return false, err
	}
	return res.DeletedCount > 0, nil
}
