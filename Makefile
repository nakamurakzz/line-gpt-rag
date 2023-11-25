build:
	GOARCH=amd64 GOOS=linux go build function/src/main.go

zip:build
	zip main.zip main	