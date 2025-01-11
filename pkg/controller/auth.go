package controller

import (
	"encoding/json"
	"github.com/eastygh/webm-nas/pkg/config"
	"net/http"

	"github.com/eastygh/webm-nas/pkg/authentication"
	"github.com/eastygh/webm-nas/pkg/common"
	"github.com/eastygh/webm-nas/pkg/model"
	"github.com/eastygh/webm-nas/pkg/service"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	userService service.UserService
	jwtService  *authentication.JWTService
	config      *config.Config
}

func NewAuthController(userService service.UserService, jwtService *authentication.JWTService, config *config.Config) Controller {
	return &AuthController{
		userService: userService,
		jwtService:  jwtService,
		config:      config,
	}
}

// @Summary Login
// @Description User login
// @Accept json
// @Produce json
// @Tags auth
// @Param user body model.AuthUser true "auth user info"
// @Success 200 {object} common.Response{data=model.JWTToken}
// @Router /api/v1/auth/token [post]
func (ac *AuthController) Login(c *gin.Context) {
	auser := new(model.AuthUser)
	if err := c.BindJSON(auser); err != nil {
		common.ResponseFailed(c, http.StatusBadRequest, err)
		return
	}

	var user *model.User
	var err error

	user, err = ac.userService.Auth(auser)

	if err != nil {
		common.ResponseFailed(c, http.StatusUnauthorized, err)
		return
	}

	token, err := ac.jwtService.CreateToken(user)
	if err != nil {
		common.ResponseFailed(c, http.StatusInternalServerError, err)
		return
	}

	userJson, err := json.Marshal(user)
	if err != nil {
		common.ResponseFailed(c, http.StatusInternalServerError, err)
		return
	}
	if auser.SetCookie {
		var secure = ac.getCookieSecureFlag(c)
		c.SetCookie(common.CookieTokenName, token, 3600*24, "/", "", secure, true)
		c.SetCookie(common.CookieLoginUser, string(userJson), 3600*24, "/", "", secure, false)
	}

	common.ResponseSuccess(c, model.JWTToken{
		Token:    token,
		Describe: "set token in Authorization Header, [Authorization: Bearer {token}]",
	})
}

// @Summary Logout
// @Description User logout
// @Produce json
// @Tags auth
// @Success 200 {object} common.Response
// @Router /api/v1/auth/token [delete]
func (ac *AuthController) Logout(c *gin.Context) {
	var secure = ac.getCookieSecureFlag(c)
	c.SetCookie(common.CookieTokenName, "", -1, "/", "", secure, true)
	c.SetCookie(common.CookieLoginUser, "", -1, "/", "", secure, false)
	common.ResponseSuccess(c, nil)
}

// @Summary Register user
// @Description Create user and storage
// @Accept json
// @Produce json
// @Tags auth
// @Param user body model.CreatedUser true "user info"
// @Success 200 {object} common.Response{data=model.User}
// @Router /api/v1/auth/user [post]
func (ac *AuthController) Register(c *gin.Context) {
	createdUser := new(model.CreatedUser)
	if err := c.BindJSON(createdUser); err != nil {
		common.ResponseFailed(c, http.StatusBadRequest, err)
		return
	}

	user := createdUser.GetUser()
	if err := ac.userService.Validate(user); err != nil {
		common.ResponseFailed(c, http.StatusBadRequest, err)
		return
	}

	ac.userService.Default(user)
	user, err := ac.userService.Create(user)
	if err != nil {
		common.ResponseFailed(c, http.StatusInternalServerError, err)
	}

	common.ResponseSuccess(c, user)
}

func (ac *AuthController) RegisterRoute(api *gin.RouterGroup) {
	api.POST("/auth/token", ac.Login)
	api.DELETE("/auth/token", ac.Logout)
	api.POST("/auth/user", ac.Register)
}

func (ac *AuthController) Name() string {
	return "Authentication"
}

func (ac *AuthController) getCookieSecureFlag(c *gin.Context) bool {
	return c.Request.TLS == nil && ac.config.Server.AllowInsecure
}
