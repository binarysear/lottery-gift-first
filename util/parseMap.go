package util

import "github.com/gin-gonic/gin"

func MustMapBodyFrom(c *gin.Context) map[string]interface{} {
	var body map[string]interface{}
	err := c.BindJSON(&body)
	LogRus.Warn(err)
	return body
}
