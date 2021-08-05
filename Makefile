.PHONY: vendor
vendor:
	go mod vendor
	go mod tidy

.PHONY: test
test:
	go test ./...

.PHONY: bench
bench:
	go test -bench=. ./...
