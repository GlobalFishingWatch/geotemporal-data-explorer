package actions

import (
	"github.com/4wings/cli/internal/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func GenerateTileGFW(c *gin.Context) {
	log.Debugf("Obtaining tile of gfw with url %s", c.Request.URL)
	utils.GetGFWTile(c)
}
