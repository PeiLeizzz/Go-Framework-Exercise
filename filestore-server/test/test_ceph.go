package main

import (
	"filestore-server/store/ceph"
	"os"
)

func main() {
	bucket := ceph.GetCephBucket("userfile")

	d, _ := bucket.Get("/ceph/8fac262b2197cb07e91ef44f258e9e55e5c4ad3f")
	tmpFile, _ := os.Create("./tmp/test_file")
	tmpFile.Write(d)
	return

	//// 创建一个新的 bucket
	//err := bucket.PutBucket(s3.PublicRead)
	//if err != nil {
	//	fmt.Printf("create bucket err: %s\n", err.Error())
	//}
	//
	//// 查询这个 bucket 下面指定条件的 object keys
	//res, err := bucket.List("", "", "", 100)
	//fmt.Printf("object keys: %+v\n", res)
	//
	//// 新上传一个对象
	//bucket.Put("/testupload/a.txt", []byte("just for test"), "octet-stream", s3.PublicRead)
	//if err != nil {
	//	fmt.Printf("upload err: %s\n", err.Error())
	//}
	//
	//// 查询这个 bucket 下面指定条件的 object keys
	//res, err = bucket.List("", "", "", 100)
	//fmt.Printf("object keys: %+v\n", res)
}
