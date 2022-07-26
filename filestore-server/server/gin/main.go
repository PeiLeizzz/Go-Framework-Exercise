package main

import "filestore-server/router"

func main() {
	r := router.Router()
	r.Run("localhost:8080")
}
