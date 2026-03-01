package listener

type ValidateLicenseEventPayload struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	LicenseNumber string `json:"licenseNumber"`
	Specialty     string `json:"specialty"`
}
