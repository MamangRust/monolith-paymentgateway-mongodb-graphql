package seeder

import (
	"context"
	"fmt"
	"math/rand"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/hash"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// userSeeder is a struct that represents a seeder for the users collection.
type userSeeder struct {
	db     *mongo.Database
	hash   hash.HashPassword
	ctx    context.Context
	logger logger.LoggerInterface
}

// NewUserSeeder creates a new instance of the userSeeder.
func NewUserSeeder(db *mongo.Database, ctx context.Context, hash hash.HashPassword, logger logger.LoggerInterface) *userSeeder {
	return &userSeeder{
		db:     db,
		ctx:    ctx,
		hash:   hash,
		logger: logger,
	}
}

// Seed populates the users collection with 15 users, 10 active and 5 trashed,
// assigning a random role to each user.
func (r *userSeeder) Seed() error {
	totalUsers := 15
	activeUsers := 10
	trashedUsers := totalUsers - activeUsers

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

	for i := 1; i <= totalUsers; i++ {
		email := fmt.Sprintf("user_%s@example.com", uuid.NewString())

		hashed, err := r.hash.HashPassword("password")
		if err != nil {
			r.logger.Error("failed to hash password", zap.Int("user", i), zap.Error(err))
			return fmt.Errorf("failed to hash password for user %d: %w", i, err)
		}

		userID, err := salemongo.NextID(r.ctx, r.db, "users")
		if err != nil {
			r.logger.Error("failed to allocate user id", zap.Int("user", i), zap.Error(err))
			return fmt.Errorf("failed to allocate user id: %w", err)
		}

		isVerified := true
		now := nowPtr()

		user := &salemongo.User{
			UserID:           userID,
			Firstname:        fmt.Sprintf("User%d", i),
			Lastname:         fmt.Sprintf("Last%d", i),
			Email:            email,
			Password:         hashed,
			VerificationCode: uuid.NewString(),
			IsVerified:       &isVerified,
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		if err := insertOne(r.ctx, r.db, "users", user, r.logger); err != nil {
			r.logger.Error("failed to create user", zap.Int("user", i), zap.Error(err))
			return fmt.Errorf("failed to create user %d: %w", i, err)
		}

		randomRole := randomRoles[rand.Intn(len(randomRoles))]

		var role salemongo.Role
		if err := r.db.Collection("roles").FindOne(r.ctx, bson.M{"role_name": randomRole, "deleted_at": nil}).Decode(&role); err != nil {
			r.logger.Error("failed to get role", zap.String("role", randomRole), zap.Error(err))
			return fmt.Errorf("failed to get role %s: %w", randomRole, err)
		}

		userRoleID, err := salemongo.NextID(r.ctx, r.db, "user_roles")
		if err != nil {
			r.logger.Error("failed to allocate user role id", zap.Int("user", i), zap.Error(err))
			return fmt.Errorf("failed to allocate user role id: %w", err)
		}

		userRole := &salemongo.UserRole{
			UserRoleID: userRoleID,
			UserID:     user.UserID,
			RoleID:     role.RoleID,
			CreatedAt:  nowPtr(),
			UpdatedAt:  nowPtr(),
		}
		if err := insertOne(r.ctx, r.db, "user_roles", userRole, r.logger); err != nil {
			r.logger.Error("failed to assign role to user", zap.Int("userID", int(user.UserID)), zap.String("role", randomRole), zap.Error(err))
			return fmt.Errorf("failed to assign role %s to user %d: %w", randomRole, user.UserID, err)
		}

		if i > activeUsers {
			if err := trashDocument(r.ctx, r.db, "users", "user_id", user.UserID); err != nil {
				r.logger.Error("failed to trash user", zap.Int("user", i), zap.Error(err))
				return fmt.Errorf("failed to trash user %d: %w", i, err)
			}
		}
	}

	r.logger.Info("user seeded successfully", zap.Int("totalUsers", totalUsers), zap.Int("activeUsers", activeUsers), zap.Int("trashedUsers", trashedUsers))

	return nil
}
