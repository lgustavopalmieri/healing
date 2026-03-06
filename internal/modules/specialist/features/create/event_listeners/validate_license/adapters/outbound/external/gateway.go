package external

import (
	"context"
	"net/http"
)

type LicenseGateway struct {
	BaseURL string
	Client  *http.Client
}

type LicenseValidationResponse struct {
	Valid bool `json:"valid"`
}

func (g *LicenseGateway) Validate(ctx context.Context, licenseNumber string) (bool, error) {
	// url := fmt.Sprintf("%s/api/v1/licenses/%s/validate", g.BaseURL, licenseNumber)

	// req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	// if err != nil {
	// 	return false, fmt.Errorf("failed to create license validation request: %w", err)
	// }

	// resp, err := g.Client.Do(req)
	// if err != nil {
	// 	return false, fmt.Errorf("failed to call license validation service: %w", err)
	// }
	// defer resp.Body.Close()

	// if resp.StatusCode != http.StatusOK {
	// 	return false, fmt.Errorf("license validation service returned status: %d", resp.StatusCode)
	// }

	// var result LicenseValidationResponse
	// if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
	// 	return false, fmt.Errorf("failed to decode license validation response: %w", err)
	// }

	// return result.Valid, nil
	return true, nil
}
