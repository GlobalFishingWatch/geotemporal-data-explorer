package middlewares

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/4wings/cli/types"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

func checkAllQueryParams(c *gin.Context, queryParamsSupported []string) error {
	var notMatch []string
	for k, _ := range c.Request.URL.Query() {
		_, ok := lo.Find(queryParamsSupported, func(i string) bool {
			match, _ := regexp.MatchString(fmt.Sprintf("%s$", i), k)
			return match
		})
		if !ok {
			notMatch = append(notMatch, k)
		}
	}
	if len(notMatch) > 0 {
		return types.NewUnprocessableEntityStandard([]types.MessageError{{
			Title:  "query-params",
			Detail: fmt.Sprintf("Query params %s not supported", strings.Join(notMatch, ",")),
		}})
	}
	return nil

}

func CheckQueryParams(queryParamsSupported []string) func(c *gin.Context) {
	return func(c *gin.Context) {
		err := checkAllQueryParams(c, queryParamsSupported)
		if err != nil {
			c.AbortWithStatusJSON(types.UnprocessableEntityCode, err)
			return
		}
		c.Next()
	}
}
