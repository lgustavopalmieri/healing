package grpcservice

import (
	"context"

	searchoutput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_output"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/application"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/infra/grpc_service/pb"
)

type SpecialistSearchCommandInterface interface {
	Execute(ctx context.Context, dto *application.SearchSpecialistsDTO) (*searchoutput.ListSearchOutput, error)
}

type SpecialistSearchGRPCService struct {
	pb.UnimplementedSearchSpecialistServiceServer
	command SpecialistSearchCommandInterface
}

func NewSpecialistSearchGRPCService(command SpecialistSearchCommandInterface) *SpecialistSearchGRPCService {
	return &SpecialistSearchGRPCService{
		command: command,
	}
}

func (s *SpecialistSearchGRPCService) SearchSpecialists(ctx context.Context, req *pb.SearchSpecialistsRequest) (*pb.SearchSpecialistsResponse, error) {
	dto, err := ToSearchDTO(req)
	if err != nil {
		return nil, err
	}

	output, err := s.command.Execute(ctx, dto)
	if err != nil {
		return nil, err
	}

	return ToSearchSpecialistsResponse(output), nil
}
