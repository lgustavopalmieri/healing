package grpcservice

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/application"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/adapters/inbound/grpc_service/pb"
)

type SpecialistCreateUseCaseInterface interface {
	Execute(ctx context.Context, input application.CreateSpecialistDTO) (*domain.Specialist, error)
}

type SpecialistCreateGRPCService struct {
	pb.UnimplementedSpecialistServiceServer
	useCase SpecialistCreateUseCaseInterface
}

func NewSpecialistCreateGRPCService(useCase SpecialistCreateUseCaseInterface) *SpecialistCreateGRPCService {
	return &SpecialistCreateGRPCService{
		useCase: useCase,
	}
}

func (s *SpecialistCreateGRPCService) CreateSpecialist(ctx context.Context, input *pb.CreateSpecialistRequest) (*pb.CreateSpecialistResponse, error) {
	dto := ToCreateSpecialistInputDTO(input)
	specialist, err := s.useCase.Execute(ctx, dto)
	if err != nil {
		return nil, err
	}
	output := ToCreateSpecialistResponse(specialist)
	return output, nil
}
