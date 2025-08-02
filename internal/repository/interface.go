package repository

import (
	"context"
	 model "github.com/Harshi-itaSinha/target-engine/internal/models"
)


type CampaignRepository interface {
	
	GetActiveCampaigns(ctx context.Context) ([]*model.Campaign, error)
	
	GetCampaignByID(ctx context.Context, id string) (*model.Campaign, error)
	
	
	CreateCampaign(ctx context.Context, campaign *model.Campaign) error
	
	UpdateCampaign(ctx context.Context, campaign *model.Campaign) error
	
	DeleteCampaign(ctx context.Context, id string) error
	
	
	UpdateCampaignStatus(ctx context.Context, id, status string) error
}

type TargetingRuleRepository interface {
	
	GetTargetingRules(ctx context.Context) ([]*model.TargetingRule, error)
	
	
	GetTargetingRulesByCampaignID(ctx context.Context, campaignID string) ([]*model.TargetingRule, error)
	
	
	CreateTargetingRule(ctx context.Context, rule *model.TargetingRule) error
	
	UpdateTargetingRule(ctx context.Context, rule *model.TargetingRule) error
	

	DeleteTargetingRule(ctx context.Context, id int64) error
	
	
	DeleteTargetingRulesByCampaignID(ctx context.Context, campaignID string) error
}

type Repository interface {
	Campaign() CampaignRepository
	TargetingRule() TargetingRuleRepository
	Close() error
}

type RepositoryManager interface {
	Repository

	Health(ctx context.Context) error
	
	Migrate(ctx context.Context) error
}