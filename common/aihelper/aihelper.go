package aihelper

import (
	"GopherAI/model"
	"sync"
)

// AIHelper AI助手结构体，包含消息历史和AI模型
type AIHelper struct {
	model    AIModel
	messages []*model.Message
	mu       sync.RWMutex
}

// NewAIHelper 创建新的AIHelper实例
func NewAIHelper(model_ AIModel) *AIHelper {
	return &AIHelper{
		model:    model_,
		messages: make([]*model.Message, 0),
	}
}

// AddMessage 添加消息到内存中（此时还没有保存到数据库中）
func (a *AIHelper) AddMessage(msg model.Message) {
	a.mu.Lock()
	defer a.mu.Unlock()
	// 创建消息的副本
	msgCopy := msg
	a.messages = append(a.messages, &msgCopy)
}

// SaveMessage 保存消息到数据库（通过回调函数避免循环依赖）
// 通过传入func，自己调用外部的保存函数，即可支持同步异步等多种策略
func (a *AIHelper) SaveMessage(msg *model.Message, saveFunc func(*model.Message) (*model.Message, error)) error {
	_, err := saveFunc(msg)
	return err
}

// GetMessages 获取所有消息历史
func (a *AIHelper) GetMessages() []model.Message {
	a.mu.RLock()
	defer a.mu.RUnlock()
	// 返回副本避免外部修改
	msgs := make([]model.Message, len(a.messages))
	for i, msg := range a.messages {
		msgs[i] = *msg
	}
	return msgs
}

// GenerateResponse 生成AI回复
func (a *AIHelper) GenerateResponse(userQuestion string) (string, error) {
	a.mu.RLock()
	messages := make([]model.Message, len(a.messages))
	for i, msg := range a.messages {
		messages[i] = *msg
	}
	a.mu.RUnlock()

	return a.model.GenerateResponse(messages, userQuestion)
}

// GetModelType 获取模型类型
func (a *AIHelper) GetModelType() string {
	return a.model.GetModelType()
}
