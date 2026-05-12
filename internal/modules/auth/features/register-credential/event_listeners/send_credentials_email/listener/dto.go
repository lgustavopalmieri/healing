package listener

type CredentialPendingPayload struct {
	SubjectID        string `json:"subject_id"`
	Role             string `json:"role"`
	Email            string `json:"email"`
	SetPasswordToken string `json:"set_password_token"`
}
