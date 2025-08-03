package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	model "github.com/Harshi-itaSinha/target-engine/internal/models"
)

type MemoryRepository struct {
	campaigns      map[string]*model.Campaign
	targetingRules map[string][]*model.TargetingRule // keyed by campaign_id
	rulesByID      map[int64]*model.TargetingRule
	mutex          sync.RWMutex
	nextRuleID     int64
}

func NewMemoryRepository() *MemoryRepository {
	repo := &MemoryRepository{
		campaigns:      make(map[string]*model.Campaign),
		targetingRules: make(map[string][]*model.TargetingRule),
		rulesByID:      make(map[int64]*model.TargetingRule),
		nextRuleID:     1,
	}

	repo.initializeSampleData()

	return repo
}

func (r *MemoryRepository) Campaign() CampaignRepository {
	return r
}

func (r *MemoryRepository) TargetingRule() TargetingRuleRepository {
	return r
}

func (r *MemoryRepository) Close() error {
	return nil
}

func (r *MemoryRepository) Health(ctx context.Context) error {
	return nil
}

func (r *MemoryRepository) Migrate(ctx context.Context) error {
	return nil
}

func (r *MemoryRepository) GetActiveCampaigns(ctx context.Context) ([]*model.Campaign, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var activeCampaigns []*model.Campaign
	for _, campaign := range r.campaigns {
		if campaign.IsActive() {
			activeCampaigns = append(activeCampaigns, campaign)
		}
	}

	return activeCampaigns, nil
}

func (r *MemoryRepository) GetMatchingCampaignIDs(ctx context.Context, dimensions []model.Dimension) ([]string, error) {
	return nil, nil
}

func (r *MemoryRepository) GetCampaignByID(ctx context.Context, id string) (*model.Campaign, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	campaign, exists := r.campaigns[id]
	if !exists {
		return nil, fmt.Errorf("campaign with ID %s not found", id)
	}

	return campaign, nil
}

func (r *MemoryRepository) GetCampaignsByIDs(ctx context.Context, ids []string) ([]*model.Campaign, error) {
	return nil, nil
}

func (r *MemoryRepository) CreateCampaign(ctx context.Context, campaign *model.Campaign) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.campaigns[campaign.ID]; exists {
		return fmt.Errorf("campaign with ID %s already exists", campaign.ID)
	}

	campaign.CreatedAt = time.Now()
	campaign.UpdatedAt = time.Now()
	r.campaigns[campaign.ID] = campaign

	return nil
}

func (r *MemoryRepository) UpdateCampaign(ctx context.Context, campaign *model.Campaign) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.campaigns[campaign.ID]; !exists {
		return fmt.Errorf("campaign with ID %s not found", campaign.ID)
	}

	campaign.UpdatedAt = time.Now()
	r.campaigns[campaign.ID] = campaign

	return nil
}

func (r *MemoryRepository) DeleteCampaign(ctx context.Context, id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.campaigns[id]; !exists {
		return fmt.Errorf("campaign with ID %s not found", id)
	}

	delete(r.campaigns, id)
	delete(r.targetingRules, id) // Also delete associated targeting rules

	return nil
}

func (r *MemoryRepository) UpdateCampaignStatus(ctx context.Context, id, status string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	campaign, exists := r.campaigns[id]
	if !exists {
		return fmt.Errorf("campaign with ID %s not found", id)
	}

	campaign.Status = status
	campaign.UpdatedAt = time.Now()

	return nil
}

// Targeting Rule Repository Methods

// GetTargetingRules returns all targeting rules
func (r *MemoryRepository) GetTargetingRules(ctx context.Context) ([]*model.TargetingRule, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var allRules []*model.TargetingRule
	for _, rules := range r.targetingRules {
		allRules = append(allRules, rules...)
	}

	return allRules, nil
}

func (r *MemoryRepository) GetTargetingRulesByCampaignID(ctx context.Context, campaignID string) ([]*model.TargetingRule, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	rules, exists := r.targetingRules[campaignID]
	if !exists {
		return []*model.TargetingRule{}, nil
	}

	return rules, nil
}

func (r *MemoryRepository) CreateTargetingRule(ctx context.Context, rule *model.TargetingRule) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	rule.ID = r.nextRuleID
	r.nextRuleID++
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	r.targetingRules[rule.CampaignID] = append(r.targetingRules[rule.CampaignID], rule)
	r.rulesByID[rule.ID] = rule

	return nil
}

func (r *MemoryRepository) UpdateTargetingRule(ctx context.Context, rule *model.TargetingRule) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	existingRule, exists := r.rulesByID[rule.ID]
	if !exists {
		return fmt.Errorf("targeting rule with ID %d not found", rule.ID)
	}

	rule.UpdatedAt = time.Now()

	r.rulesByID[rule.ID] = rule

	rules := r.targetingRules[existingRule.CampaignID]
	for i, r := range rules {
		if r.ID == rule.ID {
			rules[i] = rule
			break
		}
	}

	return nil
}

func (r *MemoryRepository) DeleteTargetingRule(ctx context.Context, id int64) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.rulesByID, id)

	return nil
}

// DeleteTargetingRulesByCampaignID deletes all targeting rules for a campaign
func (r *MemoryRepository) DeleteTargetingRulesByCampaignID(ctx context.Context, campaignID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	rules, exists := r.targetingRules[campaignID]
	if !exists {
		return nil
	}

	for _, rule := range rules {
		delete(r.rulesByID, rule.ID)
	}

	delete(r.targetingRules, campaignID)

	return nil
}

func (r *MemoryRepository) initializeSampleData() {
	now := time.Now()

	campaigns := []*model.Campaign{
		{
			ID:        "spotify",
			Name:      "Spotify - Music for everyone",
			Image:     "https://somelink",
			CTA:       "Download",
			Status:    model.StatusActive,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "duolingo",
			Name:      "Duolingo: Best way to learn",
			Image:     "https://somelink2",
			CTA:       "Install",
			Status:    model.StatusActive,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "subwaysurfer",
			Name:      "Subway Surfer",
			Image:     "https://somelink3",
			CTA:       "Play",
			Status:    model.StatusActive,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	targetingRules := []*model.TargetingRule{
		{
			ID:             1,
			CampaignID:     "spotify",
			IncludeCountry: []string{"US", "Canada"},
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             2,
			CampaignID:     "duolingo",
			IncludeOS:      []string{"Android", "iOS"},
			ExcludeCountry: []string{"US"},
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:         3,
			CampaignID: "subwaysurfer",
			IncludeOS:  []string{"Android"},
			IncludeApp: []string{"com.gametion.ludokinggame"},
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	for _, campaign := range campaigns {
		r.campaigns[campaign.ID] = campaign
	}

	r.nextRuleID = 4
	for _, rule := range targetingRules {
		r.targetingRules[rule.CampaignID] = append(r.targetingRules[rule.CampaignID], rule)
		r.rulesByID[rule.ID] = rule
	}
}
