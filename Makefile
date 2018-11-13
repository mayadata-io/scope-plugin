.PHONY: all

SUDO=$(shell docker info >/dev/null 2>&1 || echo "sudo -E")
EXE=scope-plugin
ORGANIZATION=openebs
IMAGE=$(ORGANIZATION)/$(EXE)
NAME=$(ORGANIZATION)-$(EXE)
UPTODATE=.$(EXE).uptodate

all: build image

build:
	go build -o $(EXE)

image: $(UPTODATE)
$(UPTODATE): $(EXE) Dockerfile
	$(SUDO) docker build -t $(IMAGE):latest .
	$(SUDO) docker save $(IMAGE):latest > plugin.tar

clean:
	- $(SUDO) docker rmi $(IMAGE)

# By using the `./...` notation, all the non-vendor packages are going
# to be tested if they have test files.
test:
	$(PWD)/hack/coverage.sh

lint:
	golint

# By using the `./...` notation, all the non-vendor packages are going
# to be tested if they have test files.
fmt:
	- go fmt ./...
