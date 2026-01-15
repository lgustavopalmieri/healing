package grpcservice

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/infra/grpc_service/pb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/application"
)

func ToCreateSpecialistInputDTO(
	req *pb.CreateSpecialistRequest,
) application.CreateSpecialistDTO {
	return application.CreateSpecialistDTO{
		Name:          req.GetName(),
		Email:         req.GetEmail(),
		Phone:         req.GetPhone(),
		Specialty:     req.GetSpecialty(),
		LicenseNumber: req.GetLicenseNumber(),
		Description:   req.GetDescription(),
		Keywords:      append([]string(nil), req.GetKeywords()...),
		AgreedToShare: req.GetAgreedToShare(),
	}
}

func ToCreateSpecialistResponse(
	specialist *domain.Specialist,
) *pb.CreateSpecialistResponse {
	return &pb.CreateSpecialistResponse{
		Specialist: ToProtoSpecialist(specialist),
	}
}

func ToProtoSpecialist(s *domain.Specialist) *pb.Specialist {
	if s == nil {
		return nil
	}

	return &pb.Specialist{
		Id:            s.ID,
		Name:          s.Name,
		Email:         s.Email,
		Phone:         s.Phone,
		Specialty:     s.Specialty,
		LicenseNumber: s.LicenseNumber,
		Description:   s.Description,
		Keywords:      append([]string(nil), s.Keywords...),
		AgreedToShare: s.AgreedToShare,
		CreatedAt:     timestamppb.New(s.CreatedAt),
		UpdatedAt:     timestamppb.New(s.UpdatedAt),
	}
}
