package main

import "filestore-server/rpc/apigw/router"

func main() {
	r := router.Router()
	r.Run(":8080")
}
