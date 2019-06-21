
$(GOPATH)/bin/dep:
	go get -u github.com/golang/dep/cmd/dep

dep: $(GOPATH)/bin/dep
	$(GOPATH)/bin/dep ensure -vendor-only $(verbose)

rebuild-dependencies:
	dep ensure -v -no-vendor

install: dep go-install

go-install:
	go install -v .
