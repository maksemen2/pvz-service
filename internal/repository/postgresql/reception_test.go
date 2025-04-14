//go:build integration
// +build integration

package postgresqlrepo_test

import (
	"context"
	"testing"
	"time"

	databaseerrors "github.com/maksemen2/pvz-service/internal/repository/errors"

	"github.com/google/uuid"
	domainerrors "github.com/maksemen2/pvz-service/internal/domain/errors"
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

type ReceptionRepoTestSuite struct {
	suite.Suite
	ctx     context.Context
	db      *database.PostgresDB
	repo    repositories.IReceptionRepo
	cleanup func()
}

func TestReceptionRepoTestSuite(t *testing.T) {
	suite.Run(t, new(ReceptionRepoTestSuite))
}

func (s *ReceptionRepoTestSuite) SetupSuite() {
	s.ctx = context.Background()
	cfg, cleanContainer := testhelpers.SetupPostgresContainer(s.T())

	logger := zap.NewNop()

	var err error
	s.db, err = database.NewPostgresDB(cfg, logger)
	require.NoError(s.T(), err)

	s.repo = postgresqlrepo.NewPostgresqlReceptionRepository(s.db, logger)

	cleanDB, err := testhelpers.CreateTestDB(s.db)

	s.cleanup = func() {
		cleanDB()
		cleanContainer()
	}

	require.NoError(s.T(), err)
}

func (s *ReceptionRepoTestSuite) TearDownSuite() {
	s.db.Close()
	s.cleanup()
}

func (s *ReceptionRepoTestSuite) SetupTest() {
	_, err := s.db.Exec("DELETE FROM receptions")
	require.NoError(s.T(), err)
	_, err = s.db.Exec("DELETE FROM pvzs")
	require.NoError(s.T(), err)
}

func (s *ReceptionRepoTestSuite) createTestPVZ() uuid.UUID {
	pvzID := uuid.New()
	_, err := s.db.Exec(
		"INSERT INTO pvzs (id, registration_date, city) VALUES ($1, $2, $3)",
		pvzID, time.Now(), "Москва",
	)
	require.NoError(s.T(), err)

	return pvzID
}

func (s *ReceptionRepoTestSuite) createTestReception(pvzID uuid.UUID, status models.ReceptionStatus) *models.Reception {
	reception := &models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    pvzID,
		Status:   status,
	}
	_, err := s.db.Exec(
		"INSERT INTO receptions (id, date_time, pvz_id, status) VALUES ($1, $2, $3, $4)",
		reception.ID, reception.DateTime, reception.PVZID, reception.Status.String(),
	)
	require.NoError(s.T(), err)

	return reception
}

func (s *ReceptionRepoTestSuite) TestCreateIfNoOpen_Success() {
	pvzID := s.createTestPVZ()

	reception := &models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    pvzID,
		Status:   models.ReceptionStatusInProgress,
	}

	err := s.repo.CreateIfNoOpen(s.ctx, reception)
	assert.NoError(s.T(), err)

	var exists bool
	err = s.db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM receptions WHERE id = $1)", reception.ID)
	require.NoError(s.T(), err)
	assert.True(s.T(), exists)
}

func (s *ReceptionRepoTestSuite) TestCreateIfNoOpen_OpenExists() {
	pvzID := s.createTestPVZ()
	s.createTestReception(pvzID, models.ReceptionStatusInProgress)

	newReception := &models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    pvzID,
		Status:   models.ReceptionStatusInProgress,
	}

	err := s.repo.CreateIfNoOpen(s.ctx, newReception)
	assert.ErrorIs(s.T(), err, domainerrors.ErrOpenReceptionExists)
}

func (s *ReceptionRepoTestSuite) TestCreateIfNoOpen_PVZNotExists() {
	reception := &models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    uuid.New(), // этого пвз не существует
		Status:   models.ReceptionStatusInProgress,
	}

	err := s.repo.CreateIfNoOpen(s.ctx, reception)
	assert.ErrorIs(s.T(), err, databaseerrors.ErrNoRows)
}

func (s *ReceptionRepoTestSuite) TestCloseLast_Success() {
	pvzID := s.createTestPVZ()
	reception := s.createTestReception(pvzID, models.ReceptionStatusInProgress)

	closedReception, err := s.repo.CloseLast(s.ctx, pvzID)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), models.ReceptionStatusClose, closedReception.Status)
	assert.Equal(s.T(), reception.ID, closedReception.ID)

	var status string
	err = s.db.Get(&status, "SELECT status FROM receptions WHERE id = $1", reception.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), models.ReceptionStatusClose.String(), status)
}

func (s *ReceptionRepoTestSuite) TestCloseLast_NoOpenReceptions() {
	pvzID := s.createTestPVZ()

	_, err := s.repo.CloseLast(s.ctx, pvzID)
	assert.ErrorIs(s.T(), err, domainerrors.ErrNoOpenReceptions)
}

func (s *ReceptionRepoTestSuite) TestCloseLast_PVZNotExists() {
	_, err := s.repo.CloseLast(s.ctx, uuid.New())
	assert.ErrorIs(s.T(), err, domainerrors.ErrNoOpenReceptions)
}
