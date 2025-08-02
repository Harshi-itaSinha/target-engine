package unitTest

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Harshi-itaSinha/target-engine/internal/config"
	"github.com/Harshi-itaSinha/target-engine/internal/models"
	"github.com/Harshi-itaSinha/target-engine/internal/repository"
	"github.com/Harshi-itaSinha/target-engine/internal/service"
)

func TestTargetingService_GetMatchingCampaigns(t *testing.T) {
	// Setup
	repo := repository.NewMemoryRepository()
	cfg := &config.Config{
		// Cache: config.CacheConfig{
		// 	TTL:             5 * time.Minute,
		// 	CleanupInterval: 10 * time.Minute,
		// 	MaxSize:         100,
		// },
	}
	
	targetingService := service.NewTargetingService(repo, cfg)
	ctx := context.Background()

	tests := []struct {
		name           string
		request        *model.DeliveryRequest
		expectedCampaigns []string
		expectError    bool
	}{
		{
			name: "Germany Android request should match Duolingo",
			request: &model.DeliveryRequest{
				App:     "com.abc.xyz",
				Country: "germany",
				OS:      "android",
			},
			expectedCampaigns: []string{"duolingo"},
			expectError:       false,
		},
		{
			name: "US Android request for LudoKing should match Spotify and SubwaySurfer",
			request: &model.DeliveryRequest{
				App:     "com.gametion.ludokinggame",
				Country: "us",
				OS:      "android",
			},
			expectedCampaigns: []string{"spotify", "subwaysurfer"},
			expectError:       false,
		},
		{
			name: "Canada iOS request should match Spotify only",
			request: &model.DeliveryRequest{
				App:     "com.example.app",
				Country: "canada",
				OS:      "ios",
			},
			expectedCampaigns: []string{"spotify"},
			expectError:       false,
		},
		{
			name: "Missing app parameter should return error",
			request: &model.DeliveryRequest{
				Country: "us",
				OS:      "android",
			},
			expectedCampaigns: nil,
			expectError:       true,
		},
		{
			name: "Missing country parameter should return error",
			request: &model.DeliveryRequest{
				App: "com.example.app",
				OS:  "android",
			},
			expectedCampaigns: nil,
			expectError:       true,
		},
		{
			name: "Missing OS parameter should return error",
			request: &model.DeliveryRequest{
				App:     "com.example.app",
				Country: "us",
			},
			expectedCampaigns: nil,
			expectError:       true,
		},
		{
			name: "No matching campaigns should return empty result",
			request: &model.DeliveryRequest{
				App:     "com.nonexistent.app",
				Country: "antarctica",
				OS:      "windows",
			},
			expectedCampaigns: []string{},
			expectError:       false,
		},
		{
			name: "Case insensitive OS matching",
			request: &model.DeliveryRequest{
				App:     "com.abc.xyz",
				Country: "germany",
				OS:      "ANDROID",
			},
			expectedCampaigns: []string{"duolingo"},
			expectError:       false,
		},
		{
			name: "Case insensitive country matching",
			request: &model.DeliveryRequest{
				App:     "com.example.app",
				Country: "CANADA",
				OS:      "ios",
			},
			expectedCampaigns: []string{"spotify"},
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Allow some time for cache initialization
			time.Sleep(100 * time.Millisecond)
			
			campaigns, err := targetingService.GetMatchingCampaigns(ctx, tt.request)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Convert campaigns to campaign IDs for comparison
			var campaignIDs []string
			for _, campaign := range campaigns {
				campaignIDs = append(campaignIDs, campaign.CID)
			}

			assert.ElementsMatch(t, tt.expectedCampaigns, campaignIDs)
		})
	}
}

func TestTargetingService_CacheStats(t *testing.T) {
	// Setup
	repo := repository.NewMemoryRepository()
	cfg := &config.Config{
		// Cache: config.CacheConfig{
		// 	TTL:             5 * time.Minute,
		// 	CleanupInterval: 10 * time.Minute,
		// 	MaxSize:         100,
		// },
	}
	
	targetingService := service.NewTargetingService(repo, cfg)

	// Allow some time for cache initialization
	time.Sleep(100 * time.Millisecond)

	// Get cache stats
	stats := targetingService.GetCacheStats()

	// Verify stats structure
	assert.Contains(t, stats, "campaigns_count")
	assert.Contains(t, stats, "targeting_rules_count")
	assert.Contains(t, stats, "query_cache_size")
	assert.Contains(t, stats, "last_refresh")
	assert.Contains(t, stats, "cache_age_seconds")

	// Verify expected values based on sample data
	assert.Equal(t, 3, stats["campaigns_count"])
	assert.Equal(t, 3, stats["targeting_rules_count"])
	assert.GreaterOrEqual(t, stats["query_cache_size"], 0)
}

func TestTargetingService_RequestNormalization(t *testing.T) {
	// Setup
	repo := repository.NewMemoryRepository()
	cfg := &config.Config{
		// Cache: config.CacheConfig{
		// 	TTL:             5 * time.Minute,
		// 	CleanupInterval: 10 * time.Minute,
		// 	MaxSize:         100,
		// },
	}
	
	targetingService := service.NewTargetingService(repo, cfg)
	ctx := context.Background()

	// Allow some time for cache initialization
	time.Sleep(100 * time.Millisecond)

	// Test request with whitespace and mixed case
	request := &model.DeliveryRequest{
		App:     " com.gametion.ludokinggame ",
		Country: " us ",
		OS:      " Android ",
	}

	campaigns, err := targetingService.GetMatchingCampaigns(ctx, request)
	require.NoError(t, err)

	// Should match spotify and subwaysurfer
	var campaignIDs []string
	for _, campaign := range campaigns {
		campaignIDs = append(campaignIDs, campaign.CID)
	}

	assert.ElementsMatch(t, []string{"spotify", "subwaysurfer"}, campaignIDs)
}

func TestTargetingService_CacheBehavior(t *testing.T) {
	// Setup
	repo := repository.NewMemoryRepository()
	cfg := &config.Config{
		// Cache: config.CacheConfig{
		// 	TTL:             100 * time.Millisecond, // Very short TTL for testing
		// 	CleanupInterval: 10 * time.Minute,
		// 	MaxSize:         100,
		// },
	}
	
	targetingService := service.NewTargetingService(repo, cfg)
	ctx := context.Background()

	// Allow some time for cache initialization
	time.Sleep(200 * time.Millisecond)

	request := &model.DeliveryRequest{
		App:     "com.abc.xyz",
		Country: "germany",
		OS:      "android",
	}

	// First request - should populate cache
	campaigns1, err := targetingService.GetMatchingCampaigns(ctx, request)
	require.NoError(t, err)
	assert.Len(t, campaigns1, 1)

	// Second request immediately - should use cache
	campaigns2, err := targetingService.GetMatchingCampaigns(ctx, request)
	require.NoError(t, err)
	assert.Equal(t, campaigns1, campaigns2)

	// Wait for cache to expire
	time.Sleep(200 * time.Millisecond)

	// Third request - should refresh cache
	campaigns3, err := targetingService.GetMatchingCampaigns(ctx, request)
	require.NoError(t, err)
	assert.Len(t, campaigns3, 1)
}