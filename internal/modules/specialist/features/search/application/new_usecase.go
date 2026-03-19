package application

type SearchSpecialistsUseCase struct {
	repository SpecialistSearchRepositoryInterface
}

func NewSearchSpecialistsUseCase(
	repository SpecialistSearchRepositoryInterface,
) *SearchSpecialistsUseCase {
	return &SearchSpecialistsUseCase{
		repository: repository,
	}
}
