run = no_run
test:
ifeq ($(run), no_run)
	go test -v ./...
else
	go test -v -run $(run) ./...
endif

test-cover:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html