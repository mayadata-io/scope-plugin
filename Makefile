.PHONY: all

SUDO=$(shell docker info >/dev/null 2>&1 || echo "sudo -E")
EXE=scope-plugin
ORGANIZATION=openebs
IMAGE=$(ORGANIZATION)/$(EXE)
NAME=$(ORGANIZATION)-$(EXE)
UPTODATE=.$(EXE).uptodate

all: build image

build:
	go build -v

image: $(UPTODATE)
$(UPTODATE): $(EXE) Dockerfile
	$(SUDO) docker build -t $(IMAGE):latest .

clean:
	- $(SUDO) docker rmi $(IMAGE)

# By using the `./...` notation, all the non-vendor packages are going
# to be tested if they have test files.
test:
	- go test ./...

lint:
	golint

# By using the `./...` notation, all the non-vendor packages are going
# to be tested if they have test files.
fmt:
	- go fmt ./...
