package utils

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func serveReverseProxy(target string, res http.ResponseWriter, req *http.Request) {
	url, _ := url.Parse(target)

	proxy := httputil.NewSingleHostReverseProxy(url)
	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host

	proxy.ServeHTTP(res, req)
}

/*
CreateProxyHandler Create gin handler for path parameter
*/
func GetGFWTile(c *gin.Context) {

	c.Request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", viper.GetString("gfw-token")))
	c.Request.Header.Add("Referer", "http://4wings-local-app")
	c.Request.URL.Path = strings.ReplaceAll(c.Request.URL.Path, "/v1", "/v2")

	serveReverseProxy("https://gateway.api.globalfishingwatch.org", c.Writer, c.Request)

}
