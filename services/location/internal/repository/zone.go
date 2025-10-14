package repository

import (
	"context"

	"github.com/cockroachdb/errors/grpc/status"
	locationproto "github.com/cp-rektmart/aconcert-microservice/location/proto/location"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
)
func (r *LocationRepository) AddZone(ctx context.Context, locID primitive.ObjectID, zone *locationproto.Zone) error {
	collection := r.DB.Collection(r.CollName)

	// 1. Check if location exists
	var loc struct {
		Zones []*locationproto.Zone `bson:"zones"`
	}
	err := collection.FindOne(ctx, bson.M{"_id": locID}).Decode(&loc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return status.Errorf(codes.NotFound, "location not found")
		}
		return status.Errorf(codes.Internal, "failed to fetch location: %v", err)
	}

	// 2. Check if zone_number already exists
	for _, z := range loc.Zones {
		if z.ZoneNumber == zone.ZoneNumber {
			return status.Errorf(codes.InvalidArgument, "zone_number %d already exists", zone.ZoneNumber)
		}
	}

	// 3. Add the new zone
	update := bson.M{"$push": bson.M{"zones": zone}}
	_, err = collection.UpdateOne(ctx, bson.M{"_id": locID}, update)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to add zone: %v", err)
	}

	return nil
}


func (r *LocationRepository) RemoveZone(ctx context.Context, locID primitive.ObjectID, zoneNumber int32) error {
	collection := r.DB.Collection(r.CollName)
	update := bson.M{"$pull": bson.M{"zones": bson.M{"zone_number": zoneNumber}}}
	res, err := collection.UpdateOne(ctx, bson.M{"_id": locID}, update)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to remove zone: %v", err)
	}
	if res.MatchedCount == 0 {
		return status.Errorf(codes.NotFound, "location not found")
	}
	return nil
}

func (r *LocationRepository) UpdateZone(ctx context.Context, locID primitive.ObjectID, zone *locationproto.Zone) error {
	if zone == nil {
		return status.Errorf(codes.InvalidArgument, "zone is nil")
	}
	updateFields := bson.M{}
	if zone.ZoneName != "" {
		updateFields["zones.$.zone_name"] = zone.ZoneName
	}
	if zone.Capacity != 0 {
		updateFields["zones.$.capacity"] = zone.Capacity
	}
	if zone.Price != 0 {
		updateFields["zones.$.price"] = zone.Price
	}
	if zone.Reserved {
		updateFields["zones.$.reserved"] = zone.Reserved
	}
	if len(zone.Exclusive) > 0 {
		updateFields["zones.$.exclusive"] = zone.Exclusive
	}

	if len(updateFields) == 0 {
		return status.Errorf(codes.InvalidArgument, "no fields to update")
	}

	collection := r.DB.Collection(r.CollName)
	filter := bson.M{"_id": locID, "zones.zone_number": zone.ZoneNumber}
	update := bson.M{"$set": updateFields}

	res, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to update zone: %v", err)
	}
	if res.MatchedCount == 0 {
		return status.Errorf(codes.NotFound, "zone_number %d not found", zone.ZoneNumber)
	}
	return nil
}


