GOOS := $(shell go env GOOS)
EXE_EXT := $(if $(filter windows,$(GOOS)),.exe,)
SERVER_BIN := isms-server$(EXE_EXT)
CGO_GO := $(if $(filter windows,$(GOOS)),set CGO_ENABLED=1&&,CGO_ENABLED=1)

.PHONY: run build tidy clean

run:
	$(CGO_GO) go run cmd/main.go

build:
	$(CGO_GO) go build -o $(SERVER_BIN) cmd/main.go

tidy:
	go mod tidy

clean:
	rm -f isms-server isms-server-* data/isms.db

docker-build:
	docker build -t isms-privilege .

# 初始化資料目錄
init:
	mkdir -p data logs www/html

s:
	git push -u origin main
