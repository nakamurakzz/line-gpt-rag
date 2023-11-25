FUNCTION_NAME=LineGptRagFunction

build:
	GOARCH=amd64 GOOS=linux go build function/src/main.go

zip:build
	zip main.zip main	

deploy:zip
	aws lambda update-function-code --function-name $(FUNCTION_NAME) --zip-file fileb://main.zip
