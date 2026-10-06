package handler

import (
	"context"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/role"
	"github.com/MamangRust/monolith-payment-gateway-role/service"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/errors"
	role_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/role_errors/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"time"
)

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

type roleCommandHandleGrpc struct {
	pb.UnimplementedRoleCommandServiceServer
	roleCommand service.RoleCommandService
}

func NewRoleCommandHandleGrpc(roleCommand service.RoleCommandService) RoleCommandHandlerGrpc {
	return &roleCommandHandleGrpc{
		roleCommand: roleCommand,
	}
}

func (s *roleCommandHandleGrpc) CreateRole(ctx context.Context, reqPb *pb.CreateRoleRequest) (*pb.ApiResponseRole, error) {
	req := &requests.CreateRoleRequest{
		Name: reqPb.Name,
	}

	if err := req.Validate(); err != nil {
		return nil, role_errors.ErrGrpcValidateCreateRole
	}

	role, err := s.roleCommand.CreateRole(ctx, req)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoRole := &pb.RoleResponse{
		Id:        int32(role.RoleID),
		Name:      role.RoleName,
		CreatedAt: formatTime(role.CreatedAt),
		UpdatedAt: formatTime(role.UpdatedAt),
	}

	return &pb.ApiResponseRole{
		Status:  "success",
		Message: "Successfully created role",
		Data:    protoRole,
	}, nil
}

func (s *roleCommandHandleGrpc) UpdateRole(ctx context.Context, reqPb *pb.UpdateRoleRequest) (*pb.ApiResponseRole, error) {
	roleID := int(reqPb.GetId())

	if roleID <= 0 {
		return nil, role_errors.ErrGrpcRoleInvalidId
	}

	name := reqPb.GetName()

	req := &requests.UpdateRoleRequest{
		ID:   &roleID,
		Name: name,
	}

	if err := req.Validate(); err != nil {
		return nil, role_errors.ErrGrpcValidateUpdateRole
	}

	role, err := s.roleCommand.UpdateRole(ctx, req)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoRole := &pb.RoleResponse{
		Id:        int32(role.RoleID),
		Name:      role.RoleName,
		CreatedAt: formatTime(role.CreatedAt),
		UpdatedAt: formatTime(role.UpdatedAt),
	}

	return &pb.ApiResponseRole{
		Status:  "success",
		Message: "Successfully updated role",
		Data:    protoRole,
	}, nil
}

func (s *roleCommandHandleGrpc) TrashedRole(ctx context.Context, req *pb.FindByIdRoleRequest) (*pb.ApiResponseRoleDeleteAt, error) {
	roleID := int(req.GetRoleId())

	if roleID <= 0 {
		return nil, role_errors.ErrGrpcRoleInvalidId
	}

	role, err := s.roleCommand.TrashedRole(ctx, roleID)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	deletedAt := ""
	if role.DeletedAt != nil {
		deletedAt = role.DeletedAt.Format("2006-01-02")
	}

	protoRole := &pb.RoleResponseDeleteAt{
		Id:        int32(role.RoleID),
		Name:      role.RoleName,
		CreatedAt: formatTime(role.CreatedAt),
		UpdatedAt: formatTime(role.UpdatedAt),
		DeletedAt: wrapperspb.String(deletedAt),
	}

	return &pb.ApiResponseRoleDeleteAt{
		Status:  "success",
		Message: "Successfully trashed role",
		Data:    protoRole,
	}, nil
}

func (s *roleCommandHandleGrpc) RestoreRole(ctx context.Context, req *pb.FindByIdRoleRequest) (*pb.ApiResponseRoleDeleteAt, error) {
	roleID := int(req.GetRoleId())

	if roleID <= 0 {
		return nil, role_errors.ErrGrpcRoleInvalidId
	}

	role, err := s.roleCommand.RestoreRole(ctx, roleID)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoRole := &pb.RoleResponseDeleteAt{
		Id:        int32(role.RoleID),
		Name:      role.RoleName,
		CreatedAt: formatTime(role.CreatedAt),
		UpdatedAt: formatTime(role.UpdatedAt),
		DeletedAt: wrapperspb.String(""),
	}

	return &pb.ApiResponseRoleDeleteAt{
		Status:  "success",
		Message: "Successfully restored role",
		Data:    protoRole,
	}, nil
}

func (s *roleCommandHandleGrpc) DeleteRolePermanent(ctx context.Context, req *pb.FindByIdRoleRequest) (*pb.ApiResponseRoleDelete, error) {
	roleID := int(req.GetRoleId())

	if roleID <= 0 {
		return nil, role_errors.ErrGrpcRoleInvalidId
	}

	_, err := s.roleCommand.DeleteRolePermanent(ctx, roleID)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseRoleDelete{
		Status:  "success",
		Message: "Successfully deleted role permanently",
	}, nil
}

func (s *roleCommandHandleGrpc) RestoreAllRole(ctx context.Context, _ *emptypb.Empty) (*pb.ApiResponseRoleAll, error) {
	_, err := s.roleCommand.RestoreAllRole(ctx)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseRoleAll{
		Status:  "success",
		Message: "Successfully restore all roles",
	}, nil
}

func (s *roleCommandHandleGrpc) DeleteAllRolePermanent(ctx context.Context, _ *emptypb.Empty) (*pb.ApiResponseRoleAll, error) {
	_, err := s.roleCommand.DeleteAllRolePermanent(ctx)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseRoleAll{
		Status:  "success",
		Message: "delete all roles permanent",
	}, nil
}
