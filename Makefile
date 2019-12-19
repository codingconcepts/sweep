.PHONY: cover

test:
	go test ./... -v -timeout 10s

cover:
	go test -cover -timeout 10s

cover_html:
	go test -timeout 2s --coverprofile=coverage.out &\
    go tool cover --html=coverage.out