all: fmt install

install:
	go install

fmt:
	goimports -w ./..	
	gofmt -s -w ./.. 