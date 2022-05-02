# 声明 build/clean/tool/lint/help 为伪目标
.PHONY: build clean tool lint help

# make all / make
all: build

go-gin-example: main.go
	go build -v .
	
build:
	go build -v .

tool:
	go tool vet . |& grep -v vendor; true
	gofmt -w .

lint:
	golint ./...

clean:
	rm -rf go-gin-example
	go clean -i .

help:
	@echo "make: compile packages and dependencies"
	@echo "make tool: run specified go tool"
	@echo "make lint: golint ./..."
	@echo "make clean: remove object files and cached files"