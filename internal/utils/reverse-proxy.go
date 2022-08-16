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

func rewriteBody(resp *http.Response) (err error) {
	resp.Header.Del("Access-Control-Allow-Origin")
	resp.Header.Del("Access-Control-Allow-Credentials")
	resp.Header.Del("Access-Control-Allow-Headers")
	resp.Header.Del("Access-Control-Allow-Methods")
	resp.Header.Del("Access-Control-Expose-Headers")
	return nil
}

func serveReverseProxy(target string, res http.ResponseWriter, req *http.Request) {
	url, _ := url.Parse(target)

	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ModifyResponse = rewriteBody
	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Host = url.Host

	proxy.ServeHTTP(res, req)
}

/*
CreateProxyHandler Create gin handler for path parameter
*/
func GetGFWTile(c *gin.Context) {

	c.Request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", viper.GetString("gfw-token")))
	c.Request.Header.Add("Referer", "http://geotemporal-data-explorer")
	c.Request.URL.Path = strings.ReplaceAll(c.Request.URL.Path, "/v1", "/v2")

	serveReverseProxy("https://gateway.api.globalfishingwatch.org", c.Writer, c.Request)

}
