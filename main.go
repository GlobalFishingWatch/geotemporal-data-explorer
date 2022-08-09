package main

import (
	"os"

	"github.com/4wings/cli/cmd"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetOutput(os.Stdout)

	if os.Getenv("ENV") == "pro" {
		log.SetFormatter(&log.JSONFormatter{})
		log.SetLevel(log.InfoLevel)
		gin.SetMode(gin.ReleaseMode)
	} else {
		log.SetFormatter(&log.TextFormatter{})
		log.SetLevel(log.DebugLevel)
	}
}

func main() {
	cmd.Execute()
}
