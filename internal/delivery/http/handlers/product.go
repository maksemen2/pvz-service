package httphandlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	commonerrors "github.com/maksemen2/pvz-service/internal/common/errors"
	"github.com/maksemen2/pvz-service/internal/delivery/http/httpdto"
	domainerrors "github.com/maksemen2/pvz-service/internal/domain/errors"
	"github.com/maksemen2/pvz-service/internal/pkg/auth"
	"github.com/maksemen2/pvz-service/internal/service"
	"go.uber.org/zap"
	"net/http"
)

type ProductHandler struct {
	logger         *zap.Logger
	productService service.ProductService
}

func NewProductHandler(logger *zap.Logger, productService service.ProductService) *ProductHandler {
	return &ProductHandler{
		logger:         logger,
		productService: productService,
	}
}

func (h *ProductHandler) handleDomainError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domainerrors.ErrUnexpected):
		c.AbortWithStatusJSON(http.StatusInternalServerError, commonerrors.Internal())
	case errors.Is(err, domainerrors.ErrNotEnoughRights):
		c.AbortWithStatusJSON(http.StatusForbidden, commonerrors.Forbidden())
	case errors.Is(err, domainerrors.ErrNoOpenReceptions), errors.Is(err, domainerrors.ErrInvalidProductType), errors.Is(err, domainerrors.ErrNoProductsInReception):
		c.AbortWithStatusJSON(http.StatusBadRequest, commonerrors.BadRequest(err.Error()))
	default:
		c.AbortWithStatusJSON(http.StatusInternalServerError, commonerrors.Internal())
		h.logger.Error("unexpected error", zap.Error(err))
	}
}

func (h *ProductHandler) RegisterRoutes(group *gin.RouterGroup) {
	group.POST("/pvz/:pvzId/delete_last_product", h.HandleDeleteLastProduct)
	group.POST("/products", h.HandleAddProduct)
}

func (h *ProductHandler) HandleDeleteLastProduct(c *gin.Context) {
	role, ok := auth.GetRoleFromContext(c)

	if !ok {
		// Мы считаем это ошибкой, потому что мидлварь вообще не должен был пропускать такой запрос
		h.logger.Error("no role in context handling delete last product")
		c.AbortWithStatusJSON(http.StatusForbidden, commonerrors.Forbidden())

		return
	}

	pvzID := c.Param("pvzId")

	pvzUUID, err := uuid.Parse(pvzID)

	if err != nil {
		h.logger.Debug("invalid pvzID", zap.String("pvzID", pvzID), zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest, commonerrors.BadRequest("invalid pvzID"))

		return
	}

	err = h.productService.DeleteLastProduct(c.Request.Context(), role, pvzUUID)

	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *ProductHandler) HandleAddProduct(c *gin.Context) {
	role, ok := auth.GetRoleFromContext(c)

	if !ok {
		h.logger.Error("no role in context handling add product")
		c.AbortWithStatusJSON(http.StatusForbidden, commonerrors.Forbidden())

		return
	}

	var req httpdto.PostProductsJSONRequestBody

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Debug("BindJSON error handling add product", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest, commonerrors.BadRequest("invalid request body"))

		return
	}

	domainProduct, err := h.productService.AddProduct(c.Request.Context(), role, string(req.Type), req.PvzId)

	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	c.JSON(http.StatusCreated, httpdto.ModelToProductResponse(domainProduct))
}
