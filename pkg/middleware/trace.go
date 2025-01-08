package middleware

import (
	"time"

	"github.com/eastygh/webm-nas/pkg/common"
	utiltrace "github.com/eastygh/webm-nas/pkg/utils/trace"

	"github.com/bombsimon/logrusr/v2"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		trace := utiltrace.New("Handler",
			logrusr.New(logrus.StandardLogger()),
			utiltrace.Field{"method", c.Request.Method},
			utiltrace.Field{"path", c.Request.URL.Path},
		)

		defer trace.LogIfLong(100 * time.Millisecond)

		common.SetTrace(c, trace)

		c.Next()
	}
}
