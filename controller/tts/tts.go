package tts

import (
	"GopherAI/common/code"
	"GopherAI/common/tts"
	"GopherAI/controller"
	"net/http"

	"github.com/gin-gonic/gin"
)

type (
	TTSRequest struct {
		Text string `json:"text,omitempty"`
	}
	TTSResponse struct {
		TaskID string `json:"text,omitempty"`
		controller.Response
	}
	QueryTTSResponse struct {
		TaskID     string `json:"text,omitempty"`
		TaskStatus int    `json:"task_status,omitempty"`
		TaskResult string `json:"task_result,omitempty"`
		controller.Response
	}
)

type TTSServices struct {
	ttsService *tts.TTSService
}

func NewTTSServices() *TTSServices {
	return &TTSServices{
		ttsService: tts.NewTTSService(),
	}
}

func CreateTTSTask(c *gin.Context) {
	tts := NewTTSServices()
	req := new(TTSRequest)
	res := new(TTSResponse)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}

	if req.Text == "" {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}

	// 创建TTS任务并返回任务ID，由前端轮询查询结果
	taskID, err := tts.ttsService.CreateTTS(req.Text)
	if err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.TTSFail))
		return
	}

	res.Success()
	res.TaskID = taskID
	c.JSON(http.StatusOK, res)

}

func QueryTTSTask(c *gin.Context) {
	tts := NewTTSServices()
	res := new(QueryTTSResponse)
	taskID := c.Query("task_id")
	if taskID == "" {
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}

	response, err := tts.ttsService.QueryTTS(taskID)
	if err != nil {
		c.JSON(http.StatusOK, res.CodeOf(code.TTSFail))
		return
	}

	res.Success()
	res.TaskID = response.TaskID
	res.TaskResult = response.TaskResult
	res.TaskStatus = response.TaskStatus
	c.JSON(http.StatusOK, res)
}
