module github.com/go-grpc-example

go 1.18

require (
	github.com/golang/protobuf v1.5.2
	google.golang.org/grpc v1.29.1
)

require (
	golang.org/x/net v0.0.0-20220425223048-2871e0cb64e4 // indirect
	golang.org/x/sys v0.0.0-20220503163025-988cb79eb6c6 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
)

replace github.com/go-grpc-example/proto => ./proto
