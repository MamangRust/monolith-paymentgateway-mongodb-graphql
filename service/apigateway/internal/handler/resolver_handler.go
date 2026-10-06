package graph

import (
	"context"
	"fmt"

	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type ResolverFunction[T any] func(ctx context.Context) (T, error)

type resolverHandler struct {
	observability observability.TraceLoggerObservability
	logger        logger.LoggerInterface
}

func NewResolverHandler(obs observability.TraceLoggerObservability, logger logger.LoggerInterface) *resolverHandler {
	return &resolverHandler{
		observability: obs,
		logger:        logger,
	}
}

func ResolverHandle[T any](h *resolverHandler, method string, ctx context.Context, fn ResolverFunction[T]) (T, error) {
	var zero T

	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(
		ctx,
		method,
	)
	defer func() {
		end(status)
	}()

	result, err := fn(ctx)
	if err != nil {
		status = "error"
		h.handleResolverError(err, span, method)
		return zero, err
	}

	logSuccess("Resolver completed successfully")
	return result, nil
}

func (h *resolverHandler) handleResolverError(err error, span trace.Span, method string) {
	traceID := span.SpanContext().TraceID().String()

	h.logger.Error(
		fmt.Sprintf("Resolver error in %s", method),
		zap.Error(err),
		zap.String("trace.id", traceID),
	)

	span.SetAttributes(
		attribute.String("trace.id", traceID),
		attribute.String("error", err.Error()),
	)
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}
