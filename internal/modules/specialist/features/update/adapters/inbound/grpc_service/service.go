package grpcservice

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/application"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/adapters/inbound/grpc_service/pb"
)

type SpecialistUpdateUseCaseInterface interface {
	Execute(ctx context.Context, input application.UpdateSpecialistDTO) (*domain.Specialist, error)
}

type SpecialistUpdateGRPCService struct {
	pb.UnimplementedUpdateSpecialistServiceServer
	useCase SpecialistUpdateUseCaseInterface
}

func NewSpecialistUpdateGRPCService(useCase SpecialistUpdateUseCaseInterface) *SpecialistUpdateGRPCService {
	return &SpecialistUpdateGRPCService{
		useCase: useCase,
	}
}

func (s *SpecialistUpdateGRPCService) UpdateSpecialist(ctx context.Context, req *pb.UpdateSpecialistRequest) (*pb.UpdateSpecialistResponse, error) {
	dto := ToUpdateSpecialistDTO(req)
	specialist, err := s.useCase.Execute(ctx, dto)
	if err != nil {
		return nil, err
	}
	return ToUpdateSpecialistResponse(specialist), nil
}
