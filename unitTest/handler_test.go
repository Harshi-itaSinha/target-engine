package unitTest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Harshi-itaSinha/target-engine/internal/config"
	"github.com/Harshi-itaSinha/target-engine/internal/handler"
	model "github.com/Harshi-itaSinha/target-engine/internal/models"
	"github.com/Harshi-itaSinha/target-engine/internal/repository"
	"github.com/Harshi-itaSinha/target-engine/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeliveryHandler_GetCampaigns(t *testing.T) {
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
	deliveryHandler := handler.NewDeliveryHandler(targetingService)

	// Allow some time for cache initialization
	time.Sleep(100 * time.Millisecond)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
		expectedCampaigns []string
	}{
		{
			name:           "Valid request - Germany Android",
			queryParams:    "app=com.abc.xyz&country=germany&os=android",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			expectedCampaigns: []string{"duolingo"},
		},
		{
			name:           "Valid request - US Android LudoKing",
			queryParams:    "app=com.gametion.ludokinggame&country=us&os=android",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			expectedCampaigns: []string{"spotify", "subwaysurfer"},
		},
		{
			name:           "Valid request - Canada iOS",
			queryParams:    "app=com.example.app&country=canada&os=ios",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			expectedCampaigns: []string{"spotify"},
		},
		{
			name:           "Missing app parameter",
			queryParams:    "country=germany&os=android",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
		},
		{
			name:           "Missing country parameter",
			queryParams:    "app=com.abc.xyz&os=android",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
		},
		{
			name:           "Missing OS parameter",
			queryParams:    "app=com.abc.xyz&country=germany",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
		},
		{
			name:           "No matching campaigns",
			queryParams:    "app=com.nonexistent.app&country=antarctica&os=windows",
			expectedStatus: http.StatusNoContent,
			expectedCount:  0,
		},
		{
			name:           "Case insensitive matching",
			queryParams:    "app=com.abc.xyz&country=GERMANY&os=ANDROID",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			expectedCampaigns: []string{"duolingo"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req, err := http.NewRequest("GET", "/v1/delivery?"+tt.queryParams, nil)
			require.NoError(t, err)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			deliveryHandler.GetCampaigns(rr, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				// Parse response
				var campaigns []*model.DeliveryResponse
				err := json.Unmarshal(rr.Body.Bytes(), &campaigns)
				require.NoError(t, err)

				// Check count
				assert.Len(t, campaigns, tt.expectedCount)

				// Check campaign IDs if specified
				if len(tt.expectedCampaigns) > 0 {
					var campaignIDs []string
					for _, campaign := range campaigns {
						campaignIDs = append(campaignIDs, campaign.CID)
					}
					assert.ElementsMatch(t, tt.expectedCampaigns, campaignIDs)
				}

				// Verify response structure
				for _, campaign := range campaigns {
					assert.NotEmpty(t, campaign.CID)
					assert.NotEmpty(t, campaign.Image)
					assert.NotEmpty(t, campaign.CTA)
				}
			} else if tt.expectedStatus == http.StatusBadRequest {
				// Parse error response
				var errorResp model.ErrorResponse
				err := json.Unmarshal(rr.Body.Bytes(), &errorResp)
				require.NoError(t, err)

				assert.Equal(t, "Bad Request", errorResp.Error)
				assert.NotEmpty(t, errorResp.Message)
			} else if tt.expectedStatus == http.StatusNoContent {
				// No content response should have empty body
				assert.Empty(t, rr.Body.String())
			}
		})
	}
}

func TestDeliveryHandler_GetStats(t *testing.T) {
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
	deliveryHandler := handler.NewDeliveryHandler(targetingService)

	// Allow some time for cache initialization
	time.Sleep(100 * time.Millisecond)

	// Create request
	req, err := http.NewRequest("GET", "/v1/stats", nil)
	require.NoError(t, err)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	deliveryHandler.GetStats(rr, req)

	// Check status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Parse response
	var stats map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &stats)
	require.NoError(t, err)

	// Verify stats structure
	assert.Contains(t, stats, "campaigns_count")
	assert.Contains(t, stats, "targeting_rules_count")
	assert.Contains(t, stats, "query_cache_size")
	assert.Contains(t, stats, "last_refresh")
	assert.Contains(t, stats, "cache_age_seconds")
}

func TestDeliveryHandler_Health(t *testing.T) {
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
	deliveryHandler := handler.NewDeliveryHandler(targetingService)

	// Create request
	req, err := http.NewRequest("GET", "/health", nil)
	require.NoError(t, err)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	deliveryHandler.Health(rr, req)

	// Check status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Parse response
	var health map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &health)
	require.NoError(t, err)

	// Verify health response structure
	assert.Equal(t, "ok", health["status"])
	assert.Equal(t, "targeting-engine", health["service"])
	assert.Contains(t, health, "version")
	assert.Contains(t, health, "timestamp")
}

func BenchmarkDeliveryHandler_GetCampaigns(b *testing.B) {
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
	deliveryHandler := handler.NewDeliveryHandler(targetingService)

	// Allow some time for cache initialization
	time.Sleep(100 * time.Millisecond)

	// Create request
	req, _ := http.NewRequest("GET", "/v1/delivery?app=com.abc.xyz&country=germany&os=android", nil)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rr := httptest.NewRecorder()
			deliveryHandler.GetCampaigns(rr, req)
		}
	})
}