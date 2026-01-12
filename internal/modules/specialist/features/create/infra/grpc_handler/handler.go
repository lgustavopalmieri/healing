package grpchandler

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/application"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/infra/grpc_handler/pb"
)

type SpecialistCreateCommandInterface interface {
	Execute(ctx context.Context, input application.CreateSpecialistDTO) (*domain.Specialist, error)
}

type SpecialistCreateGRPCHandler struct {
	pb.UnimplementedSpecialistServiceServer
	command SpecialistCreateCommandInterface
}

func NewSpecialistCreateGRPCHandler(command SpecialistCreateCommandInterface) *SpecialistCreateGRPCHandler {
	return &SpecialistCreateGRPCHandler{
		command: command,
	}
}

func (s *SpecialistCreateGRPCHandler) Handle(ctx context.Context, input *pb.CreateSpecialistRequest) (*pb.CreateSpecialistResponse, error) {
	dto := ToCreateSpecialistInputDTO(input)
	specialist, err := s.command.Execute(ctx, dto)
	if err != nil {
		return nil, err
	}
	output := ToCreateSpecialistResponse(specialist)
	return output, nil
}
