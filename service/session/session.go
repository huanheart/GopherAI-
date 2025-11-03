package session

import (
	"GopherAI/common/aihelper"
	"GopherAI/common/code"
	"GopherAI/common/rabbitmq"
	"GopherAI/dao/session"
	"GopherAI/model"
)

func GetUserSessionsByUserName(userName string) ([]model.SessionInfo, error) {
	//获取用户的所有会话ID

	manager := aihelper.GetGlobalManager()
	Sessions := manager.GetUserSessions(userName)

	var SessionInfos []model.SessionInfo

	for _, session := range Sessions {
		SessionInfos = append(SessionInfos, model.SessionInfo{
			SessionID: session,
			Title:     session, // 暂时用sessionID作为标题，后续重构需要的时候可以更改
		})
	}

	return SessionInfos, nil
}

func CreateSessionAndSendMessage(userName string, userQuestion string, modelType string) (string, string, code.Code) {
	//1：创建一个新的会话
	newSession := &model.Session{
		UserName: userName,
		Title:    userQuestion, // 可以根据需求设置标题，这边暂时用用户第一次的问题作为标题
	}
	createdSession, err := session.CreateSession(newSession)
	if err != nil {
		return "", "", code.CodeServerBusy
	}

	//2：获取AIHelper并通过其管理消息
	manager := aihelper.GetGlobalManager()
	config := map[string]interface{}{
		"apiKey": "your-api-key", // TODO: 从配置中获取
	}
	helper, err := manager.GetOrCreateAIHelper(userName, createdSession.ID, modelType, config)
	if err != nil {
		return "", "", code.CodeServerBusy
	}

	// 添加用户消息到AIHelper并保存到数据库
	userMsg := model.Message{
		SessionID: createdSession.ID,
		Content:   userQuestion,
	}
	helper.AddMessage(userMsg)
	//同步插入到数据库
	// err = helper.SaveMessage(&userMsg, message.CreateMessage)
	//更改成消息队列异步处理
	err = helper.SaveMessage(&userMsg, func(message *model.Message) (*model.Message, error) {
		data := rabbitmq.GenerateMessageMQParam(createdSession.ID, userQuestion, userName)
		err := rabbitmq.RMQMessage.Publish(data)
		return message, err
	})

	if err != nil {
		return "", "", code.CodeServerBusy
	}

	//3：生成AI回复
	var aiResponse string
	aiResponse, err = helper.GenerateResponse(userQuestion)
	if err != nil {
		return "", "", code.CodeServerBusy
	}

	// 添加AI回复到AIHelper并保存到数据库
	aiMsg := model.Message{
		SessionID: createdSession.ID,
		Content:   aiResponse,
	}
	helper.AddMessage(aiMsg)
	//同步插入到数据库
	// err = helper.SaveMessage(&aiMsg, message.CreateMessage)
	//更改成消息队列异步处理
	err = helper.SaveMessage(&userMsg, func(message *model.Message) (*model.Message, error) {
		data := rabbitmq.GenerateMessageMQParam(createdSession.ID, userQuestion, userName)
		err := rabbitmq.RMQMessage.Publish(data)
		return message, err
	})
	if err != nil {
		return "", "", code.CodeServerBusy
	}

	return createdSession.ID, aiResponse, code.CodeSuccess
}

func ChatSend(userName string, sessionID string, userQuestion string, modelType string) (string, code.Code) {
	//1：获取AIHelper
	manager := aihelper.GetGlobalManager()
	config := map[string]interface{}{
		"apiKey": "your-api-key", // TODO: 从配置中获取
	}
	helper, err := manager.GetOrCreateAIHelper(userName, sessionID, modelType, config)
	if err != nil {
		return "", code.CodeServerBusy
	}

	// 添加用户消息到AIHelper并保存到数据库
	userMsg := model.Message{
		SessionID: sessionID,
		Content:   userQuestion,
	}
	helper.AddMessage(userMsg)
	//同步插入到数据库
	// err = helper.SaveMessage(&userMsg, message.CreateMessage)
	//更改成消息队列异步处理
	err = helper.SaveMessage(&userMsg, func(message *model.Message) (*model.Message, error) {
		data := rabbitmq.GenerateMessageMQParam(sessionID, userQuestion, userName)
		err := rabbitmq.RMQMessage.Publish(data)
		return message, err
	})

	if err != nil {
		return "", code.CodeServerBusy
	}

	//2：生成AI回复
	var aiResponse string
	aiResponse, err = helper.GenerateResponse(userQuestion)
	if err != nil {
		return "", code.CodeServerBusy
	}

	// 添加AI回复到AIHelper并保存到数据库
	aiMsg := model.Message{
		SessionID: sessionID,
		Content:   aiResponse,
	}
	helper.AddMessage(aiMsg)
	//同步插入到数据库
	// err = helper.SaveMessage(&userMsg, message.CreateMessage)
	//更改成消息队列异步处理
	err = helper.SaveMessage(&userMsg, func(message *model.Message) (*model.Message, error) {
		data := rabbitmq.GenerateMessageMQParam(sessionID, userQuestion, userName)
		err := rabbitmq.RMQMessage.Publish(data)
		return message, err
	})

	if err != nil {
		return "", code.CodeServerBusy
	}

	return aiResponse, code.CodeSuccess
}

func GetChatHistory(userName string, sessionID string) ([]model.History, code.Code) {
	// 获取AIHelper中的消息历史
	manager := aihelper.GetGlobalManager()
	helper, exists := manager.GetAIHelper(userName, sessionID)
	if !exists {
		return nil, code.CodeServerBusy
	}

	messages := helper.GetMessages()
	history := make([]model.History, 0, len(messages))

	// 转换消息为历史格式（根据消息顺序或内容判断用户/AI消息）
	for i, msg := range messages {
		isUser := i%2 == 0
		history = append(history, model.History{
			IsUser:  isUser,
			Content: msg.Content,
		})
	}

	return history, code.CodeSuccess
}
