package handler

import (
	"context"

	pb "github.com/MamangRust/monolith-payment-gateway-pb/role"
	pbuserrole "github.com/MamangRust/monolith-payment-gateway-pb/user_role"
	"google.golang.org/protobuf/types/known/emptypb"
)

type RoleQueryHandlerGrpc interface {
	pb.RoleQueryServiceServer

	FindAllRole(ctx context.Context, req *pb.FindAllRoleRequest) (*pb.ApiResponsePaginationRole, error)
	FindByIdRole(ctx context.Context, req *pb.FindByIdRoleRequest) (*pb.ApiResponseRole, error)
	FindByNameRole(ctx context.Context, req *pb.FindByNameRoleRequest) (*pb.ApiResponseRole, error)
	FindByUserId(ctx context.Context, req *pb.FindByIdUserRoleRequest) (*pb.ApiResponsesRole, error)
	FindByActive(ctx context.Context, req *pb.FindAllRoleRequest) (*pb.ApiResponsePaginationRoleDeleteAt, error)
	FindByTrashed(ctx context.Context, req *pb.FindAllRoleRequest) (*pb.ApiResponsePaginationRoleDeleteAt, error)
}

type RoleCommandHandlerGrpc interface {
	pb.RoleCommandServiceServer

	CreateRole(ctx context.Context, req *pb.CreateRoleRequest) (*pb.ApiResponseRole, error)
	UpdateRole(ctx context.Context, req *pb.UpdateRoleRequest) (*pb.ApiResponseRole, error)
	TrashedRole(ctx context.Context, req *pb.FindByIdRoleRequest) (*pb.ApiResponseRoleDeleteAt, error)
	RestoreRole(ctx context.Context, req *pb.FindByIdRoleRequest) (*pb.ApiResponseRoleDeleteAt, error)
	DeleteRolePermanent(ctx context.Context, req *pb.FindByIdRoleRequest) (*pb.ApiResponseRoleDelete, error)
	RestoreAllRole(ctx context.Context, req *emptypb.Empty) (*pb.ApiResponseRoleAll, error)
	DeleteAllRolePermanent(ctx context.Context, req *emptypb.Empty) (*pb.ApiResponseRoleAll, error)
}

type UserRoleHandlerGrpc interface {
	pbuserrole.UserRoleServiceServer

	AssignRoleToUser(ctx context.Context, req *pbuserrole.AssignRoleToUserRequest) (*pbuserrole.ApiResponseUserRole, error)
	RemoveRoleFromUser(ctx context.Context, req *pbuserrole.RemoveRoleFromUserRequest) (*emptypb.Empty, error)
}
