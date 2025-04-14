//go:build integration
// +build integration

package postgresqlrepo_test

import (
	"context"
	"testing"
	"time"

	databaseerrors "github.com/maksemen2/pvz-service/internal/repository/errors"

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

type PVZRepoTestSuite struct {
	suite.Suite
	ctx     context.Context
	db      *database.PostgresDB
	repo    repositories.IPVZRepo
	cleanup func()
}

func TestPVZRepoTestSuite(t *testing.T) {
	suite.Run(t, new(PVZRepoTestSuite))
}

func (s *PVZRepoTestSuite) SetupSuite() {
	s.ctx = context.Background()
	cfg, cleanContainer := testhelpers.SetupPostgresContainer(s.T())

	logger := zap.NewNop()

	var err error
	s.db, err = database.NewPostgresDB(cfg, logger)
	require.NoError(s.T(), err)

	s.repo = postgresqlrepo.NewPostgresqlPVZRepository(s.db, logger)

	// Инициализация схемы для тестов
	cleanDB, err := testhelpers.CreateTestDB(s.db)

	s.cleanup = func() {
		cleanDB()
		cleanContainer()
	}

	require.NoError(s.T(), err)
}

func (s *PVZRepoTestSuite) TearDownSuite() {
	s.db.Close()
	s.cleanup()
}

func (s *PVZRepoTestSuite) SetupTest() {
	// Очистка данных перед каждым тестом
	_, err := s.db.Exec("DELETE FROM products")
	require.NoError(s.T(), err)
	_, err = s.db.Exec("DELETE FROM receptions")
	require.NoError(s.T(), err)
	_, err = s.db.Exec("DELETE FROM pvzs")
	require.NoError(s.T(), err)
}

// хелпер для создания тестового ПВЗ
func (s *PVZRepoTestSuite) createTestPVZ() *models.PVZ {
	pvz := &models.PVZ{
		ID:               uuid.New(),
		RegistrationDate: time.Now(),
		City:             models.CityTypeMoscow,
	}
	err := s.repo.Create(s.ctx, pvz)
	require.NoError(s.T(), err)

	return pvz
}

// хелпер для создания тестовой приёмки
func (s *PVZRepoTestSuite) createTestReception(pvzID uuid.UUID, status string, date time.Time) uuid.UUID {
	receptionID := uuid.New()
	_, err := s.db.Exec(
		"INSERT INTO receptions (id, pvz_id, date_time, status) VALUES ($1, $2, $3, $4)",
		receptionID, pvzID, date, status,
	)
	require.NoError(s.T(), err)

	return receptionID
}

// хелпер для создания тестового продукта
func (s *PVZRepoTestSuite) createTestProduct(receptionID uuid.UUID, productType models.ProductType, date time.Time) {
	_, err := s.db.Exec(
		"INSERT INTO products (id, date_time, type, reception_id) VALUES ($1, $2, $3, $4)",
		uuid.New(), date, productType.String(), receptionID,
	)
	require.NoError(s.T(), err)
}

func (s *PVZRepoTestSuite) TestCreatePVZ_Success() {
	pvz := &models.PVZ{
		ID:               uuid.New(),
		RegistrationDate: time.Now(),
		City:             models.CityTypeSPB,
	}

	err := s.repo.Create(s.ctx, pvz)
	assert.NoError(s.T(), err)

	var exists bool
	err = s.db.Get(&exists, "SELECT EXISTS (SELECT 1 FROM pvzs WHERE id = $1)", pvz.ID)
	require.NoError(s.T(), err)
	assert.True(s.T(), exists)
}

func (s *PVZRepoTestSuite) TestCreatePVZ_DuplicateID() {
	pvz := s.createTestPVZ()

	newPVZ := &models.PVZ{
		ID:               pvz.ID,
		RegistrationDate: time.Now(),
		City:             models.CityTypeKazan,
	}

	err := s.repo.Create(s.ctx, newPVZ)
	assert.ErrorIs(s.T(), err, databaseerrors.ErrUniqueViolation)
}

func (s *PVZRepoTestSuite) TestListPVZs_WithoutDateFilter() {
	pvz1 := s.createTestPVZ()
	pvz2 := s.createTestPVZ()
	s.createTestPVZ() // это будет пвз без приемок

	// Добавляем приёмки только для первых двух PVZ
	now := time.Now()
	s.createTestReception(pvz1.ID, "in_progress", now)
	s.createTestReception(pvz2.ID, "close", now.Add(-24*time.Hour))

	filter := &models.PVZFilter{
		Page:     1,
		PageSize: 10,
	}

	result, err := s.repo.List(s.ctx, filter)
	require.NoError(s.T(), err)

	assert.Len(s.T(), result, 3)
}

func (s *PVZRepoTestSuite) TestListPVZs_WithDateFilter() {
	pvz1 := s.createTestPVZ()
	pvz2 := s.createTestPVZ()
	now := time.Now()

	// Приёмки для pvz1
	s.createTestReception(pvz1.ID, "in_progress", now.Add(-24*time.Hour))
	s.createTestReception(pvz1.ID, "close", now.Add(-12*time.Hour))

	// Приёмки для pvz2 вне диапазона
	s.createTestReception(pvz2.ID, "in_progress", now.Add(-48*time.Hour))

	sD := now.Add(-36 * time.Hour)
	eD := now.Add(-6 * time.Hour)

	filter := &models.PVZFilter{
		Page:      1,
		PageSize:  10,
		StartDate: &sD,
		EndDate:   &eD,
	}

	result, err := s.repo.List(s.ctx, filter)
	require.NoError(s.T(), err)

	// Должен вернуться только pvz1
	assert.Len(s.T(), result, 1)
	assert.Equal(s.T(), pvz1.ID, result[0].PVZ.ID)
}

func (s *PVZRepoTestSuite) TestListPVZs_Pagination() {
	for i := 0; i < 5; i++ {
		s.createTestPVZ()
	}

	filter := &models.PVZFilter{
		Page:     2,
		PageSize: 2,
	}

	result, err := s.repo.List(s.ctx, filter)
	require.NoError(s.T(), err)

	// offset будет (2 - 1) * 2 = 2, лимит - 2, ожидаем 2 элемента
	assert.Len(s.T(), result, 2)

	filter.Page = 3
	result, err = s.repo.List(s.ctx, filter)
	require.NoError(s.T(), err)

	// offset будет (3 - 1) * 2 = 4, лимит - 2, соответственно ожидаем 1 оставшийся элемент
	assert.Len(s.T(), result, 1)

}

func (s *PVZRepoTestSuite) TestGetAllPVZs() {
	for i := 0; i < 5; i++ {
		s.createTestPVZ()
	}

	result, err := s.repo.GetAll(s.ctx)
	require.NoError(s.T(), err)
	assert.Len(s.T(), result, 5)
}

func (s *PVZRepoTestSuite) TestListPVZs_WithProducts() {
	pvz := s.createTestPVZ()
	receptionID := s.createTestReception(pvz.ID, "in_progress", time.Now())
	s.createTestProduct(receptionID, models.ProductTypeClothes, time.Now())

	filter := &models.PVZFilter{
		Page:     1,
		PageSize: 10,
	}

	result, err := s.repo.List(s.ctx, filter)
	require.NoError(s.T(), err)

	require.Len(s.T(), result, 1)
	require.Len(s.T(), result[0].Receptions, 1)
	assert.Len(s.T(), result[0].Receptions[0].Products, 1)
}
