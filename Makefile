.PHONY: default

default:
	rm -fv lambda-handler.zip bootstrap
	GOOS=linux GOARCH=arm64 go build -o bootstrap invalidate.go
	zip -9 lambda-handler.zip bootstrap
	rm -fv bootstrap
