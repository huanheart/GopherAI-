package rabbitmq

var (
// RMQUpload          *RabbitMQ
)

func InitRabbitMQ() {
	// 创建MQ并启动消费者
	// 无论调用多少次 NewWorkRabbitMQ，只会创建一次连接
	// 不同队列共用一个连接，可以保持不同队列消费消息的顺序

	// RMQUpload = NewWorkRabbitMQ("Upload")
	// go RMQUpload.Consume(Upload)

}

// DestroyRabbitMQ 销毁RabbitMQ

func DestroyRabbitMQ() {
	// RMQUpload.Destroy()
}
