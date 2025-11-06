package image

import (
	"GopherAI/service/image"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RecognizeImageHandler(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status_msg": "Image upload failed: " + err.Error(),
		})
		return
	}

	className, err := image.RecognizeImage(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status_msg": "Image recognition failed: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"class_name": className,
	})
}
