all: built-in built-out built-check

built-in: in/main.go
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o built-in in/main.go

built-out: out/main.go
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o built-out out/main.go

built-check: check/main.go
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o built-check check/main.go
