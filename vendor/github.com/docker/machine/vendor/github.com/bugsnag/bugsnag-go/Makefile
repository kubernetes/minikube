TEST?=./...

default: alldeps test

deps:
	go get -v -d ./...

alldeps:
	go get -v -d -t ./...

updatedeps:
	go get -v -d -u ./...

test: alldeps
	go test
	@go vet 2>/dev/null ; if [ $$? -eq 3 ]; then \
		go get golang.org/x/tools/cmd/vet; \
	fi
	@go vet $(TEST) ; if [ $$? -eq 1 ]; then \
		echo "go-vet: Issues running go vet ./..."; \
		exit 1; \
	fi

ci: alldeps test

bench:
	go test --bench=.*


.PHONY: bin checkversion ci default deps generate releasebin test testacc testrace updatedeps
