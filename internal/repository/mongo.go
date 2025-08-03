package repository

import (
	"context"
	"errors"
	"fmt"
	"log"

	models "github.com/Harshi-itaSinha/target-engine/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MongoCampaignRepo struct {
	collection *mongo.Collection
}

func NewMongoCampaignRepo(db *mongo.Database) *MongoCampaignRepo {
	return &MongoCampaignRepo{
		collection: db.Collection("target-engine"),
	}
}

const (
	CollectionCampaigns      = "campaigns"
	CollectionTargetingRules = "targeting_rules"
	CollectionActiveCampaign = "active_targeting_rules" // pre-computed
)

type RepositoryImpl struct {
	client   *mongo.Client
	database *mongo.Database
}

// NewRepository creates a new RepositoryImpl with an injected MongoDB collection.
func NewRepository(database *mongo.Database, client *mongo.Client) *RepositoryImpl {
	if database == nil {
		panic("database cannot be nil")
	}
	return &RepositoryImpl{
		database: database,
		client:   client,
	}
}

func (r *RepositoryImpl) GetCollection(name string) *mongo.Collection {
	if r.client == nil {
		panic("MongoDB client is not initialized")
	}
	if r.database == nil {
		panic("MongoDB database name is not set")
	}
	return r.database.Collection(name)
}

// Campaign returns the CampaignRepository implementation.
func (r *RepositoryImpl) Campaign() CampaignRepository {
	return r
}

// TargetingRule returns the TargetingRuleRepository implementation.
func (r *RepositoryImpl) TargetingRule() TargetingRuleRepository {
	return r
}

// Close closes the MongoDB client (noop if not set, assuming collection is injected).
func (r *RepositoryImpl) Close() error {
	// Note: Client is not managed here since collection is injected. Close should be handled by the caller (e.g., config).
	return nil
}

// Health checks the MongoDB connection health.
func (r *RepositoryImpl) Health(ctx context.Context) error {
	if r.client == nil {
		return errors.New("MongoDB client not initialized")
	}
	return r.client.Ping(ctx, nil)
}

// Migrate sets up the MongoDB collection with indexes (simplified migration).
func (r *RepositoryImpl) Migrate(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "type", Value: 1}, {Key: "dimension", Value: 1}, {Key: "value", Value: 1}}},
		{Keys: bson.D{{Key: "type", Value: 1}, {Key: "campaign_id", Value: 1}}},
		{Keys: bson.D{{Key: "campaign_details.status", Value: 1}}},
	}
	_, err := r.GetCollection(CollectionCampaigns).Indexes().CreateMany(ctx, indexes)
	return err
}

func (r *MongoCampaignRepo) FindActiveCampaigns() ([]*models.Campaign, error) {
	filter := bson.M{"status": "ACTIVE"}

	cursor, err := r.collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var results []*models.Campaign
	for cursor.Next(context.Background()) {
		var c models.Campaign
		if err := cursor.Decode(&c); err != nil {
			return nil, err
		}
		results = append(results, &c)
	}

	return results, nil
}

