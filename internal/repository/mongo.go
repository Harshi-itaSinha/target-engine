package repository

import (
	"context"
	model "github.com/Harshi-itaSinha/target-engine/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoCampaignRepo struct {
	collection *mongo.Collection
}

func NewMongoCampaignRepo(db *mongo.Database) *MongoCampaignRepo {
	return &MongoCampaignRepo{
		collection: db.Collection("campaigns"),
	}
}

func (r *MongoCampaignRepo) FindActiveCampaigns() ([]*model.Campaign, error) {
	filter := bson.M{"status": "ACTIVE"}

	cursor, err := r.collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var results []*model.Campaign
	for cursor.Next(context.Background()) {
		var c model.Campaign
		if err := cursor.Decode(&c); err != nil {
			return nil, err
		}
		results = append(results, &c)
	}

	return results, nil
}
