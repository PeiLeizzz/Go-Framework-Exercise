package config

const (
	AsyncTransferEnable = true
	RabbitURL           = "amqp://peileiuser:peilei777@127.0.0.1:5672/peilei"
	TransExchangeName   = "uploadserver.trans"
	TransOSSQueueName   = "uploadserver.trans.oss"
	// oss 转移失败后写入另一个队列
	TransOssErrQueueName = "uploadserver.trans.oss.err"
	TransOSSRoutingKey   = "oss"
)
