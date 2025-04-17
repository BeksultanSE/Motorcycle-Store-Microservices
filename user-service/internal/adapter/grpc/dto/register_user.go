package dto

import (
	"github.com/BeksultanSE/Assignment1-user/internal/domain"
	proto "github.com/BeksultanSE/Assignment1-user/protos/gen/golang"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RegisterUserRequestDTO maps to proto.RegisterUserRequest
type RegisterUserRequestDTO struct {
	Email    string
	Password string
	Name     string
}

// Validate ensures required fields are present
func (dto *RegisterUserRequestDTO) ValidateUserRequest() error {
	if dto.Email == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}
	if dto.Password == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}
	if dto.Name == "" {
		return status.Error(codes.InvalidArgument, "name is required")
	}
	return nil
}

// ToDomain converts DTO to domain.User
func (dto *RegisterUserRequestDTO) ToDomainUserRequest() domain.User {
	return domain.User{
		Email:          dto.Email,
		HashedPassword: dto.Password, // Plain password, hashed by usecase
		Name:           dto.Name,
	}
}

// FromRegisterUserRequestProto converts proto.RegisterUserRequest to DTO
func FromRegisterUserRequestProto(req *proto.UserRequest) RegisterUserRequestDTO {
	return RegisterUserRequestDTO{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	}
}

// RegisterUserResponseDTO maps to proto.RegisterUserResponse
type RegisterUserResponseDTO struct {
	UserID uint64
	Name   string
}

// ToProto converts DTO to proto.RegisterUserResponse
func (dto *RegisterUserResponseDTO) ToProtoUserResponse() *proto.UserResponse {
	return &proto.UserResponse{
		UserId: dto.UserID,
	}
}

// FromRegisterUserDomain converts domain.User to DTO
func FromRegisterUserDomain(user domain.User) RegisterUserResponseDTO {
	return RegisterUserResponseDTO{
		UserID: user.ID,
		Name:   user.Name,
	}
}
