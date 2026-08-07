.PHONY: run build tidy clean

run:
	go run cmd/main.go

build:
	go build -o isms-server cmd/main.go

tidy:
	go mod tidy

clean:
	rm -f isms-server data/isms.db

docker-build:
	docker build -t isms-privilege .

# 初始化資料目錄
init:
	mkdir -p data logs www/html
