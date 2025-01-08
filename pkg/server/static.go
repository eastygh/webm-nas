package server

import (
	"github.com/eastygh/webm-nas/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
)

func MapStaticContent(engine *gin.Engine, config *config.StaticContentConfig, logger *logrus.Logger) {
	if config == nil || !config.Enable || len(config.Contents) == 0 {
		return
	}
	for k, v := range config.Contents {
		if k == "/" {
			// vite application dist
			staticContentDir := v
			logger.Warn("Static root path will use middleware.")
			engine.Use(func(c *gin.Context) {
				getStaticOrContinue(c, staticContentDir)
			})
			engine.NoRoute(func(c *gin.Context) {
				c.File(staticContentDir + "index.html")
			})
		} else {
			engine.Static(k, v)
		}
	}
}

func getStaticOrContinue(c *gin.Context, basePath string) {
	path := c.Request.URL.Path
	filePath := basePath + path
	if isFileInDirectory(filePath, basePath) && isFileExist(filePath) {
		c.File(filePath)
		c.Abort()
	}
}

// Check if the file is within the specified directory
func isFileInDirectory(filePath, directory string) bool {
	rel, err := filepath.Rel(directory, filePath)
	if err != nil || strings.Contains(rel, "..") {
		return false
	}
	return true
}

// Check if the file exists
func isFileExist(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}
