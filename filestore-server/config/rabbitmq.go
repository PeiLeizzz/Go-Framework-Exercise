package config

const (
	AsyncTransferEnable = true
	RabbitURL           = ""
	TransExchangeName   = "uploadserver.trans"
	TransOSSQueueName   = "uploadserver.trans.oss"
	// oss 转移失败后写入另一个队列
	TransOssErrQueueName = "uploadserver.trans.oss.err"
	TransOSSRoutingKey   = "oss"
)
