package httphandlers

import (
	"errors"
	"github.com/maksemen2/pvz-service/internal/service"
	"go.uber.org/zap"
	"net/http"

	"github.com/gin-gonic/gin"
	commonerrors "github.com/maksemen2/pvz-service/internal/common/errors"
	"github.com/maksemen2/pvz-service/internal/delivery/http/httpdto"
	domainerrors "github.com/maksemen2/pvz-service/internal/domain/errors"
)

// AuthHandler - обработчик аутентификации пользователей.
type AuthHandler struct {
	logger      *zap.Logger
	authService service.AuthService
}

func NewAuthHandler(logger *zap.Logger, authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		logger:      logger,
		authService: authService,
	}
}

func (h *AuthHandler) handleDomainError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domainerrors.ErrUnexpected):
		c.AbortWithStatusJSON(http.StatusInternalServerError, commonerrors.Internal())
	case errors.Is(err, domainerrors.ErrUserExists):
		c.AbortWithStatusJSON(http.StatusBadRequest, commonerrors.BadRequest("user already exists"))
	case errors.Is(err, domainerrors.ErrInvalidRole):
		c.AbortWithStatusJSON(http.StatusBadRequest, commonerrors.BadRequest("invalid role provided"))
	case errors.Is(err, domainerrors.ErrInvalidCredentials), errors.Is(err, domainerrors.ErrUserNotFound):
		c.AbortWithStatusJSON(http.StatusUnauthorized, commonerrors.InvalidCredentials()) // Лучше не указывать, что пользователь не найден
	case errors.Is(err, domainerrors.ErrPasswordTooLong):
		c.AbortWithStatusJSON(http.StatusBadRequest, commonerrors.BadRequest(err.Error()))
	default:
		c.AbortWithStatusJSON(http.StatusInternalServerError, commonerrors.Internal())
		h.logger.Error("unexpected error", zap.Error(err))
	}
}

func (h *AuthHandler) validateRegisterRequestBody(req httpdto.PostRegisterJSONRequestBody) bool {
	return req.Role != "" && req.Email != "" && req.Password != ""
}

func (h *AuthHandler) validateLoginRequestBody(req httpdto.PostLoginJSONRequestBody) bool {
	return req.Email != "" && req.Password != ""
}

func (h *AuthHandler) validateDummyLoginRequestBody(req httpdto.PostDummyLoginJSONRequestBody) bool {
	return req.Role != ""
}

func (h *AuthHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/register", h.HandleRegister)
	r.POST("/login", h.HandleLogin)
	r.POST("/dummyLogin", h.HandleDummyLogin)
}

func (h *AuthHandler) HandleRegister(c *gin.Context) {
	var req httpdto.PostRegisterJSONRequestBody

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Debug("BindJSON error handling register", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest, commonerrors.BadRequest("invalid request body"))

		return
	}

	if !h.validateRegisterRequestBody(req) {
		h.logger.Debug("invalid request body handling register")
		c.AbortWithStatusJSON(http.StatusBadRequest, commonerrors.BadRequest("invalid request body"))

		return
	}

	domainUser, err := h.authService.RegisterUser(c.Request.Context(), string(req.Email), req.Password, string(req.Role))

	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	c.JSON(http.StatusCreated, httpdto.ModelToUserResponse(domainUser))
}

func (h *AuthHandler) HandleLogin(c *gin.Context) {
	var req httpdto.PostLoginJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Debug("BindJSON error handling login", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest, commonerrors.BadRequest("invalid request body"))

		return
	}

	if !h.validateLoginRequestBody(req) {
		h.logger.Debug("invalid request body handling login")
		c.AbortWithStatusJSON(http.StatusBadRequest, commonerrors.BadRequest("invalid request body"))

		return
	}

	token, err := h.authService.AuthenticateUser(c.Request.Context(), string(req.Email), req.Password)

	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, httpdto.Token(token))
}

func (h *AuthHandler) HandleDummyLogin(c *gin.Context) {
	var req httpdto.PostDummyLoginJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Debug("BindJSON error handling dummy login", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest, commonerrors.BadRequest("invalid request body"))

		return
	}

	if !h.validateDummyLoginRequestBody(req) {
		h.logger.Debug("invalid request body handling dummy login")
		c.AbortWithStatusJSON(http.StatusBadRequest, commonerrors.BadRequest("invalid request body"))

		return
	}

	token, err := h.authService.DummyLogin(c.Request.Context(), string(req.Role))

	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, httpdto.Token(token))
}
