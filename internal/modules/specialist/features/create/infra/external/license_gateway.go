package external

import (
	"context"
	"fmt"
	"time"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/application"
)

type LicenseValidationGateway struct {
}

func NewLicenseValidationGateway() application.SpecialistCreateExternalGatewayInterface {
	return &LicenseValidationGateway{}
}

func (g *LicenseValidationGateway) ValidateLicenseNumber(ctx context.Context, licenseNumber string) (bool, error) {
	if licenseNumber == "" {
		return false, fmt.Errorf("license number cannot be empty")
	}

	time.Sleep(500 * time.Millisecond)
	
	// TODO: Implement real external API call
	// For now, accept all non-empty license numbers
	return true, nil
}
