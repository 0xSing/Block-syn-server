package middleware

import (
	"github.com/gin-gonic/gin"
	"walletSynV2/pkg"
)

// 授权中间件
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authentication")
		if token == "" {
			resp := pkg.MakeResp(pkg.AuthenticationEmpty, nil)
			c.JSON(resp.HttpCode, resp)
			c.Abort()
			return
		}

		if token != "f0610db4-1276-462b-82fa-875373fc0650" {
			resp := pkg.MakeResp(pkg.AuthenticationInvalid, nil)
			c.JSON(resp.HttpCode, resp)
			c.Abort()
			return
		}

		c.Next()
	}
}
