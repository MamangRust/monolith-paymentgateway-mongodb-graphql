package handler

import (
	"context"
	"math"

	pbhelper "github.com/MamangRust/monolith-payment-gateway-pb/common"
	pb "github.com/MamangRust/monolith-payment-gateway-pb/transfer"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/errors"
	transfer_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/transfer_errors/grpc"
	"github.com/MamangRust/monolith-payment-gateway-transfer/service"
)

type transferQueryHandleGrpc struct {
	pb.UnimplementedTransferQueryServiceServer

	transferQueryService service.TransferQueryService
}

func NewTransferQueryHandler(service service.TransferQueryService) TransferQueryHandleGrpc {
	return &transferQueryHandleGrpc{
		transferQueryService: service,
	}
}

func (s *transferQueryHandleGrpc) FindAllTransfer(ctx context.Context, request *pb.FindAllTransferRequest) (*pb.ApiResponsePaginationTransfer, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllTransfers{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	transfers, totalRecords, err := s.transferQueryService.FindAll(ctx, &reqService)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbhelper.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	transferResponses := make([]*pb.TransferResponse, len(transfers))
	for i, transfer := range transfers {
		transferResponses[i] = toTransferResponse(transfer.Transfer)
	}

	return &pb.ApiResponsePaginationTransfer{
		Status:         "success",
		Message:        "Successfully fetch transfer records",
		Data:           transferResponses,
		PaginationMeta: paginationMeta,
	}, nil
}

func (s *transferQueryHandleGrpc) FindByIdTransfer(ctx context.Context, request *pb.FindByIdTransferRequest) (*pb.ApiResponseTransfer, error) {
	id := int(request.GetTransferId())

	if id == 0 {
		return nil, transfer_errors.ErrGrpcTransferInvalidID
	}

	transfer, err := s.transferQueryService.FindById(ctx, id)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseTransfer{
		Status:  "success",
		Message: "Successfully fetch transfer record",
		Data:    toTransferResponse(transfer),
	}, nil
}

func (s *transferQueryHandleGrpc) FindByTransferByTransferFrom(ctx context.Context, request *pb.FindTransferByTransferFromRequest) (*pb.ApiResponseTransfers, error) {
	transfer_from := request.GetTransferFrom()

	if transfer_from == "" {
		return nil, transfer_errors.ErrGrpcInvalidCardNumber
	}

	transfers, err := s.transferQueryService.FindTransferByTransferFrom(ctx, transfer_from)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	transferResponses := make([]*pb.TransferResponse, len(transfers))
	for i, transfer := range transfers {
		transferResponses[i] = toTransferResponse(transfer)
	}

	return &pb.ApiResponseTransfers{
		Status:  "success",
		Message: "Successfully fetch transfer records",
		Data:    transferResponses,
	}, nil
}

func (s *transferQueryHandleGrpc) FindByTransferByTransferTo(ctx context.Context, request *pb.FindTransferByTransferToRequest) (*pb.ApiResponseTransfers, error) {
	transfer_to := request.GetTransferTo()

	if transfer_to == "" {
		return nil, transfer_errors.ErrGrpcInvalidCardNumber
	}

	transfers, err := s.transferQueryService.FindTransferByTransferTo(ctx, transfer_to)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	transferResponses := make([]*pb.TransferResponse, len(transfers))
	for i, transfer := range transfers {
		transferResponses[i] = toTransferResponse(transfer)
	}

	return &pb.ApiResponseTransfers{
		Status:  "success",
		Message: "Successfully fetch transfer records",
		Data:    transferResponses,
	}, nil
}

func (s *transferQueryHandleGrpc) FindByActiveTransfer(ctx context.Context, req *pb.FindAllTransferRequest) (*pb.ApiResponsePaginationTransferDeleteAt, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllTransfers{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	transfers, totalRecords, err := s.transferQueryService.FindByActive(ctx, &reqService)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbhelper.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	transferResponses := make([]*pb.TransferResponseDeleteAt, len(transfers))
	for i, transfer := range transfers {
		transferResponses[i] = toTransferDeleteAtResponse(transfer.Transfer)
	}

	return &pb.ApiResponsePaginationTransferDeleteAt{
		Status:         "success",
		Message:        "Successfully fetch transfer records",
		Data:           transferResponses,
		PaginationMeta: paginationMeta,
	}, nil
}

func (s *transferQueryHandleGrpc) FindByTrashedTransfer(ctx context.Context, req *pb.FindAllTransferRequest) (*pb.ApiResponsePaginationTransferDeleteAt, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllTransfers{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	transfers, totalRecords, err := s.transferQueryService.FindByTrashed(ctx, &reqService)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbhelper.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	transferResponses := make([]*pb.TransferResponseDeleteAt, len(transfers))
	for i, transfer := range transfers {
		transferResponses[i] = toTransferDeleteAtResponse(transfer.Transfer)
	}

	return &pb.ApiResponsePaginationTransferDeleteAt{
		Status:         "success",
		Message:        "Successfully fetch transfer records",
		Data:           transferResponses,
		PaginationMeta: paginationMeta,
	}, nil
}
