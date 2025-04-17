package dto

import (
	"github.com/BeksultanSE/Assignment1-user/internal/domain"
	proto "github.com/BeksultanSE/Assignment1-user/protos/gen/golang"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetRequestDTO maps to proto.GetUserProfileRequest
type GetRequestDTO struct {
	UserID uint64
}

func (dto *GetRequestDTO) ToDomainFilterUserID() domain.UserFilter {
	id := dto.UserID
	return domain.UserFilter{
		ID: &id,
	}
}

func (dto *GetRequestDTO) ValidateUserID() error {
	if dto.UserID == 0 {
		return status.Error(codes.InvalidArgument, "user_id is required")
	}
	return nil
}

// FromGetRequestProto converts proto.GetUserProfileRequest to DTO
func FromGetRequestProto(req *proto.UserID) GetRequestDTO {
	return GetRequestDTO{
		UserID: req.UserId,
	}
}

// GetResponseDTO maps to proto.GetUserProfileResponse
type GetResponseDTO struct {
	UserID uint64
	Email  string
	Name   string
}

func (dto *GetResponseDTO) ToProtoUserProfile() *proto.UserProfile {
	return &proto.UserProfile{
		UserId: dto.UserID,
		Email:  dto.Email,
		Name:   dto.Name,
	}
}

// FromUserProfileDomain converts domain.User to DTO
func FromUserProfileDomain(user domain.User) GetResponseDTO {
	return GetResponseDTO{
		UserID: user.ID,
		Email:  user.Email,
		Name:   user.Name,
	}
}
