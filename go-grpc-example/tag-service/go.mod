module github.com/tag-service

go 1.18

replace github.com/tag-service => ./

require (
	github.com/golang/protobuf v1.5.2
	google.golang.org/genproto v0.0.0-20220505152158-f39f71e6c8f3
	google.golang.org/grpc v1.46.0
)

require (
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4 // indirect
	golang.org/x/sys v0.0.0-20210510120138-977fb7262007 // indirect
	golang.org/x/text v0.3.5 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
)
