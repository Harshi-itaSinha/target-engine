package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/Harshi-itaSinha/target-engine/internal/config"
	"github.com/Harshi-itaSinha/target-engine/internal/models"
	"github.com/Harshi-itaSinha/target-engine/internal/repository"
)

// TargetingService handles the core business logic for campaign targeting
type TargetingService struct {
	repo        repository.Repository
	cache       *targetingCache
	config      *config.Config
	mutex       sync.RWMutex
	lastRefresh time.Time
}

// targetingCache represents an in-memory cache for targeting data
type targetingCache struct {
	campaigns      map[string]*model.Campaign
	targetingRules map[string][]*model.TargetingRule
	queryCache     map[string][]*model.DeliveryResponse
	mutex          sync.RWMutex
	lastUpdate     time.Time
}

// NewTargetingService creates a new targeting service
func NewTargetingService(repo repository.Repository, cfg *config.Config) *TargetingService {
	service := &TargetingService{
		repo:   repo,
		config: cfg,
		cache: &targetingCache{
			campaigns:      make(map[string]*model.Campaign),
			targetingRules: make(map[string][]*model.TargetingRule),
			queryCache:     make(map[string][]*model.DeliveryResponse),
		},
	}

	// Initialize cache
	go service.refreshCache()

	// Start periodic cache refresh
	go service.startCacheRefreshWorker()

	return service
}

// GetMatchingCampaigns returns campaigns that match the targeting criteria
func (s *TargetingService) GetMatchingCampaigns(ctx context.Context, req *model.DeliveryRequest) ([]*model.DeliveryResponse, error) {
	// Validate request
	if err := s.validateRequest(req); err != nil {
		return nil, err
	}

	// Normalize request parameters
	normalizedReq := s.normalizeRequest(req)

	// Check query cache first
	cacheKey := s.generateCacheKey(normalizedReq)
	if cached := s.getFromQueryCache(cacheKey); cached != nil {
		return cached, nil
	}

	// Get matching campaigns
	matches, err := s.findMatchingCampaigns(ctx, normalizedReq)
	if err != nil {
		return nil, fmt.Errorf("failed to find matching campaigns: %w", err)
	}

	// Cache the result
	s.setToQueryCache(cacheKey, matches)

	return matches, nil
}

// validateRequest validates the delivery request
func (s *TargetingService) validateRequest(req *model.DeliveryRequest) error {
	var validate = validator.New()
	return validate.Struct(req)
}

// normalizeRequest normalizes request parameters for consistent matching
func (s *TargetingService) normalizeRequest(req *model.DeliveryRequest) *model.DeliveryRequest {
	return &model.DeliveryRequest{
		App:     strings.TrimSpace(req.App),
		Country: strings.ToUpper(strings.TrimSpace(req.Country)),
		OS:      strings.TrimSpace(req.OS),
	}
}

// generateCacheKey generates a cache key for the request
func (s *TargetingService) generateCacheKey(req *model.DeliveryRequest) string {
	return fmt.Sprintf("%s|%s|%s", req.App, req.Country, strings.ToLower(req.OS))
}

// findMatchingCampaigns finds campaigns that match the targeting criteria
func (s *TargetingService) findMatchingCampaigns(ctx context.Context, req *model.DeliveryRequest) ([]*model.DeliveryResponse, error) {
	s.cache.mutex.RLock()
	defer s.cache.mutex.RUnlock()

	var matches []*model.DeliveryResponse

	for campaignID, campaign := range s.cache.campaigns {
		// Only consider active campaigns
		if !campaign.IsActive() {
			continue
		}

		// Check if campaign matches targeting rules
		if s.campaignMatches(campaignID, req) {
			matches = append(matches, campaign.ToDeliveryResponse())
		}
	}

	return matches, nil
}

// campaignMatches checks if a campaign matches the targeting criteria
func (s *TargetingService) campaignMatches(campaignID string, req *model.DeliveryRequest) bool {
	rules, exists := s.cache.targetingRules[campaignID]
	if !exists || len(rules) == 0 {
		// No targeting rules means the campaign matches all requests
		return true
	}

	// Check each targeting rule (OR logic between rules, AND logic within a rule)
	for _, rule := range rules {
		if s.ruleMatches(rule, req) {
			return true
		}
	}

	return false
}

