package mq

import "filestore-server/common"

// 消息格式
type TransferData struct {
	FileHash      string
	CurLocation   string           // 临时存储地址
	DestLocation  string           // 要转移的目标地址
	DestStoreType common.StoreType // 存储类型（location/ceph/oss）
}
