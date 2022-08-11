package middlewares

import (
	"github.com/gin-gonic/gin"
)

func writeHeaders(c *gin.Context) {
	origin := c.Request.Header.Get("Origin")
	if len(origin) == 0 {
		origin = "*"
	}
	c.Header("Access-Control-Allow-Origin", origin)
	c.Header("Access-Control-Allow-Credentials", "true")
	c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Accept,Content-Type,cookie,refresh-token,Accept-Encoding,Authorization,origin,referer,user-agent,Access-Control-Allow-Origin,indexes-0,indexes-1,indexes-2,indexes-3,indexes-4,indexes-5,indexes-6,indexes-7")
	c.Header("Access-Control-Expose-Headers", "indexes-0,indexes-1,indexes-2,indexes-3,indexes-4,indexes-5,indexes-6,indexes-7")
}

func CorsMiddleware(c *gin.Context) {
	writeHeaders(c)

	if c.Request.Method != "OPTIONS" {
		c.Next()
	} else {
		c.Writer.WriteHeader(204)
		return
	}

	if c.IsAborted() {
		writeHeaders(c)
	}
}
