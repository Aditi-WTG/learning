PROTO_DIR := proto
OUT_DIR   := pb

.PHONY: proto build server order duplicate reports clean

proto:
	protoc \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/store.proto

build:
	go build ./...

server:
	go run ./cmd/server

reports:
	go run ./cmd/client -mode=sub -topic=report.daily

order:
	go run ./cmd/client -mode=pub -topic=order.created -message='{"orderId":"O1001","customerId":"C001","items":[{"itemId":"I001","quantity":2}],"destination":"Mumbai","date":"2026-07-06"}'

duplicate:
	go run ./cmd/client -mode=pub -topic=order.created -message='{"orderId":"O1001","customerId":"C001","items":[{"itemId":"I001","quantity":2}],"destination":"Mumbai","date":"2026-07-06"}'

clean:
	rm -f $(OUT_DIR)/*.pb.go $(OUT_DIR)/*_grpc.pb.go