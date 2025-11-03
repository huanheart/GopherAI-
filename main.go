package main

import (
	"GopherAI/common/aihelper"
	"GopherAI/common/mysql"
	"GopherAI/common/rabbitmq"
	"GopherAI/common/redis"
	"GopherAI/config"
	"GopherAI/dao/message"
	"GopherAI/router"
	"fmt"
	"log"
)

func StartServer(addr string, port int) error {
	r := router.InitRouter()
	//服务器静态资源路径映射关系，这里目前不需要
	// r.Static(config.GetConfig().HttpFilePath, config.GetConfig().MusicFilePath)
	return r.Run(fmt.Sprintf("%s:%d", addr, port))
}

func readDataFromDB() error {
	manager := aihelper.GetGlobalManager()
	//从数据库中加载已有的会话和消息到内存中
	msgs, err := message.GetAllMessages()
	if err != nil {
		return err
	}

	// 2. 遍历并放入 manager
	for i := range msgs {
		m := &msgs[i]
		//默认类型
		modelType := "gpt-4"
		config := make(map[string]interface{})

		// 从 manager 获取或创建 AIHelper
		helper, err := manager.GetOrCreateAIHelper(m.UserName, m.SessionID, modelType, config)
		if err != nil {
			log.Printf("readDataFromDB: failed to create helper for user=%s session=%s: %v", m.UserName, m.SessionID, err)
			continue
		}

		// 添加消息到内存
		helper.AddMessage(*m)
	}

	log.Println("AIHelperManager init success  ", manager)
	return nil

}

func main() {
	conf := config.GetConfig()
	host := conf.MainConfig.Host
	port := conf.MainConfig.Port
	//初始化mysql
	if err := mysql.InitMysql(); err != nil {
		log.Println("InitMysql error , " + err.Error())
		return
	}
	//初始化AIHelperManager
	readDataFromDB()

	//初始化redis
	redis.Init()
	log.Println("redis init success  ")
	rabbitmq.InitRabbitMQ()
	log.Println("rabbitmq init success  ")

	err := StartServer(host, port) // 启动 HTTP 服务
	if err != nil {
		panic(err)
	}
}
