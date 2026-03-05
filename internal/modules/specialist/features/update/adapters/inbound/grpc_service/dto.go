package grpcservice

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/application"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/adapters/inbound/grpc_service/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToUpdateSpecialistDTO(req *pb.UpdateSpecialistRequest) application.UpdateSpecialistDTO {
	dto := application.UpdateSpecialistDTO{
		ID: req.GetId(),
	}

	if req.GetName() != nil {
		v := req.GetName().GetValue()
		dto.Name = &v
	}

	if req.GetEmail() != nil {
		v := req.GetEmail().GetValue()
		dto.Email = &v
	}

	if req.GetPhone() != nil {
		v := req.GetPhone().GetValue()
		dto.Phone = &v
	}

	if req.GetSpecialty() != nil {
		v := req.GetSpecialty().GetValue()
		dto.Specialty = &v
	}

	if req.GetLicenseNumber() != nil {
		v := req.GetLicenseNumber().GetValue()
		dto.LicenseNumber = &v
	}

	if req.GetDescription() != nil {
		v := req.GetDescription().GetValue()
		dto.Description = &v
	}

	if len(req.GetKeywords()) > 0 {
		dto.Keywords = append([]string(nil), req.GetKeywords()...)
	}

	if req.GetAgreedToShare() != nil {
		v := req.GetAgreedToShare().GetValue()
		dto.AgreedToShare = &v
	}

	if req.GetStatus() != nil {
		v := domain.SpecialistStatus(req.GetStatus().GetValue())
		dto.Status = &v
	}

	return dto
}

func ToUpdateSpecialistResponse(specialist *domain.Specialist) *pb.UpdateSpecialistResponse {
	return &pb.UpdateSpecialistResponse{
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
		Rating:        s.Rating,
		Status:        string(s.Status),
		CreatedAt:     timestamppb.New(s.CreatedAt),
		UpdatedAt:     timestamppb.New(s.UpdatedAt),
	}
}
