package ceph

import (
	"gopkg.in/amz.v1/aws"
	"gopkg.in/amz.v1/s3"
)

// 集群部署：https://blog.csdn.net/Eoneanyna/article/details/112304809
var cephConn *s3.S3

func GetCephConnection() *s3.S3 {
	if cephConn != nil {
		return cephConn
	}

	// 初始化 ceph
	auth := aws.Auth{
		AccessKey: "TEC5W8Y86HT9Q6JRXA61",
		SecretKey: "uwyRwsoWnNnTqyqI6zgvVPNw95yrC3NsQQkPWFVY",
	}

	curRegion := aws.Region{
		Name:                 "default",
		EC2Endpoint:          "http://192.168.10.107:9080",
		S3Endpoint:           "http://192.168.10.107:9080",
		S3BucketEndpoint:     "",
		S3LocationConstraint: false,
		S3LowercaseBucket:    false,
		Sign:                 aws.SignV2,
	}

	// 创建 S3 类型的连接
	cephConn = s3.New(auth, curRegion)
	return cephConn
}

// GetCephBucket：获取指定的 bucket 对象
func GetCephBucket(bucket string) *s3.Bucket {
	conn := GetCephConnection()
	return conn.Bucket(bucket)
}
