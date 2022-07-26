package main

import (
	"filestore-server/handler"
	"filestore-server/middleware"
	"fmt"
	"net/http"
)

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	http.HandleFunc("/file/upload", middleware.HTTPInterceptor(handler.UploadHandler))
	http.HandleFunc("/file/fastupload", middleware.HTTPInterceptor(handler.TryFastUploadHandler))
	http.HandleFunc("/file/upload/success", handler.UploadSuccessHandler)
	http.HandleFunc("/file/meta", middleware.HTTPInterceptor(handler.GetFileMetaHandler))
	http.HandleFunc("/file/query", middleware.HTTPInterceptor(handler.FileQueryHandler))
	http.HandleFunc("/file/download", middleware.HTTPInterceptor(handler.DownloadHandler))
	http.HandleFunc("/file/downloadurl", middleware.HTTPInterceptor(handler.DownloadURLHandler))
	http.HandleFunc("/file/update", middleware.HTTPInterceptor(handler.FileMetaUpdateHandler))
	http.HandleFunc("/file/delete", middleware.HTTPInterceptor(handler.FileDeleteHandler))

	http.HandleFunc("/file/mpupload/init", middleware.HTTPInterceptor(handler.InitialMultipartUploadHandler))
	http.HandleFunc("/file/mpupload/uppart", middleware.HTTPInterceptor(handler.UploadPartHandler))
	http.HandleFunc("/file/mpupload/complete", middleware.HTTPInterceptor(handler.CompleteUploadHandler))

	http.HandleFunc("/user/signup", handler.SignUpHandler)
	http.HandleFunc("/user/signin", handler.SignInHandler)
	http.HandleFunc("/user/info", middleware.HTTPInterceptor(handler.UserInfoHandler))

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Failed to start server, err: %s", err.Error())
	}
}
