package aihelper

import (
	"fmt"
	"sync"
)

// ModelCreator 定义模型创建函数类型
type ModelCreator func(config map[string]interface{}) (AIModel, error)

// AIModelFactory AI模型工厂
type AIModelFactory struct {
	creators map[string]ModelCreator
}

// 全局工厂实例
var globalFactory *AIModelFactory
var factoryOnce sync.Once

// GetGlobalFactory 获取全局工厂实例
func GetGlobalFactory() *AIModelFactory {
	factoryOnce.Do(func() {
		globalFactory = &AIModelFactory{
			creators: make(map[string]ModelCreator),
		}
		globalFactory.registerCreators()
	})
	return globalFactory
}

// registerCreators 注册所有模型创建器
func (f *AIModelFactory) registerCreators() {
	f.creators["openai"] = func(config map[string]interface{}) (AIModel, error) {
		apiKey, ok := config["apiKey"].(string)
		if !ok {
			return nil, fmt.Errorf("OpenAI model requires apiKey")
		}
		return &OpenAIModel{apiKey: apiKey}, nil
	}

	f.creators["ollama"] = func(config map[string]interface{}) (AIModel, error) {
		modelName, ok := config["modelName"].(string)
		if !ok {
			return nil, fmt.Errorf("Ollama model requires modelName")
		}
		return &OllamaModel{modelName: modelName}, nil
	}
}

// CreateAIModel 根据模型类型创建对应的AI模型
func (f *AIModelFactory) CreateAIModel(modelType string, config map[string]interface{}) (AIModel, error) {
	creator, exists := f.creators[modelType]
	if !exists {
		return nil, fmt.Errorf("unsupported model type: %s", modelType)
	}
	return creator(config)
}

// CreateAIHelper 创建AIHelper实例
func (f *AIModelFactory) CreateAIHelper(modelType string, config map[string]interface{}) (*AIHelper, error) {
	model, err := f.CreateAIModel(modelType, config)
	if err != nil {
		return nil, err
	}
	return NewAIHelper(model), nil
}

// RegisterModel 动态注册新的模型类型（可选，用于扩展）
func (f *AIModelFactory) RegisterModel(modelType string, creator ModelCreator) {
	f.creators[modelType] = creator
}
