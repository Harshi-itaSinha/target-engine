package handler

import (
	"net/http"
    "github.com/Harshi-itaSinha/target-engine/internal/models"
	"github.com/Harshi-itaSinha/target-engine/internal/service"
	"github.com/Harshi-itaSinha/target-engine/pkg/response"
)

// DeliveryHandler handles delivery endpoint requests
type DeliveryHandler struct {
	targetingService *service.TargetingService
}

// NewDeliveryHandler creates a new delivery handler
func NewDeliveryHandler(targetingService *service.TargetingService) *DeliveryHandler {
	return &DeliveryHandler{
		targetingService: targetingService,
	}
}

// GetCampaigns handles GET /v1/delivery requests
func (h *DeliveryHandler) GetCampaigns(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	req := &model.DeliveryRequest{
		App:     query.Get("app"),
		Country: query.Get("country"),
		OS:      query.Get("os"),
	}

	// Get matching campaigns from service
	campaigns, err := h.targetingService.GetMatchingCampaigns(r.Context(), req)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	// Return appropriate response
	if len(campaigns) == 0 {
		response.NoContent(w)
		return
	}

	response.Success(w, campaigns)
}

// GetStats handles GET /v1/stats requests for monitoring
func (h *DeliveryHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := h.targetingService.GetCacheStats()
	response.Success(w, stats)
}

// Health handles GET /health requests
func (h *DeliveryHandler) Health(w http.ResponseWriter, r *http.Request) {
	healthStatus := map[string]interface{}{
		"status":    "ok",
		"service":   "targeting-engine",
		"version":   "1.0.0",
		"timestamp": "2025-01-31T00:00:00Z",
	}
	response.Success(w, healthStatus)
}