// ruleMatches checks if a single targeting rule matches the request
func (s *TargetingService) ruleMatches(rule *model.TargetingRule, req *model.DeliveryRequest) bool {
	// Check country targeting
	if !s.matchesDimension(req.Country, rule.IncludeCountry, rule.ExcludeCountry, true) {
		return false
	}

	// Check OS targeting
	if !s.matchesDimension(req.OS, rule.IncludeOS, rule.ExcludeOS, false) {
		return false
	}

	// Check app targeting
	if !s.matchesDimension(req.App, rule.IncludeApp, rule.ExcludeApp, true) {
		return false
	}

	return true
}

// matchesDimension checks if a value matches the include/exclude lists for a dimension
func (s *TargetingService) matchesDimension(value string, include, exclude []string, caseSensitive bool) bool {
	// Check exclusions first
	if len(exclude) > 0 {
		if s.containsValue(exclude, value, caseSensitive) {
			return false
		}
	}

	// Check inclusions
	if len(include) > 0 {
		return s.containsValue(include, value, caseSensitive)
	}

	// No include/exclude rules for this dimension means it matches
	return true
}

// containsValue checks if a slice contains a value
func (s *TargetingService) containsValue(slice []string, value string, caseSensitive bool) bool {
	for _, item := range slice {
		if caseSensitive {
			if item == value {
				return true
			}
		} else {
			if strings.EqualFold(item, value) {
				return true
			}
		}
	}
	return false
}

// getFromQueryCache retrieves a cached query result
func (s *TargetingService) getFromQueryCache(key string) []*model.DeliveryResponse {
	s.cache.mutex.RLock()
	defer s.cache.mutex.RUnlock()

	// // Check if cache is still valid
	// if time.Since(s.cache.lastUpdate) > s.config.Cache.TTL {
	// 	return nil
	// }

	if result, exists := s.cache.queryCache[key]; exists {
		return result
	}
	return nil
}

// setToQueryCache stores a query result in cache
func (s *TargetingService) setToQueryCache(key string, result []*model.DeliveryResponse) {
	s.cache.mutex.Lock()
	defer s.cache.mutex.Unlock()

	//Implement simple LRU eviction if cache is full
	if len(s.cache.queryCache) >= s.config.Cache.MaxSize {
		// Remove oldest entries (simple approach - in production, use proper LRU)
		for k := range s.cache.queryCache {
			delete(s.cache.queryCache, k)
			break
		}
	}

	s.cache.queryCache[key] = result
}

// refreshCache refreshes the campaign and targeting rule cache from repository
func (s *TargetingService) refreshCache() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get active campaigns
	campaigns, err := s.repo.Campaign().GetActiveCampaigns(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active campaigns: %w", err)
	}

	// Get targeting rules
	targetingRules, err := s.repo.TargetingRule().GetTargetingRules(ctx)
	if err != nil {
		return fmt.Errorf("failed to get targeting rules: %w", err)
	}

	// Update cache
	s.cache.mutex.Lock()
	defer s.cache.mutex.Unlock()

	// Clear existing cache
	s.cache.campaigns = make(map[string]*model.Campaign)
	s.cache.targetingRules = make(map[string][]*model.TargetingRule)
	s.cache.queryCache = make(map[string][]*model.DeliveryResponse) // Clear query cache too

	// Populate campaigns
	for _, campaign := range campaigns {
		s.cache.campaigns[campaign.ID] = campaign
	}

	// Populate targeting rules grouped by campaign ID
	for _, rule := range targetingRules {
		s.cache.targetingRules[rule.CampaignID] = append(s.cache.targetingRules[rule.CampaignID], rule)
	}

	s.cache.lastUpdate = time.Now()
	s.lastRefresh = time.Now()

	return nil
}

// startCacheRefreshWorker starts a background worker to refresh cache periodically
func (s *TargetingService) startCacheRefreshWorker() {
	// // ticker := time.NewTicker(s.config.Cache.CleanupInterval)
	// defer ticker.Stop()

	// for range ticker.C {
	// 	if err := s.refreshCache(); err != nil {
	// 		// In production, use proper logging
	// 		fmt.Printf("Failed to refresh cache: %v\n", err)
	// 	}
	// }
}

// GetCacheStats returns cache statistics for monitoring
func (s *TargetingService) GetCacheStats() map[string]interface{} {
	s.cache.mutex.RLock()
	defer s.cache.mutex.RUnlock()

	return map[string]interface{}{
		"campaigns_count":       len(s.cache.campaigns),
		"targeting_rules_count": len(s.cache.targetingRules),
		"query_cache_size":      len(s.cache.queryCache),
		"last_refresh":          s.lastRefresh,
		"cache_age_seconds":     time.Since(s.cache.lastUpdate).Seconds(),
	}
}
