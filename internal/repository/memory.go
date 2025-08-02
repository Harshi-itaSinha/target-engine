package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	model "github.com/Harshi-itaSinha/target-engine/internal/models"
)

// MemoryRepository implements Repository interface using in-memory storage
type MemoryRepository struct {
	campaigns      map[string]*model.Campaign
	targetingRules map[string][]*model.TargetingRule // keyed by campaign_id
	rulesByID      map[int64]*model.TargetingRule
	mutex          sync.RWMutex
	nextRuleID     int64
}

// NewMemoryRepository creates a new in-memory repository
func NewMemoryRepository() *MemoryRepository {
	repo := &MemoryRepository{
		campaigns:      make(map[string]*model.Campaign),
		targetingRules: make(map[string][]*model.TargetingRule),
		rulesByID:      make(map[int64]*model.TargetingRule),
		nextRuleID:     1,
	}

	// Initialize with sample data
	repo.initializeSampleData()

	return repo
}

// Campaign returns the campaign repository
func (r *MemoryRepository) Campaign() CampaignRepository {
	return r
}

// TargetingRule returns the targeting rule repository
func (r *MemoryRepository) TargetingRule() TargetingRuleRepository {
	return r
}

// Close closes the repository connection
func (r *MemoryRepository) Close() error {
	return nil
}

// Health checks if the repository is healthy
func (r *MemoryRepository) Health(ctx context.Context) error {
	return nil
}

// Migrate runs database migrations (no-op for memory)
func (r *MemoryRepository) Migrate(ctx context.Context) error {
	return nil
}

// Campaign Repository Methods

// GetActiveCampaigns returns all active campaigns
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

// GetCampaignByID returns a campaign by its ID
func (r *MemoryRepository) GetCampaignByID(ctx context.Context, id string) (*model.Campaign, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	campaign, exists := r.campaigns[id]
	if !exists {
		return nil, fmt.Errorf("campaign with ID %s not found", id)
	}

	return campaign, nil
}

// CreateCampaign creates a new campaign
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

// UpdateCampaign updates an existing campaign
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

// DeleteCampaign deletes a campaign by ID
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

// GetTargetingRulesByCampaignID returns targeting rules for a specific campaign
func (r *MemoryRepository) GetTargetingRulesByCampaignID(ctx context.Context, campaignID string) ([]*model.TargetingRule, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	rules, exists := r.targetingRules[campaignID]
	if !exists {
		return []*model.TargetingRule{}, nil
	}

	return rules, nil
}

// CreateTargetingRule creates a new targeting rule
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

// UpdateTargetingRule updates an existing targeting rule
func (r *MemoryRepository) UpdateTargetingRule(ctx context.Context, rule *model.TargetingRule) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	existingRule, exists := r.rulesByID[rule.ID]
	if !exists {
		return fmt.Errorf("targeting rule with ID %d not found", rule.ID)
	}

	rule.UpdatedAt = time.Now()

	// Update in both maps
	r.rulesByID[rule.ID] = rule

	// Update in the campaign rules slice
	rules := r.targetingRules[existingRule.CampaignID]
	for i, r := range rules {
		if r.ID == rule.ID {
			rules[i] = rule
			break
		}
	}

	return nil
}

// DeleteTargetingRule deletes a targeting rule by ID
func (r *MemoryRepository) DeleteTargetingRule(ctx context.Context, id int64) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// rule, exists := r.rulesByID[id]
	// if !exists {
	// 	return fmt.Errorf("targeting rule with ID %d not found", id)
	// }

	delete(r.rulesByID, id)

	// Remove from campaign rules slice
	// rules := r.targetingRules[rule.CampaignID]
	// for i, r := range rules {
	// 	if r.ID == id {
	// 		r.targetingRules[rule.CampaignID] = append(rules[:i], rules[i+1:]...)
	// 		break
	// 	}
	// }

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

	// Remove from rulesByID map
	for _, rule := range rules {
		delete(r.rulesByID, rule.ID)
	}

	// Remove from targetingRules map
	delete(r.targetingRules, campaignID)

	return nil
}

// initializeSampleData loads the sample data from the problem statement
func (r *MemoryRepository) initializeSampleData() {
	now := time.Now()

	// Sample campaigns
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

	// Sample targeting rules
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

	// Initialize campaigns
	for _, campaign := range campaigns {
		r.campaigns[campaign.ID] = campaign
	}

	// Initialize targeting rules
	r.nextRuleID = 4 // Start from 4 since we have 3 sample rules
	for _, rule := range targetingRules {
		r.targetingRules[rule.CampaignID] = append(r.targetingRules[rule.CampaignID], rule)
		r.rulesByID[rule.ID] = rule
	}
}
