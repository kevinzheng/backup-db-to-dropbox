all: build

test:
	go test -v -race -cover .

build:
	env GOOS=linux GOARCH=amd64 go build -o build/backup-db-to-dropbox ./*.go

clean:
	@rm -rf build