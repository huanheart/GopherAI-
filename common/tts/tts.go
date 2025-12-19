package tts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	API_KEY    = "Vl4nfabwhipYddiGMTXI1sXZ"
	SECRET_KEY = "hKzdAp1djkLLU1XsraadhVxLDjdkDVOU"
)

type TTSService struct{}

type TTSRequest struct {
	Text           string `json:"text"`
	Format         string `json:"format"`
	Voice          int    `json:"voice"`
	Lang           string `json:"lang"`
	Speed          int    `json:"speed"`
	Pitch          int    `json:"pitch"`
	Volume         int    `json:"volume"`
	EnableSubtitle int    `json:"enable_subtitle"`
}

type TTSResponse struct {
	TaskID string `json:"task_id"`
}

type QueryResponse struct {
	TaskID     string `json:"task_id"`
	TaskStatus int    `json:"task_status"`
	TaskResult string `json:"task_result"`
}

func NewTTSService() *TTSService {
	return &TTSService{}
}

// 创建TTS任务
func (s *TTSService) CreateTTS(text string) (string, error) {
	accessToken := s.GetAccessToken()
	if accessToken == "" {
		return "", fmt.Errorf("failed to get access token")
	}

	payload := TTSRequest{
		Text:           text,
		Format:         "mp3-16k",
		Voice:          4194,
		Lang:           "zh",
		Speed:          5,
		Pitch:          5,
		Volume:         5,
		EnableSubtitle: 0,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	url := "https://aip.baidubce.com/rpc/2.0/tts/v1/create?access_token=" + accessToken
	req, err := http.NewRequest("POST", url, bytes.NewReader(payloadBytes))
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	var ttsResponse TTSResponse
	if err := json.Unmarshal(body, &ttsResponse); err != nil {
		return "", err
	}

	return ttsResponse.TaskID, nil
}

// 查询TTS任务状态
func (s *TTSService) QueryTTS(taskID string) (QueryResponse, error) {
	accessToken := s.GetAccessToken()
	if accessToken == "" {
		return QueryResponse{}, fmt.Errorf("failed to get access token")
	}

	url := "https://aip.baidubce.com/rpc/2.0/tts/v1/query?access_token=" + accessToken
	payloadMap := map[string]string{"task_id": taskID}
	payloadBytes, _ := json.Marshal(payloadMap)

	req, err := http.NewRequest("POST", url, bytes.NewReader(payloadBytes))
	if err != nil {
		return QueryResponse{}, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return QueryResponse{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return QueryResponse{}, err
	}

	var queryResponse QueryResponse
	if err := json.Unmarshal(body, &queryResponse); err != nil {
		return QueryResponse{}, err
	}

	return queryResponse, nil
}

// 轮询获取TTS结果
func (s *TTSService) PollTTSResult(taskID string, timeout time.Duration) (string, error) {
	startTime := time.Now()
	for {
		if time.Since(startTime) > timeout {
			return "", fmt.Errorf("polling timeout")
		}

		response, err := s.QueryTTS(taskID)
		if err != nil {
			return "", err
		}

		if response.TaskStatus == 3 { // 完成
			return response.TaskResult, nil
		}

		time.Sleep(2 * time.Second)
	}
}

// 获取Access Token
func (s *TTSService) GetAccessToken() string {
	url := "https://aip.baidubce.com/oauth/2.0/token"
	postData := fmt.Sprintf("grant_type=client_credentials&client_id=%s&client_secret=%s", API_KEY, SECRET_KEY)
	resp, err := http.Post(url, "application/x-www-form-urlencoded", bytes.NewReader([]byte(postData)))
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	accessTokenObj := map[string]interface{}{}
	if err := json.Unmarshal(body, &accessTokenObj); err != nil {
		fmt.Println(err)
		return ""
	}

	if token, ok := accessTokenObj["access_token"].(string); ok {
		return token
	}
	return ""
}
