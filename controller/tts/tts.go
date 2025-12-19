package tts

import (
	"GopherAI/common/tts"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TTSController struct {
	ttsService *tts.TTSService
}

func NewTTSController() *TTSController {
	return &TTSController{
		ttsService: tts.NewTTSService(),
	}
}

func (c *TTSController) CreateTTSTask(ctx *gin.Context) {
	var request struct {
		Text string `json:"text"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status_code": 1001,
			"status_msg":  "Invalid request body",
		})
		return
	}

	if request.Text == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status_code": 1002,
			"status_msg":  "Text cannot be empty",
		})
		return
	}

	// 创建TTS任务并返回任务ID，由前端轮询查询结果
	taskID, err := c.ttsService.CreateTTS(request.Text)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status_code": 1003,
			"status_msg":  "Failed to create TTS task",
			"error":       err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status_code": 1000,
		"status_msg":  "Success",
		"task_id":     taskID,
	})
}

func (c *TTSController) QueryTTSTask(ctx *gin.Context) {
	taskID := ctx.Query("task_id")
	if taskID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status_code": 1001,
			"status_msg":  "Task ID cannot be empty",
		})
		return
	}

	response, err := c.ttsService.QueryTTS(taskID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status_code": 1003,
			"status_msg":  "Failed to query TTS task",
			"error":       err.Error(),
		})
		return
	}

	// 返回任务状态和结果
	ctx.JSON(http.StatusOK, gin.H{
		"status_code": 1000,
		"status_msg":  "Success",
		"task_id":     response.TaskID,
		"task_status": response.TaskStatus,
		"task_result": response.TaskResult,
	})
}
