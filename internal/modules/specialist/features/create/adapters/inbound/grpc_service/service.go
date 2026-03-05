package grpcservice

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/application"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/adapters/inbound/grpc_service/pb"
)

type SpecialistCreateCommandInterface interface {
	Execute(ctx context.Context, input application.CreateSpecialistDTO) (*domain.Specialist, error)
}

type SpecialistCreateGRPCService struct {
	pb.UnimplementedSpecialistServiceServer
	command SpecialistCreateCommandInterface
}

func NewSpecialistCreateGRPCService(command SpecialistCreateCommandInterface) *SpecialistCreateGRPCService {
	return &SpecialistCreateGRPCService{
		command: command,
	}
}

func (s *SpecialistCreateGRPCService) CreateSpecialist(ctx context.Context, input *pb.CreateSpecialistRequest) (*pb.CreateSpecialistResponse, error) {
	dto := ToCreateSpecialistInputDTO(input)
	specialist, err := s.command.Execute(ctx, dto)
	if err != nil {
		return nil, err
	}
	output := ToCreateSpecialistResponse(specialist)
	return output, nil
}
