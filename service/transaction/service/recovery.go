package service

import (
	"context"
	"time"

	"github.com/MamangRust/monolith-payment-gateway-shared/state"
	"go.uber.org/zap"
)

// RecoverStuckOperations claims stale processing/compensation rows and moves
// them to unknown. Since the database currently records operation state but not
// per-leg settlement receipts, recovery must not guess whether a debit or
// credit happened; unknown is the safe reconciliation state.
func (s *transactionCommandService) RecoverStuckOperations(ctx context.Context, olderThan time.Duration, maxRows int32) error {
	rows, err := s.transactionCommandRepository.ListStuck(ctx, olderThan, maxRows)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row.Status == state.Unknown {
			continue
		}
		if _, err := s.transactionCommandRepository.TransitionStatus(ctx, int(row.TransactionID), row.Status, state.Unknown, "recovery could not prove settlement outcome"); err != nil {
			s.logger.Warn("recovery: failed to mark transaction unknown", zap.Error(err), zap.Int32("transaction_id", row.TransactionID))
		}
	}
	return nil
}

func (s *transactionCommandService) StartRecoveryWorker(ctx context.Context, interval, olderThan time.Duration, maxRows int32) {
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
				s.logger.Error("transaction recovery worker failed", zap.Error(err))
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}
