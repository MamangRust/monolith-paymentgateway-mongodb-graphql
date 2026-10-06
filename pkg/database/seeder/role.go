package seeder

import (
	"context"
	"fmt"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// roleSeeder is a struct that represents a seeder for the roles collection.
type roleSeeder struct {
	db     *mongo.Database
	ctx    context.Context
	logger logger.LoggerInterface
}

// NewRoleSeeder creates a new instance of the roleSeeder.
func NewRoleSeeder(db *mongo.Database, ctx context.Context, logger logger.LoggerInterface) *roleSeeder {
	return &roleSeeder{
		db:     db,
		ctx:    ctx,
		logger: logger,
	}
}

// Seed populates the roles collection with the default role set.
func (r *roleSeeder) Seed() error {
	randomRoles := []string{
		"Super Admin",
		"Admin",
		"Merchant Admin",
		"Merchant Operator",
		"Finance",
		"Compliance",
		"Auditor",
		"Support",
		"Viewer",
		"User",
	}

	totalRoles := len(randomRoles)

	docs := make([]interface{}, 0, totalRoles)
	for i, roleName := range randomRoles {
		roleID, err := salemongo.NextID(r.ctx, r.db, "roles")
		if err != nil {
			r.logger.Error("failed to allocate role id", zap.Int("role", i+1), zap.Error(err))
			return fmt.Errorf("failed to allocate role id: %w", err)
		}

		now := nowPtr()
		docs = append(docs, salemongo.Role{
			RoleID:    roleID,
			RoleName:  roleName,
			CreatedAt: now,
			UpdatedAt: now,
		})
	}

	if err := insertMany(r.ctx, r.db, "roles", docs, r.logger); err != nil {
		r.logger.Error("failed to seed roles", zap.Error(err))
		return fmt.Errorf("failed to seed roles: %w", err)
	}

	r.logger.Debug("role seeded successfully", zap.Int("totalRoles", totalRoles))
	return nil
}
