package aihelper

import (
	"GopherAI/model"
)

// AIModel 定义AI模型接口
type AIModel interface {
	GenerateResponse(messages []model.Message, userQuestion string) (string, error)
	GetModelType() string
}

// OpenAIModel OpenAI模型实现
type OpenAIModel struct {
	apiKey string
}

func (o *OpenAIModel) GenerateResponse(messages []model.Message, userQuestion string) (string, error) {
	// TODO: 实现OpenAI API调用
	return "OpenAI response: " + userQuestion, nil
}

func (o *OpenAIModel) GetModelType() string {
	return "openai"
}

// DeepSeekModel DeepSeek模型实现
type DeepSeekModel struct {
	apiKey string
}

func (d *DeepSeekModel) GenerateResponse(messages []model.Message, userQuestion string) (string, error) {
	// TODO: 实现DeepSeek API调用
	return "DeepSeek response: " + userQuestion, nil
}

func (d *DeepSeekModel) GetModelType() string {
	return "deepseek"
}

// OllamaModel Ollama模型实现
type OllamaModel struct {
	modelName string
}

func (o *OllamaModel) GenerateResponse(messages []model.Message, userQuestion string) (string, error) {
	// TODO: 实现Ollama API调用
	return "Ollama response: " + userQuestion, nil
}

func (o *OllamaModel) GetModelType() string {
	return "ollama"
}
