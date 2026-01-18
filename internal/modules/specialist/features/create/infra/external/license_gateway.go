package external

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/application"
)

// LicenseValidationResponse represents the response from the external license validation API
type LicenseValidationResponse struct {
	LicenseNumber string `json:"licenseNumber"`
	Status        string `json:"status"` // "valid", "invalid", or "idle"
}

type LicenseValidationGateway struct {
	httpClient *http.Client
	baseURL    string
}

func NewLicenseValidationGateway() application.SpecialistCreateExternalGatewayInterface {
	return &LicenseValidationGateway{
		httpClient: &http.Client{
			Timeout: 1 * time.Second, // 2s timeout for HTTP calls
		},
		baseURL: "http://license-api:7500", // License validation service (local for testing)
	}
}

func (g *LicenseValidationGateway) ValidateLicenseNumber(ctx context.Context, licenseNumber string) (bool, error) {
	if licenseNumber == "" {
		return false, fmt.Errorf("license number cannot be empty")
	}

	// Build the request URL with license as parameter (keeping original parameter name)
	requestURL := fmt.Sprintf("%s/validate-license?license=%s", g.baseURL, url.QueryEscape(licenseNumber))

	// Create HTTP request with context for cancellation
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "healing-specialist/1.0")

	// Make the HTTP call
	resp, err := g.httpClient.Do(req)
	if err != nil {
		// Check if it's a context cancellation
		if ctx.Err() != nil {
			return false, ctx.Err()
		}
		return false, fmt.Errorf("failed to call license validation API: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("license validation API returned status %d", resp.StatusCode)
	}

	// Parse the JSON response
	var licenseResp LicenseValidationResponse
	if err := json.NewDecoder(resp.Body).Decode(&licenseResp); err != nil {
		return false, fmt.Errorf("failed to parse license validation response: %w", err)
	}

	// DEBUG: Log para verificar o que está sendo recebido
	fmt.Printf("🔍 DEBUG Gateway - License: %s, Status: %s\n", licenseResp.LicenseNumber, licenseResp.Status)

	// Return based on status
	switch licenseResp.Status {
	case "valid":
		fmt.Printf("✅ DEBUG Gateway - Returning TRUE (valid license)\n")
		return true, nil
	case "invalid", "idle":
		fmt.Printf("❌ DEBUG Gateway - Returning FALSE (invalid/idle license) - THIS IS NOT AN ERROR!\n")
		return false, nil
	default:
		fmt.Printf("❓ DEBUG Gateway - Unknown status: %s\n", licenseResp.Status)
		return false, fmt.Errorf("unknown license status: %s", licenseResp.Status)
	}
}
