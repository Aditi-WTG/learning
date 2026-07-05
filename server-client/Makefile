PROTO_DIR := proto
OUT_DIR   := pb

.PHONY: proto server publisher consumer clean

proto:
	protoc \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/store.proto

server:
	go run ./server

publisher:
	go run ./publisher

consumer:
	go run ./consumer

clean:
	rm -f $(OUT_DIR)/*.pb.go $(OUT_DIR)/*_grpc.pb.go