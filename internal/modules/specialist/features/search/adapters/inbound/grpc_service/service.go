package grpcservice

import (
	"context"
	"errors"

	searchoutput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_output"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/application"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/search/adapters/inbound/grpc_service/pb"
)

var ErrNilRequest = errors.New("request cannot be nil")

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
	if req == nil {
		return nil, ErrNilRequest
	}

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
