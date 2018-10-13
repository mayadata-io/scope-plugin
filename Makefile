.PHONY: all

SUDO=$(shell docker info >/dev/null 2>&1 || echo "sudo -E")
EXE=scope-plugin
ORGANIZATION=openebs
IMAGE=$(ORGANIZATION)/scope-$(EXE)
NAME=$(ORGANIZATION)-scope-$(EXE)
UPTODATE=.$(EXE).uptodate

all: run build 

run:
	go build -v

build: $(UPTODATE)
$(UPTODATE): $(EXE) Dockerfile
	$(SUDO) docker build -t $(IMAGE):latest .

clean:
	- $(SUDO) docker rmi $(IMAGE)

test:
	- go test 
lint:
	golint
fmt:
	- go fmt 
