package handler

import (
	"filestore-byceph/common"
	"filestore-byceph/utils"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)


//HTTPInterceptor:http拦截器
func HTTPInterceptor() gin.HandlerFunc{
	return func(c *gin.Context) {
				username :=c.Request.FormValue("username")
				token := c.Request.FormValue("token")
				log.Println(username)
				log.Println(token)

				if len(username) < 3 || !ValidToken(token) {
					c.Abort()
					resp := utils.RespMsg{
						Code: int(common.StatusParamInvalid),
						Msg:  "token无效",
					}
					c.JSON(http.StatusOK, resp)
					return
				}
				c.Next()
	}
}