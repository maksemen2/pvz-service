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

type ReceptionHandler struct {
	receptionService service.ReceptionService
	logger           *zap.Logger
}

func NewReceptionHandler(logger *zap.Logger, receptionService service.ReceptionService) *ReceptionHandler {
	return &ReceptionHandler{
		receptionService: receptionService,
		logger:           logger,
	}
}

func (h *ReceptionHandler) handleDomainError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domainerrors.ErrUnexpected):
		c.AbortWithStatusJSON(http.StatusInternalServerError, commonerrors.Internal())
	case errors.Is(err, domainerrors.ErrNoOpenReceptions), errors.Is(err, domainerrors.ErrOpenReceptionExists), errors.Is(err, domainerrors.ErrPVZNotFound):
		c.AbortWithStatusJSON(http.StatusBadRequest, commonerrors.BadRequest(err.Error()))
	case errors.Is(err, domainerrors.ErrNotEnoughRights):
		c.AbortWithStatusJSON(http.StatusForbidden, commonerrors.Forbidden())
	default:
		c.AbortWithStatusJSON(http.StatusInternalServerError, commonerrors.Internal())
		h.logger.Error("unexpected error", zap.Error(err))
	}
}

func (h *ReceptionHandler) RegisterRoutes(group *gin.RouterGroup) {
	group.POST("/pvz/:pvzId/close_last_reception", h.HandleCloseLastReception)
	group.POST("/receptions", h.HandleCreateReception)
}

func (h *ReceptionHandler) HandleCloseLastReception(c *gin.Context) {
	role, ok := auth.GetRoleFromContext(c)

	if !ok {
		h.logger.Error("Failed to get role from context")
		c.AbortWithStatusJSON(http.StatusForbidden, commonerrors.Forbidden())

		return
	}

	pvzID := c.Param("pvzId")

	pvzUUID, err := uuid.Parse(pvzID)

	if err != nil {
		h.logger.Debug("invalid pvzID", zap.String("pvzID", pvzID))
		c.AbortWithStatusJSON(http.StatusBadRequest, commonerrors.BadRequest("invalid pvzID"))

		return
	}

	domainReception, err := h.receptionService.CloseLastReception(c.Request.Context(), role, pvzUUID)

	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, httpdto.ModelToReceptionResponse(domainReception))
}

func (h *ReceptionHandler) HandleCreateReception(c *gin.Context) {
	role, ok := auth.GetRoleFromContext(c)

	if !ok {
		h.logger.Error("Failed to get role from context")
		c.AbortWithStatusJSON(http.StatusForbidden, commonerrors.Forbidden())

		return
	}

	var req httpdto.PostReceptionsJSONRequestBody

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Debug("BindJSON error handling create reception", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest, commonerrors.BadRequest("invalid request body"))

		return
	}

	domainReception, err := h.receptionService.CreateReceptionIfNoOpen(c.Request.Context(), role, req.PvzId)

	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	c.JSON(http.StatusCreated, httpdto.ModelToReceptionResponse(domainReception))
}
