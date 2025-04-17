package grpc

import (
	"context"
	"github.com/BeksultanSE/Assignment1-user/internal/adapter/grpc/dto"
	"github.com/BeksultanSE/Assignment1-user/internal/domain"
	"github.com/BeksultanSE/Assignment1-user/internal/usecase"
	proto "github.com/BeksultanSE/Assignment1-user/protos/gen/golang"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserGRPCServer struct {
	proto.UnimplementedAuthServer
	userUsecase usecase.UserUsecase
}

func NewUserGRPCServer(userUsecase usecase.UserUsecase) *UserGRPCServer {
	return &UserGRPCServer{userUsecase: userUsecase}
}

func (s *UserGRPCServer) RegisterUser(ctx context.Context, req *proto.UserRequest) (*proto.UserResponse, error) {
	requestDTO := dto.FromRegisterUserRequestProto(req)

	// Validate
	if err := requestDTO.ValidateUserRequest(); err != nil {
		return nil, err
	}

	user, err := s.userUsecase.Register(ctx, requestDTO.ToDomainUserRequest())
	if err != nil {
		switch err {
		case domain.ErrUserExists:
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	// Convert domain to DTO and protobuf
	responseDTO := dto.FromRegisterUserDomain(user)
	return responseDTO.ToProtoUserResponse(), nil
}

func (s *UserGRPCServer) AuthenticateUser(ctx context.Context, req *proto.AuthRequest) (*proto.AuthResponse, error) {
	// Convert protobuf to DTO
	requestDTO := dto.FromAuthenticateUserRequestProto(req)

	// Validate
	if err := requestDTO.ValidateAuthRequest(); err != nil {
		return nil, err
	}

	// Call usecase
	user, err := s.userUsecase.Authenticate(ctx, requestDTO.ToDomainAuthRequest())
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			return nil, status.Error(codes.NotFound, "user not found")
		case domain.ErrInvalidPassword:
			return nil, status.Error(codes.Unauthenticated, "invalid password")
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	// Convert domain to DTO and protobuf
	responseDTO := dto.FromAuthenticateUserDomain(user)
	return responseDTO.ToProtoAuthResponse(), nil
}

func (s *UserGRPCServer) GetUserProfile(ctx context.Context, req *proto.UserID) (*proto.UserProfile, error) {
	// Convert protobuf to DTO
	requestDTO := dto.FromGetRequestProto(req)

	// Validate
	if err := requestDTO.ValidateUserID(); err != nil {
		return nil, err
	}

	// Call usecase
	user, err := s.userUsecase.Get(ctx, requestDTO.ToDomainFilterUserID())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if user.ID == 0 {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	// Convert domain to DTO and protobuf
	responseDTO := dto.FromUserProfileDomain(user)
	return responseDTO.ToProtoUserProfile(), nil
}
