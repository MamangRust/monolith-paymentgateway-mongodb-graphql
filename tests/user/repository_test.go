package user_test

import (
	"context"
	"fmt"
	models "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"testing"
	"time"

	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/MamangRust/monolith-payment-gateway-user/repository"

	"github.com/stretchr/testify/suite"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	ts   *tests.TestSuite
	repo repository.Repositories
}

func (s *UserRepositoryTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	mongoDB, err := s.ts.MongoDB()
	if err != nil {
		s.Require().NoError(err)
	}
	s.repo = repository.NewRepositories(mongoDB, nil, nil)
}

func (s *UserRepositoryTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *UserRepositoryTestSuite) createSeedUser() (*models.User, error) {
	return s.repo.UserCommand().CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName:       "User",
		LastName:        "Tester",
		Email:           fmt.Sprintf("user.tester-%d@example.com", time.Now().UnixNano()),
		Password:        "password123",
		ConfirmPassword: "password123",
	})
}

func (s *UserRepositoryTestSuite) TestCreateUser() {
	ctx := context.Background()
	req := &requests.CreateUserRequest{
		FirstName:       "User",
		LastName:        "Tester",
		Email:           fmt.Sprintf("user.tester.%d@example.com", time.Now().UnixNano()),
		Password:        "password123",
		ConfirmPassword: "password123",
	}

	res, err := s.repo.UserCommand().CreateUser(ctx, req)
	s.NoError(err)
	s.NotNil(res)
}

func (s *UserRepositoryTestSuite) TestFindAllUsers() {
	_, err := s.createSeedUser()
	s.Require().NoError(err)
	ctx := context.Background()

	res, err := s.repo.UserQuery().FindAllUsers(ctx, &requests.FindAllUsers{
		Page:     1,
		PageSize: 10,
		Search:   "",
	})
	s.NoError(err)
	s.GreaterOrEqual(len(res), 1)
}

func (s *UserRepositoryTestSuite) TestFindById() {
	user, err := s.createSeedUser()
	s.Require().NoError(err)
	ctx := context.Background()

	found, err := s.repo.UserQuery().FindById(ctx, int(user.UserID))
	s.NoError(err)
	s.NotNil(found)
	s.Equal(user.UserID, found.UserID)
}

func (s *UserRepositoryTestSuite) TestFindByActive() {
	_, err := s.createSeedUser()
	s.Require().NoError(err)
	ctx := context.Background()

	res, err := s.repo.UserQuery().FindByActive(ctx, &requests.FindAllUsers{
		Page:     1,
		PageSize: 10,
		Search:   "",
	})
	s.NoError(err)
	s.GreaterOrEqual(len(res), 1)
}

func (s *UserRepositoryTestSuite) TestFindByTrashed() {
	user, err := s.createSeedUser()
	s.Require().NoError(err)
	ctx := context.Background()

	_, err = s.repo.UserCommand().TrashedUser(ctx, int(user.UserID))
	s.Require().NoError(err)

	res, err := s.repo.UserQuery().FindByTrashed(ctx, &requests.FindAllUsers{
		Page:     1,
		PageSize: 10,
		Search:   "",
	})
	s.NoError(err)
	s.GreaterOrEqual(len(res), 1)
}

func (s *UserRepositoryTestSuite) TestUpdateUser() {
	user, err := s.createSeedUser()
	s.Require().NoError(err)
	ctx := context.Background()

	id := int(user.UserID)
	req := &requests.UpdateUserRequest{
		UserID:          &id,
		FirstName:       "Updated",
		LastName:        "User",
		Email:           user.Email,
		Password:        "newpassword123",
		ConfirmPassword: "newpassword123",
	}

	res, err := s.repo.UserCommand().UpdateUser(ctx, req)
	s.NoError(err)
	s.NotNil(res)
}

func (s *UserRepositoryTestSuite) TestTrashUser() {
	user, err := s.createSeedUser()
	s.Require().NoError(err)
	ctx := context.Background()

	trashed, err := s.repo.UserCommand().TrashedUser(ctx, int(user.UserID))
	s.NoError(err)
	s.NotNil(trashed)
}

func (s *UserRepositoryTestSuite) TestRestoreUser() {
	user, err := s.createSeedUser()
	s.Require().NoError(err)
	ctx := context.Background()

	_, err = s.repo.UserCommand().TrashedUser(ctx, int(user.UserID))
	s.Require().NoError(err)

	restored, err := s.repo.UserCommand().RestoreUser(ctx, int(user.UserID))
	s.NoError(err)
	s.NotNil(restored)
}

func (s *UserRepositoryTestSuite) TestDeleteUserPermanent() {
	user, err := s.createSeedUser()
	s.Require().NoError(err)
	ctx := context.Background()

	_, err = s.repo.UserCommand().TrashedUser(ctx, int(user.UserID))
	s.Require().NoError(err)

	success, err := s.repo.UserCommand().DeleteUserPermanent(ctx, int(user.UserID))
	s.NoError(err)
	s.True(success)
}

func (s *UserRepositoryTestSuite) TestRestoreAllUser() {
	user, err := s.createSeedUser()
	s.Require().NoError(err)
	ctx := context.Background()

	_, err = s.repo.UserCommand().TrashedUser(ctx, int(user.UserID))
	s.Require().NoError(err)

	success, err := s.repo.UserCommand().RestoreAllUser(ctx)
	s.NoError(err)
	s.True(success)
}

func (s *UserRepositoryTestSuite) TestDeleteAllUserPermanent() {
	_, err := s.createSeedUser()
	s.Require().NoError(err)
	ctx := context.Background()

	success, err := s.repo.UserCommand().DeleteAllUserPermanent(ctx)
	s.NoError(err)
	s.True(success)
}

func TestUserRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(UserRepositoryTestSuite))
}
