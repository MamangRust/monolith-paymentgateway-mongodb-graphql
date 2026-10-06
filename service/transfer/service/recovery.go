package service

import (
	"context"
	"time"

	"github.com/MamangRust/monolith-payment-gateway-shared/state"
	"go.uber.org/zap"
)

// RecoverStuckOperations claims stale rows and moves them to unknown. Without
// durable per-leg receipts, guessing a compensating debit/credit would risk a
// second money movement; unknown is deliberately handed to reconciliation.
func (s *transferCommandService) RecoverStuckOperations(ctx context.Context, olderThan time.Duration, maxRows int32) error {
	rows, err := s.transferCommandRepository.ListStuck(ctx, olderThan, maxRows)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row.Status == state.Unknown {
			continue
		}
		if _, err := s.transferCommandRepository.TransitionStatus(ctx, int(row.TransferID), row.Status, state.Unknown, "recovery could not prove settlement outcome"); err != nil {
			s.logger.Warn("recovery: failed to mark transfer unknown", zap.Error(err), zap.Int32("transfer_id", row.TransferID))
		}
	}
	return nil
}

func (s *transferCommandService) StartRecoveryWorker(ctx context.Context, interval, olderThan time.Duration, maxRows int32) {
	if interval <= 0 {
		interval = time.Minute
	}
	if olderThan <= 0 {
		olderThan = 2 * interval
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			if err := s.RecoverStuckOperations(ctx, olderThan, maxRows); err != nil {
				s.logger.Error("transfer recovery worker failed", zap.Error(err))
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}
