//go:build integration
// +build integration

package postgresqlrepo_test

import (
	"context"
	databaseerrors "github.com/maksemen2/pvz-service/internal/repository/errors"
	"testing"

	"github.com/google/uuid"
	"github.com/maksemen2/pvz-service/internal/domain/models"
	"github.com/maksemen2/pvz-service/internal/domain/repositories"
	"github.com/maksemen2/pvz-service/internal/pkg/database"
	"github.com/maksemen2/pvz-service/internal/pkg/testhelpers"
	postgresqlrepo "github.com/maksemen2/pvz-service/internal/repository/postgresql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type UserRepoTestSuite struct {
	suite.Suite
	ctx     context.Context
	db      *database.PostgresDB
	repo    repositories.IUserRepo
	cleanup func()
}

func TestUserRepoTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepoTestSuite))
}

func (s *UserRepoTestSuite) SetupSuite() {
	s.ctx = context.Background()
	cfg, cleanContainer := testhelpers.SetupPostgresContainer(s.T())

	logger := zap.NewNop()

	var err error
	s.db, err = database.NewPostgresDB(cfg, logger)
	require.NoError(s.T(), err)

	s.repo = postgresqlrepo.NewPostgresqlUserRepository(s.db, logger)

	cleanDB, err := testhelpers.CreateTestDB(s.db)

	s.cleanup = func() {
		cleanDB()
		cleanContainer()
	}

	require.NoError(s.T(), err)
}

func (s *UserRepoTestSuite) TearDownSuite() {
	s.db.Close()
	s.cleanup()
}

func (s *UserRepoTestSuite) SetupTest() {
	_, err := s.db.Exec("DELETE FROM users")
	require.NoError(s.T(), err)
}

func (s *UserRepoTestSuite) createTestUser() *models.User {
	user := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         models.RoleEmployee,
	}
	err := s.repo.Create(s.ctx, user)
	require.NoError(s.T(), err)

	return user
}

func (s *UserRepoTestSuite) TestCreateUser_Success() {
	user := &models.User{
		ID:           uuid.New(),
		Email:        "new@example.com",
		PasswordHash: "new_hashed_password",
		Role:         models.RoleModerator,
	}

	err := s.repo.Create(s.ctx, user)
	assert.NoError(s.T(), err)

	var exists bool
	err = s.db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", user.ID)
	require.NoError(s.T(), err)
	assert.True(s.T(), exists)
}

func (s *UserRepoTestSuite) TestCreateUser_DuplicateEmail() {
	user := s.createTestUser()

	newUser := &models.User{
		ID:           uuid.New(),
		Email:        user.Email,
		PasswordHash: "another_hash",
		Role:         models.RoleModerator,
	}

	err := s.repo.Create(s.ctx, newUser)
	assert.ErrorIs(s.T(), err, databaseerrors.ErrUniqueViolation)
}

func (s *UserRepoTestSuite) TestGetByEmail_Success() {
	createdUser := s.createTestUser()

	foundUser, err := s.repo.GetByEmail(s.ctx, createdUser.Email)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), createdUser.ID, foundUser.ID)
	assert.Equal(s.T(), createdUser.Email, foundUser.Email)
	assert.Equal(s.T(), createdUser.PasswordHash, foundUser.PasswordHash)
	assert.Equal(s.T(), createdUser.Role, foundUser.Role)
}

func (s *UserRepoTestSuite) TestGetByEmail_NotFound() {
	_, err := s.repo.GetByEmail(s.ctx, "nonexistent@example.com")
	assert.ErrorIs(s.T(), err, databaseerrors.ErrNoRows)
}
