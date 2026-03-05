package grpcservice

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/application"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/adapters/inbound/grpc_service/pb"
)

type SpecialistUpdateCommandInterface interface {
	Execute(ctx context.Context, input application.UpdateSpecialistDTO) (*domain.Specialist, error)
}

type SpecialistUpdateGRPCService struct {
	pb.UnimplementedUpdateSpecialistServiceServer
	command SpecialistUpdateCommandInterface
}

func NewSpecialistUpdateGRPCService(command SpecialistUpdateCommandInterface) *SpecialistUpdateGRPCService {
	return &SpecialistUpdateGRPCService{
		command: command,
	}
}

func (s *SpecialistUpdateGRPCService) UpdateSpecialist(ctx context.Context, req *pb.UpdateSpecialistRequest) (*pb.UpdateSpecialistResponse, error) {
	dto := ToUpdateSpecialistDTO(req)
	specialist, err := s.command.Execute(ctx, dto)
	if err != nil {
		return nil, err
	}
	return ToUpdateSpecialistResponse(specialist), nil
}
