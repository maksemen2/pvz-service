package httphandlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	commonerrors "github.com/maksemen2/pvz-service/internal/common/errors"
	"github.com/maksemen2/pvz-service/internal/delivery/http/httpdto"
	domainerrors "github.com/maksemen2/pvz-service/internal/domain/errors"
	"github.com/maksemen2/pvz-service/internal/pkg/auth"
	"github.com/maksemen2/pvz-service/internal/service"
	"go.uber.org/zap"
	"net/http"
)

type PVZHandler struct {
	logger     *zap.Logger
	pzvService service.PVZService
}

func NewPVZHandler(logger *zap.Logger, pzvService service.PVZService) *PVZHandler {
	return &PVZHandler{
		logger:     logger,
		pzvService: pzvService,
	}
}

func (h *PVZHandler) handleDomainError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domainerrors.ErrUnexpected):
		c.AbortWithStatusJSON(http.StatusInternalServerError, commonerrors.Internal())
	case errors.Is(err, domainerrors.ErrInvalidCity):
		c.AbortWithStatusJSON(http.StatusBadRequest, commonerrors.BadRequest("invalid city provided"))
	case errors.Is(err, domainerrors.ErrUserNotModerator):
		c.AbortWithStatusJSON(http.StatusForbidden, commonerrors.Forbidden())
	case errors.Is(err, domainerrors.ErrPVZAlreadyExists):
		c.AbortWithStatusJSON(http.StatusBadRequest, commonerrors.BadRequest("pvz already exists"))
	case errors.Is(err, domainerrors.ErrInvalidLimit), errors.Is(err, domainerrors.ErrInvalidPage), errors.Is(err, domainerrors.ErrInvalidStartDate), errors.Is(err, domainerrors.ErrInvalidDateRange):
		c.AbortWithStatusJSON(http.StatusBadRequest, commonerrors.BadRequest(err.Error()))
	default:
		c.AbortWithStatusJSON(http.StatusInternalServerError, commonerrors.Internal())
		h.logger.Error("unexpected error", zap.Error(err))
	}
}

func (h *PVZHandler) RegisterRoutes(group *gin.RouterGroup) {
	group.POST("/pvz", h.HandleCreatePVZ)
	group.GET("/pvz", h.HandleListPVZ)
}

func (h *PVZHandler) HandleCreatePVZ(c *gin.Context) {
	userRole, ok := auth.GetRoleFromContext(c)
	if !ok {
		h.logger.Error("no role in context handling create pvz")
		c.AbortWithStatusJSON(http.StatusUnauthorized, commonerrors.Unauthorized())

		return
	}

	var req httpdto.PostPvzJSONRequestBody

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Debug("BindJSON error handling create pvz", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest, commonerrors.BadRequest("invalid request body"))

		return
	}

	domainPVZ, err := h.pzvService.CreatePVZ(c.Request.Context(), userRole, string(req.City), req.Id, req.RegistrationDate)

	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	c.JSON(http.StatusCreated, httpdto.ToPVZResponse(domainPVZ))
}

func (h *PVZHandler) HandleListPVZ(c *gin.Context) {
	userRole, ok := auth.GetRoleFromContext(c)
	if !ok {
		h.logger.Error("no role in context handling list pvz")
		c.AbortWithStatusJSON(http.StatusUnauthorized, commonerrors.Unauthorized())

		return
	}

	var query httpdto.GetPvzParams

	if err := c.ShouldBindQuery(&query); err != nil {
		h.logger.Debug("BindQuery error handling list pvz", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest, commonerrors.BadRequest("invalid query parameters"))

		return
	}

	pvzsWithReceptions, err := h.pzvService.ListPVZs(c.Request.Context(), userRole, query.StartDate, query.EndDate, query.Page, query.Limit)

	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	answer := make([]*httpdto.PVZWithReceptionsResponse, 0, len(pvzsWithReceptions))

	for _, pvz := range pvzsWithReceptions {
		answer = append(answer, httpdto.ModelToPVZWithReceptionsResponse(pvz))
	}

	c.JSON(http.StatusOK, answer)
}
