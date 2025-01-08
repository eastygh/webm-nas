package server

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/eastygh/webm-nas/docs"
	"github.com/eastygh/webm-nas/pkg/authentication"
	"github.com/eastygh/webm-nas/pkg/authentication/oauth"
	"github.com/eastygh/webm-nas/pkg/authorization"
	"github.com/eastygh/webm-nas/pkg/common"
	"github.com/eastygh/webm-nas/pkg/config"
	"github.com/eastygh/webm-nas/pkg/controller"
	"github.com/eastygh/webm-nas/pkg/database"
	"github.com/eastygh/webm-nas/pkg/middleware"
	"github.com/eastygh/webm-nas/pkg/repository"
	"github.com/eastygh/webm-nas/pkg/service"
	"github.com/eastygh/webm-nas/pkg/utils/request"
	"github.com/eastygh/webm-nas/pkg/utils/set"
	"github.com/eastygh/webm-nas/pkg/version"
	"github.com/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func New(conf *config.Config, logger *logrus.Logger) (*Server, error) {
	rateLimitMiddleware, err := middleware.RateLimitMiddleware(conf.Server.LimitConfigs)
	if err != nil {
		return nil, err
	}

	var db *gorm.DB
	if conf.DB.Type == "sqlite" {
		db, err = database.NewSqlite(&conf.DB)
	} else {
		db, err = database.NewPostgres(&conf.DB)
	}
	if err != nil {
		return nil, errors.Wrap(err, "db init failed")
	}

	rdb, err := database.NewRedisClient(&conf.Redis)
	if err != nil {
		return nil, errors.Wrap(err, "redis client failed")
	}

	modelRepository := repository.NewRepository(db, rdb)
	if conf.DB.Migrate {
		if err := modelRepository.Migrate(); err != nil {
			return nil, err
		}
	}

	if err := modelRepository.Init(); err != nil {
		return nil, err
	}

	userService := service.NewUserService(modelRepository.User())
	groupService := service.NewGroupService(modelRepository.Group(), modelRepository.User())
	jwtService := authentication.NewJWTService(conf.Server.JWTSecret)
	rbacService := service.NewRBACService(modelRepository.RBAC())
	oauthManager := oauth.NewOAuthManager(conf.OAuthConfig)

	userController := controller.NewUserController(userService)
	groupController := controller.NewGroupController(groupService)
	authController := controller.NewAuthController(userService, jwtService, oauthManager)
	rbacController := controller.NewRbacController(rbacService)
	postController := controller.NewPostController(service.NewPostService(modelRepository.Post()))

	if err := authorization.InitAuthorization(modelRepository); err != nil {
		return nil, err
	}

	controllers := []controller.Controller{userController, groupController, authController, rbacController, postController}

	gin.SetMode(conf.Server.ENV)

	e := gin.New()
	e.Use(
		gin.Recovery(),
		rateLimitMiddleware,
		middleware.MonitorMiddleware(),
		middleware.CORSMiddleware(),
		middleware.RequestInfoMiddleware(&request.RequestInfoFactory{APIPrefixes: set.NewString("api")}),
		middleware.LogMiddleware(logger, "/"),
		middleware.AuthenticationMiddleware(jwtService, modelRepository.User()),
		middleware.AuthorizationMiddleware(),
		middleware.TraceMiddleware(),
	)

	e.LoadHTMLFiles("static/terminal.html")

	return &Server{
		engine:      e,
		config:      conf,
		logger:      logger,
		repository:  modelRepository,
		controllers: controllers,
	}, nil
}

type Server struct {
	engine *gin.Engine
	config *config.Config
	logger *logrus.Logger

	repository repository.Repository

	controllers []controller.Controller
}

// graceful shutdown
func (s *Server) Run() error {
	defer s.Close()

	s.initRouter()

	addr := fmt.Sprintf("%s:%d", s.config.Server.Address, s.config.Server.Port)
	s.logger.Infof("Start server on: %s", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: s.engine,
	}

	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			s.logger.Fatalf("Failed to start server, %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.config.Server.GracefulShutdownPeriod)*time.Second)
	defer cancel()

	ch := <-sig
	s.logger.Infof("Receive signal: %s", ch)

	return server.Shutdown(ctx)
}

func (s *Server) Close() {
	if err := s.repository.Close(); err != nil {
		s.logger.Warnf("failed to close repository, %v", err)
	}

}

func (s *Server) initRouter() {
	root := s.engine

	// Set if static content is enabled
	MapStaticContent(s.engine, &s.config.Static, s.logger)
	// Set if revers proxies are enabled
	CreateProxies(s.engine, &s.config.Revers, s.logger)

	// register non-resource routers
	root.GET("/routes", common.WrapFunc(s.getRoutes))

	root.GET("/index", controller.Index)
	root.GET("/healthz", common.WrapFunc(s.Ping))
	root.GET("/version", common.WrapFunc(version.Get))
	root.GET("/metrics", gin.WrapH(promhttp.Handler()))
	root.Any("/debug/pprof/*any", gin.WrapH(http.DefaultServeMux))
	if gin.Mode() != gin.ReleaseMode {
		root.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}

	api := root.Group("/api/v1")
	controllers := make([]string, 0, len(s.controllers))
	for _, router := range s.controllers {
		router.RegisterRoute(api)
		controllers = append(controllers, router.Name())
	}
	logrus.Infof("server enabled controllers: %v", controllers)
}

func (s *Server) getRoutes() []string {
	paths := set.NewString()
	for _, r := range s.engine.Routes() {
		if r.Path != "" {
			paths.Insert(r.Path)
		}
	}
	return paths.Slice()
}

type ServerStatus struct {
	Ping         bool `json:"ping"`
	DBRepository bool `json:"dbRepository"`
}

func (s *Server) Ping() *ServerStatus {
	status := &ServerStatus{Ping: true}

	ctx, cannel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cannel()

	if err := s.repository.Ping(ctx); err == nil {
		status.DBRepository = true
	}

	return status
}
