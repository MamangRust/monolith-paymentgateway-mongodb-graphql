package repository

import "go.mongodb.org/mongo-driver/mongo"

// Repositories is a struct containing role command and query repositories.
type Repositories interface {
	RoleQueryRepository
	RoleCommandRepository
	UserRoleCommandRepository
}

type repositories struct {
	RoleQueryRepository
	RoleCommandRepository
	UserRoleCommandRepository
}

// NewRepositories creates a new instance of Repositories with the provided
// MongoDB database client.
func NewRepositories(db *mongo.Database) Repositories {
	return &repositories{
		RoleQueryRepository:       NewRoleQueryRepository(db.Collection("roles")),
		RoleCommandRepository:     NewRoleCommandRepository(db.Collection("roles")),
		UserRoleCommandRepository: NewUserRoleCommandRepository(db.Collection("user_roles")),
	}
}
