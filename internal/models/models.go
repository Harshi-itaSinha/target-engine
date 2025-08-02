package model

import "time"

// Campaign represents an advertising campaign
type Campaign struct {
	ID        string    `json:"cid" db:"id"`
	Name      string    `json:"name" db:"name"`
	Image     string    `json:"img" db:"image"`
	CTA       string    `json:"cta" db:"cta"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// TargetingRule represents targeting criteria for campaigns
type TargetingRule struct {
	ID             int64     `json:"id" db:"id"`
	CampaignID     string    `json:"campaign_id" db:"campaign_id"`
	IncludeCountry []string  `json:"include_country" db:"include_country"`
	ExcludeCountry []string  `json:"exclude_country" db:"exclude_country"`
	IncludeOS      []string  `json:"include_os" db:"include_os"`
	ExcludeOS      []string  `json:"exclude_os" db:"exclude_os"`
	IncludeApp     []string  `json:"include_app" db:"include_app"`
	ExcludeApp     []string  `json:"exclude_app" db:"exclude_app"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// DeliveryRequest represents the incoming request parameters
type DeliveryRequest struct {
	App     string `json:"app" validate:"required"`
	Country string `json:"country" validate:"required"`
	OS      string `json:"os" validate:"required"`
}

// DeliveryResponse represents the response for matching campaigns
type DeliveryResponse struct {
	CID   string `json:"cid"`
	Image string `json:"img"`
	CTA   string `json:"cta"`
}

// ErrorResponse represents error response structure
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

// CampaignStatus constants
const (
	StatusActive   = "ACTIVE"
	StatusInactive = "INACTIVE"
)

// IsActive checks if the campaign is active
func (c *Campaign) IsActive() bool {
	return c.Status == StatusActive
}

// ToDeliveryResponse converts Campaign to DeliveryResponse
func (c *Campaign) ToDeliveryResponse() *DeliveryResponse {
	return &DeliveryResponse{
		CID:   c.ID,
		Image: c.Image,
		CTA:   c.CTA,
	}
}