// CampaignRepository implementation
func (r *RepositoryImpl) GetActiveCampaigns(ctx context.Context) ([]*models.Campaign, error) {
	filter := bson.M{"type": "rule", "campaign_details.status": "ACTIVE"}
	cursor, err := r.GetCollection(CollectionCampaigns).Find(ctx, filter, options.Find().SetProjection(bson.M{
		"campaign_id":            1,
		"campaign_details.name":  1,
		"campaign_details.image": 1,
		"campaign_details.cta":   1,
	}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	campaigns := make([]*models.Campaign, 0)
	for cursor.Next(ctx) {
		var result struct {
			CampaignID string `bson:"campaign_id"`
			Details    struct {
				Name  string `bson:"name"`
				Image string `bson:"image"`
				CTA   string `bson:"cta"`
			} `bson:"campaign_details"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		campaigns = append(campaigns, &models.Campaign{
			ID:    result.CampaignID,
			Name:  result.Details.Name,
			Image: result.Details.Image,
			CTA:   result.Details.CTA,
		})
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return campaigns, nil
}

func (r *RepositoryImpl) GetCampaignByID(ctx context.Context, id string) (*models.Campaign, error) {
	filter := bson.M{"type": "rule", "campaign_id": id, "campaign_details.status": "ACTIVE"}
	var result struct {
		CampaignID string `bson:"campaign_id"`
		Details    struct {
			Name  string `bson:"name"`
			Image string `bson:"image"`
			CTA   string `bson:"cta"`
		} `bson:"campaign_details"`
	}
	if err := r.GetCollection(CollectionCampaigns).FindOne(ctx, filter).Decode(&result); err != nil {
		return nil, err
	}
	return &models.Campaign{
		ID:    result.CampaignID,
		Name:  result.Details.Name,
		Image: result.Details.Image,
		CTA:   result.Details.CTA,
	}, nil
}

func (r *RepositoryImpl) GetCampaignsByIDs(ctx context.Context, ids []string) ([]*models.Campaign, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	filter := bson.M{"cid": bson.M{"$in": ids}}
	cursor, err := r.GetCollection(CollectionCampaigns).Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch campaigns by cid: %w", err)
	}
	defer cursor.Close(ctx)

	campaigns := []*models.Campaign{}
	if err := cursor.All(ctx, &campaigns); err != nil {
		return nil, fmt.Errorf("failed to decode campaigns: %w", err)
	}

	log.Printf("Found %d campaigns using cid field", len(campaigns))
	
	if len(campaigns) == 0 {
		return nil, nil
	}
    

	return campaigns, nil
	


}

func (r *RepositoryImpl) CreateCampaign(ctx context.Context, campaign *models.Campaign) error {
	// Assuming campaign includes rules; create rule documents
	// for _, rule := range campaign.Rules {
	// 	doc := bson.M{
	// 		"type":        "rule",
	// 		"campaign_id": campaign.ID,
	// 		"dimension":   rule.Dimension,
	// 		"include":     rule.Include,
	// 		"exclude":     rule.Exclude,
	// 		"campaign_details": bson.M{
	// 			"name":   campaign.Name,
	// 			"image":  campaign.Image,
	// 			"cta":    campaign.CTA,
	// 			"status": campaign.Status,
	// 		},
	// 	}
	// 	if _, err := r.GetCollection(CollectionCampaigns).InsertOne(ctx, doc); err != nil {
	// 		return err
	// 	}
	// 	// Update mappings (simplified; use Change Streams in production)
	// 	if err := r.updateMappings(ctx, campaign.ID, rule); err != nil {
	// 		return err
	// 	}
	// }
	// return nil
	return nil
}

func buildMappingMatchPipeline(dimensions []models.Dimension) mongo.Pipeline {


	//Build filters for each dimension-value pair
	filters := bson.A{}
	for _, d := range dimensions {
		dimensionFilter := bson.D{
			{Key: "dimension", Value: d.Name}, // Match specific dimension
			{Key: "$or", Value: bson.A{
				bson.D{
					{Key: "type", Value: "include"},
					{Key: "values", Value: bson.D{{Key: "$in", Value: bson.A{d.Value}}}},
				},
				bson.D{
					{Key: "type", Value: "exclude"},
					{Key: "values", Value: bson.D{{Key: "$nin", Value: bson.A{d.Value}}}},
				},
				bson.D{{Key: "type", Value: primitive.Null{}}}, // Handle null type
			}},
		}

		
		filters = append(filters, dimensionFilter)
	}

	
	pipeline := mongo.Pipeline{
		//Stage 1: Match documents for any dimension
		{{Key: "$match", Value: bson.D{{Key: "$or", Value: filters}}}},
		
		//Stage 2: Group by campaign_id and collect covered dimensions
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$campaign_id"},
			{Key: "coveredDimensions", Value: bson.D{{Key: "$addToSet", Value: "$dimension"}}},
		}}},
		
		// Stage 3: Filter campaigns that cover all required dimensions
		{{Key: "$match", Value: bson.D{
			{Key: "coveredDimensions", Value: bson.D{{Key: "$size", Value: len(dimensions)}}},
		}}},
		
		//Stage 4: Project the final result
		{{Key: "$project", Value: bson.D{
			{Key: "campaign_id", Value: "$_id"},
			{Key: "_id", Value: 0},
		}}},
	}

	return pipeline

}

func fetchValidCampaignIDs(ctx context.Context, collection *mongo.Collection, pipeline mongo.Pipeline) ([]string, error) {
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var campaignIDs []string
	for cursor.Next(ctx) {
		
		var result struct {
			ID string `bson:"campaign_id"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		campaignIDs = append(campaignIDs, result.ID)
	}
	
	return campaignIDs, nil
}

func (r *RepositoryImpl) GetMatchingCampaignIDs(ctx context.Context, dimensions []models.Dimension) ([]string, error) {
	collection := r.GetCollection(CollectionActiveCampaign)
	pipeline := buildMappingMatchPipeline(dimensions)

	allCampaigns, err := fetchValidCampaignIDs(ctx, collection, pipeline)
	if err != nil {
		return nil, err
	}
	return allCampaigns, nil

}

func (r *RepositoryImpl) UpdateCampaign(ctx context.Context, campaign *models.Campaign) error {
	// Delete existing rules and recreate
	if err := r.DeleteCampaign(ctx, campaign.ID); err != nil {
		return err
	}
	return r.CreateCampaign(ctx, campaign)
}

func (r *RepositoryImpl) DeleteCampaign(ctx context.Context, id string) error {
	// _, err := r.GetCollection(CollectionCampaigns).DeleteMany(ctx, bson.M{"type": "rule", "campaign_id": id})
	// if err != nil {
	// 	return err
	// }
	// // Clean up mappings (simplified; use Change Streams in production)
	// return r.GetCollection(CollectionCampaigns).DeleteMany(ctx, bson.M{"type": "mapping", "valid_campaigns": id})
	return nil
}

func (r *RepositoryImpl) UpdateCampaignStatus(ctx context.Context, id, status string) error {
	update := bson.M{"$set": bson.M{"campaign_details.status": status}}
	_, err := r.GetCollection(CollectionCampaigns).UpdateMany(ctx, bson.M{"type": "rule", "campaign_id": id}, update)
	return err
}

func (r *RepositoryImpl) GetTargetingRules(ctx context.Context) ([]*models.TargetingRule, error) {
	// filter := bson.M{"type": "rule"}
	// cursor, err := r.GetCollection(CollectionCampaigns).Find(ctx, filter)
	// if err != nil {
	// 	return nil, err
	// }
	// defer cursor.Close(ctx)

	// rules := make([]*models.TargetingRule, 0)
	// for cursor.Next(ctx) {
	// 	var result struct {
	// 		CampaignID string   `bson:"campaign_id"`
	// 		Dimension  string   `bson:"dimension"`
	// 		Include    []string `bson:"include"`
	// 		Exclude    []string `bson:"exclude"`
	// 	}
	// 	if err := cursor.Decode(&result); err != nil {
	// 		return nil, err
	// 	}
	// 	rules = append(rules, &models.TargetingRule{
	// 		CampaignID: result.CampaignID,
	// 		Dimension:  result.Dimension,
	// 		Include:    result.Include,
	// 		Exclude:    result.Exclude,
	// 	})
	// }
	// if err := cursor.Err(); err != nil {
	// 	return nil, err
	// }
	// return rules, nil
	return nil, nil
}

func (r *RepositoryImpl) GetTargetingRulesByCampaignID(ctx context.Context, campaignID string) ([]*models.TargetingRule, error) {
	// filter := bson.M{"type": "rule", "campaign_id": campaignID}
	// cursor, err := r.GetCollection(CollectionCampaigns).Find(ctx, filter)
	// if err != nil {
	// 	return nil, err
	// }
	// defer cursor.Close(ctx)

	// rules := make([]*models.TargetingRule, 0)
	// for cursor.Next(ctx) {
	// 	var result struct {
	// 		CampaignID string   `bson:"campaign_id"`
	// 		Dimension  string   `bson:"dimension"`
	// 		Include    []string `bson:"include"`
	// 		Exclude    []string `bson:"exclude"`
	// 	}
	// 	if err := cursor.Decode(&result); err != nil {
	// 		return nil, err
	// 	}
	// 	rules = append(rules, &models.TargetingRule{
	// 		CampaignID: result.CampaignID,
	// 		Dimension:  result.Dimension,
	// 		Include:    result.Include,
	// 		Exclude:    result.Exclude,
	// 	})
	// }
	// if err := cursor.Err(); err != nil {
	// 	return nil, err
	// }
	// return rules, nil
	return nil, nil
}

func (r *RepositoryImpl) CreateTargetingRule(ctx context.Context, rule *models.TargetingRule) error {
	// doc := bson.M{
	// 	"type":        "rule",
	// 	"campaign_id": rule.CampaignID,
	// 	"dimension":   rule.Dimension,
	// 	"include":     rule.Include,
	// 	"exclude":     rule.Exclude,
	// 	"created_at":  time.Now().UTC(),
	// 	"updated_at":  time.Now().UTC(),
	// }

	// if _, err := r.GetCollection(CollectionCampaigns).InsertOne(ctx, doc); err != nil {
	// 	return err
	// }
	// // Update mappings (simplified; use Change Streams in production)
	// return r.updateMappings(ctx, rule.CampaignID, rule)
	return nil
}

func (r *RepositoryImpl) UpdateTargetingRule(ctx context.Context, rule *models.TargetingRule) error {
	// filter := bson.M{"type": "rule", "campaign_id": rule.CampaignID, "dimension": rule.Dimension}
	// update := bson.M{"$set": bson.M{
	// 	"include": rule.Include,
	// 	"exclude": rule.Exclude,
	// }}
	// if _, err := r.collection.UpdateOne(ctx, filter, update); err != nil {
	// 	return err
	// }
	// // Update mappings (simplified; use Change Streams in production)
	// return r.updateMappings(ctx, rule.CampaignID, rule)
	return nil
}

func (r *RepositoryImpl) DeleteTargetingRule(ctx context.Context, id int64) error {
	// Assuming id is a placeholder; adjust if it's a different field (e.g., _id as ObjectID)
	return fmt.Errorf("DeleteTargetingRule not implemented: id type mismatch")
}

func (r *RepositoryImpl) DeleteTargetingRulesByCampaignID(ctx context.Context, campaignID string) error {
	// _, err := r.GetCollection(CollectionCampaigns).DeleteMany(ctx, bson.M{"type": "rule", "campaign_id": campaignID})
	// if err != nil {
	// 	return err
	// }
	// // Clean up mappings (simplified; use Change Streams in production)
	// return r.GetCollection(CollectionCampaigns).DeleteMany(ctx, bson.M{"type": "mapping", "valid_campaigns": campaignID})
	return nil
}

// updateMappings updates pre-aggregated mappings (simplified implementation).
func (r *RepositoryImpl) updateMappings(ctx context.Context, campaignID string, rule *models.TargetingRule) error {
	// This is a simplified version; in production, use Change Streams to recompute all mappings
	// if rule.Include != nil {
	// 	for _, value := range rule.Include {
	// 		filter := bson.M{"type": "mapping", "dimension": rule.Dimension, "value": value}
	// 		update := bson.M{"$addToSet": bson.M{"valid_campaigns": campaignID}}
	// 		_, err := r.GetCollection(CollectionCampaigns).UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// }
	// if rule.Exclude != nil {
	// 	for _, value := range rule.Exclude {
	// 		filter := bson.M{"type": "mapping", "dimension": rule.Dimension, "value": value}
	// 		update := bson.M{"$pull": bson.M{"valid_campaigns": campaignID}}
	// 		_, err := r.GetCollection(CollectionCampaigns).UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// }
	return nil
}
