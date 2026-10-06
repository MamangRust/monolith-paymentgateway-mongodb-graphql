package handler

import (
	"context"

	pbuserrole "github.com/MamangRust/monolith-payment-gateway-pb/user_role"
	"github.com/MamangRust/monolith-payment-gateway-role/service"
	"github.com/MamangRust/monolith-payment-gateway-shared/errors"
	role_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/role_errors/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type userRoleHandleGrpc struct {
	pbuserrole.UnimplementedUserRoleServiceServer
	userRole service.UserRoleService
}

func NewUserRoleHandleGrpc(userRole service.UserRoleService) UserRoleHandlerGrpc {
	return &userRoleHandleGrpc{
		userRole: userRole,
	}
}

func (s *userRoleHandleGrpc) AssignRoleToUser(ctx context.Context, reqPb *pbuserrole.AssignRoleToUserRequest) (*pbuserrole.ApiResponseUserRole, error) {
	userID := int(reqPb.GetUserId())
	roleID := int(reqPb.GetRoleId())

	if userID <= 0 || roleID <= 0 {
		return nil, role_errors.ErrGrpcRoleInvalidId
	}

	userRole, err := s.userRole.AssignRoleToUser(ctx, userID, roleID)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pbuserrole.ApiResponseUserRole{
		Status:  "success",
		Message: "Successfully assigned role to user",
		Data: &pbuserrole.UserRoleResponse{
			UserRoleId: userRole.UserRoleID,
			UserId:     userRole.UserID,
			RoleId:     userRole.RoleID,
		},
	}, nil
}

func (s *userRoleHandleGrpc) RemoveRoleFromUser(ctx context.Context, reqPb *pbuserrole.RemoveRoleFromUserRequest) (*emptypb.Empty, error) {
	userID := int(reqPb.GetUserId())
	roleID := int(reqPb.GetRoleId())

	if userID <= 0 || roleID <= 0 {
		return nil, role_errors.ErrGrpcRoleInvalidId
	}

	if err := s.userRole.RemoveRoleFromUser(ctx, userID, roleID); err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &emptypb.Empty{}, nil
}
