package server

import (
	"github.com/eastygh/webm-nas/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func CreateProxies(engine *gin.Engine, config *config.ReversProxyConfig, logger *logrus.Logger) {
	if config == nil || !config.Enable || len(config.ProxyUrls) == 0 {
		return
	}
	//for k, v := range config.ProxyUrls {
	//	targetURL, err := url.Parse(v)
	//	if err != nil {
	//		logger.Error("Error while parsing reverse proxy url", err)
	//	}
	//	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	//	engine.Any(k, func(c *gin.Context) {
	//		proxy.ServeHTTP(c.Writer, c.Request)
	//	})
	//}
}
