all:    install

install:
	go install
	go install ./proto

test:
	go test

clean:
	go clean ./...

nuke:
	go clean -i ./...

regenerate:
	make -C proto regenerate
