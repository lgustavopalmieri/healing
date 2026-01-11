package application

type CreateSpecialistDTO struct {
	Name          string   
	Email         string   
	Phone         string  
	Specialty     string  
	LicenseNumber string   
	Description   string   
	Keywords      []string
	AgreedToShare bool    
}