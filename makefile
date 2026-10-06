GOOS := $(shell go env GOOS)
EXE_EXT := $(if $(filter windows,$(GOOS)),.exe,)
SERVER_BIN := isms-server-$(GOOS)$(EXE_EXT)

.PHONY: run build tidy clean

run:
	go run cmd/main.go

build:
	go build -o $(SERVER_BIN) cmd/main.go

tidy:
	go mod tidy

clean:
	rm -f isms-server isms-server-* data/isms.db

docker-build:
	docker build -t isms-privilege .

# 初始化資料目錄
init:
	mkdir -p data logs www/html
