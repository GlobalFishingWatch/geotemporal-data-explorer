package middlewares

import (
	"github.com/4wings/cli/types"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func ErrorHandle() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		detectedErrors := c.Errors
		if len(detectedErrors) > 0 {
			err := detectedErrors[0].Err

			var parsedError *types.AppError
			switch err.(type) {
			case *types.AppError:
				parsedError = err.(*types.AppError)
				if parsedError.Code > 450 {
					log.Error("Error", err)
					parsedError.ErrorMessage = "Internal Server Error"
					parsedError.Messages = []types.MessageError{{
						Title:  "Internal Server Error",
						Detail: "Internal Server Error",
					}}
				} else {
					log.Debug(parsedError)
				}
			default:
				log.Error("Error", err)
				parsedError = types.NewUnprocessableEntityStandard([]types.MessageError{{
					Title:  "Internal Server Error",
					Detail: "Internal Server Error",
				}})

				// Put the error into response

			}
			c.AbortWithStatusJSON(parsedError.Code, parsedError)

			return
		}
	}
}
