package oss

import (
	"filestore-server/config"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"log"
)

var ossCli *oss.Client

func Client() *oss.Client {
	if ossCli != nil {
		return ossCli
	}

	cli, err := oss.New(config.OSSEndpoint, config.OSSAccesskeyID, config.OSSAccessKeySecret)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	ossCli = cli
	return ossCli
}

func Bucket() *oss.Bucket {
	cli := Client()
	if cli != nil {
		bucker, err := cli.Bucket(config.OSSBucket)
		if err != nil {
			log.Println(err.Error())
			return nil
		}
		return bucker
	}
	return nil
}

// 临时授权下载
func DownloadURL(objName string) string {
	signedURL, err := Bucket().SignURL(objName, oss.HTTPGet, 3600)
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	return signedURL
}

// 针对指定的 bucket 设置生命周期规则
func BuildLifecycleRule(bucketName string) {
	// 前缀 test/，30 天内无修改，就会被删除
	ruleTest := oss.BuildLifecycleRuleByDays("rule1", "test/", true, 30)
	rules := []oss.LifecycleRule{ruleTest}

	Client().SetBucketLifecycle(bucketName, rules)
}
