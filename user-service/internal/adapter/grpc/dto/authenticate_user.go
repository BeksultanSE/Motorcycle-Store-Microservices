package dto

import (
	"github.com/BeksultanSE/Assignment1-user/internal/domain"
	proto "github.com/BeksultanSE/Assignment1-user/protos/gen/golang"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthenticateUserRequestDTO maps to proto.AuthenticateUserRequest
type AuthenticateUserRequestDTO struct {
	Email    string
	Password string
}

// Validate ensures required fields are present
func (dto *AuthenticateUserRequestDTO) ValidateAuthRequest() error {
	if dto.Email == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}
	if dto.Password == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}
	return nil
}

// ToDomain converts DTO to domain.User
func (dto *AuthenticateUserRequestDTO) ToDomainAuthRequest() domain.User {
	return domain.User{
		Email:          dto.Email,
		HashedPassword: dto.Password, // Plain password, verified by usecase
	}
}

// FromAuthenticateUserRequestProto converts proto.AuthenticateUserRequest to DTO
func FromAuthenticateUserRequestProto(req *proto.AuthRequest) AuthenticateUserRequestDTO {
	return AuthenticateUserRequestDTO{
		Email:    req.Email,
		Password: req.Password,
	}
}

// AuthenticateUserResponseDTO maps to proto.AuthenticateUserResponse
type AuthenticateUserResponseDTO struct {
	UserID        uint64
	Name          string
	Authenticated bool
}

// ToProto converts DTO to proto.AuthenticateUserResponse
func (dto *AuthenticateUserResponseDTO) ToProtoAuthResponse() *proto.AuthResponse {
	return &proto.AuthResponse{
		UserId:        dto.UserID,
		Authenticated: dto.Authenticated,
	}
}

// FromDomain converts domain.User to DTO
func FromAuthenticateUserDomain(user domain.User) AuthenticateUserResponseDTO {
	return AuthenticateUserResponseDTO{
		UserID:        user.ID,
		Name:          user.Name,
		Authenticated: true, // Only called on successful authentication
	}
}
