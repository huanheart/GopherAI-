package router

import (
	"GopherAI/controller/image"

	"github.com/gin-gonic/gin"
)

func ImageRouter(r *gin.Engine) {
	imageRouter := r.Group("/image")
	{
		imageRouter.POST("/recognize", image.RecognizeImageHandler)
	}
}
