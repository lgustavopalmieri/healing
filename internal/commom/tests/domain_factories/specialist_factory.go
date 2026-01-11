package domainfactories

import "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"

func CreateSpecialistFactory(overrides ...func(*domain.Specialist)) (*domain.Specialist, error) {
	specialist, err := domain.CreateSpecialist(
		"Dr. João Silva",
		"joao@example.com",
		"+5511999999999",
		"Cardiologia",
		"CRM123456",
		"Especialista em cardiologia",
		[]string{"coração", "arritmia"},
		true,
	)
	if err != nil {
		return nil, err
	}

	for _, override := range overrides {
		override(specialist)
	}

	return specialist, nil
}
