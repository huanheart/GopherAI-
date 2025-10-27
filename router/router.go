package router

import (
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {

	r := gin.Default()
	enterRouter := r.Group("/api/v1")
	{
		RegisterUserRouter(enterRouter.Group("/user"))
	}
	//后续登录的接口需要jwt鉴权
	// {
	// 	musicGroup := enterRouter.Group("/music")
	// 	musicGroup.Use(jwt.Auth())
	// 	MusicRouter(musicGroup)
	// }

	return r
}
