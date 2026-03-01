package external

import (
	"net/http"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/event_listeners/validate_license/application"
)

func NewLicenseGateway(baseURL string, client *http.Client) application.LicenseGatewayInterface {
	if client == nil {
		client = http.DefaultClient
	}
	return &LicenseGateway{
		BaseURL: baseURL,
		Client:  client,
	}
}
