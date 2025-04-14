//go:build integration
// +build integration

package postgresqlrepo_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/maksemen2/pvz-service/internal/domain/errors"
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

type ProductRepoTestSuite struct {
	suite.Suite
	ctx     context.Context
	db      *database.PostgresDB
	repo    repositories.IProductRepo
	cleanup func()
}

func TestProductRepoTestSuite(t *testing.T) {
	suite.Run(t, new(ProductRepoTestSuite))
}

func (s *ProductRepoTestSuite) SetupSuite() {
	s.ctx = context.Background()
	cfg, cleanContainer := testhelpers.SetupPostgresContainer(s.T())

	logger := zap.NewNop()

	var err error
	s.db, err = database.NewPostgresDB(cfg, logger)
	require.NoError(s.T(), err)

	s.repo = postgresqlrepo.NewPostgresqlProductRepository(s.db, logger)

	cleanDB, err := testhelpers.CreateTestDB(s.db)

	s.cleanup = func() {
		cleanDB()
		cleanContainer()
	}

	require.NoError(s.T(), err)
}

func (s *ProductRepoTestSuite) TearDownSuite() {
	s.db.Close()
	s.cleanup()
}

func (s *ProductRepoTestSuite) SetupTest() {
	_, err := s.db.Exec("DELETE FROM products")
	require.NoError(s.T(), err)
	_, err = s.db.Exec("DELETE FROM receptions")
	require.NoError(s.T(), err)
}

// хелпер для создания приемки
func (s *ProductRepoTestSuite) createReception(pvzID uuid.UUID, status string) uuid.UUID {
	receptionID := uuid.New()
	query := `INSERT INTO receptions (id, pvz_id, date_time, status) VALUES ($1, $2, $3, $4)`
	_, err := s.db.Exec(query, receptionID, pvzID, time.Now(), status)
	require.NoError(s.T(), err)

	return receptionID
}

// хелпер для создания ПВЗ (для самих тестов не требуется, но нужно, чтобы не нарушать foreign key constraint)
func (s *ProductRepoTestSuite) createPVZ() uuid.UUID {
	pvzID := uuid.New()
	query := `INSERT INTO pvzs (id, registration_date, city) VALUES ($1, $2, $3)`
	_, err := s.db.Exec(query, pvzID, time.Now(), "Москва")
	require.NoError(s.T(), err)

	return pvzID
}

func (s *ProductRepoTestSuite) TestCreateProduct_Success() {
	pvzID := s.createPVZ()
	s.createReception(pvzID, "in_progress")

	product := &models.AddProduct{
		ID:       uuid.New(),
		DateTime: time.Now(),
		Type:     models.ProductTypeElectronics,
		PVZID:    pvzID,
	}

	created, err := s.repo.Create(s.ctx, product)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), product.ID, created.ID)
	assert.Equal(s.T(), product.Type, created.Type)
}

func (s *ProductRepoTestSuite) TestCreateProduct_NoOpenReception() {
	product := &models.AddProduct{
		ID:       uuid.New(),
		DateTime: time.Now(),
		Type:     models.ProductTypeClothes,
		PVZID:    uuid.New(),
	}

	_, err := s.repo.Create(s.ctx, product)
	assert.ErrorIs(s.T(), err, domainerrors.ErrNoOpenReceptions)
}

func (s *ProductRepoTestSuite) TestDeleteLast_Success() {
	pvzID := s.createPVZ()
	receptionID := s.createReception(pvzID, "in_progress")

	productID := uuid.New()
	query := `INSERT INTO products (id, date_time, type, reception_id) VALUES ($1, $2, $3, $4)`
	_, err := s.db.Exec(query, productID, time.Now(), "food", receptionID)
	require.NoError(s.T(), err)

	err = s.repo.DeleteLast(s.ctx, pvzID)
	require.NoError(s.T(), err)

	var count int
	err = s.db.Get(&count, "SELECT COUNT(*) FROM products WHERE id = $1", productID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 0, count)
}

func (s *ProductRepoTestSuite) TestDeleteLast_NoOpenReception() {
	err := s.repo.DeleteLast(s.ctx, uuid.New())
	assert.ErrorIs(s.T(), err, domainerrors.ErrNoOpenReceptions)
}

func (s *ProductRepoTestSuite) TestDeleteLast_NoProducts() {
	pvzID := s.createPVZ()
	s.createReception(pvzID, "in_progress")

	err := s.repo.DeleteLast(s.ctx, pvzID)
	assert.ErrorIs(s.T(), err, domainerrors.ErrNoProductsInReception)
}
