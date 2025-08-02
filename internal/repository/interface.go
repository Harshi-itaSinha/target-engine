package repository

import (
	"context"
	 model "github.com/Harshi-itaSinha/target-engine/internal/models"
)

// CampaignRepository defines the interface for campaign data operations
type CampaignRepository interface {
	// GetActiveCampaigns returns all active campaigns
	GetActiveCampaigns(ctx context.Context) ([]*model.Campaign, error)
	
	// GetCampaignByID returns a campaign by its ID
	GetCampaignByID(ctx context.Context, id string) (*model.Campaign, error)
	
	// CreateCampaign creates a new campaign
	CreateCampaign(ctx context.Context, campaign *model.Campaign) error
	
	// UpdateCampaign updates an existing campaign
	UpdateCampaign(ctx context.Context, campaign *model.Campaign) error
	
	// DeleteCampaign deletes a campaign by ID
	DeleteCampaign(ctx context.Context, id string) error
	
	// UpdateCampaignStatus updates the status of a campaign
	UpdateCampaignStatus(ctx context.Context, id, status string) error
}

// TargetingRuleRepository defines the interface for targeting rule data operations
type TargetingRuleRepository interface {
	// GetTargetingRules returns all targeting rules
	GetTargetingRules(ctx context.Context) ([]*model.TargetingRule, error)
	
	// GetTargetingRulesByCampaignID returns targeting rules for a specific campaign
	GetTargetingRulesByCampaignID(ctx context.Context, campaignID string) ([]*model.TargetingRule, error)
	
	// CreateTargetingRule creates a new targeting rule
	CreateTargetingRule(ctx context.Context, rule *model.TargetingRule) error
	
	// UpdateTargetingRule updates an existing targeting rule
	UpdateTargetingRule(ctx context.Context, rule *model.TargetingRule) error
	
	// DeleteTargetingRule deletes a targeting rule by ID
	DeleteTargetingRule(ctx context.Context, id int64) error
	
	// DeleteTargetingRulesByCampaignID deletes all targeting rules for a campaign
	DeleteTargetingRulesByCampaignID(ctx context.Context, campaignID string) error
}

// Repository combines all repository interfaces
type Repository interface {
	Campaign() CampaignRepository
	TargetingRule() TargetingRuleRepository
	Close() error
}

// RepositoryManager manages repository connections and transactions
type RepositoryManager interface {
	Repository
	
	// Health checks if the repository is healthy
	Health(ctx context.Context) error
	
	// Migrate runs database migrations
	Migrate(ctx context.Context) error
}