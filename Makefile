VER=v1
SERVICE=?

generate-proto:
	protoc --proto_path=./protos/ \
		--go_out=./$(SERVICE)-service/api --go_opt=paths=source_relative \
		--go-grpc_out=./$(SERVICE)-service/api --go-grpc_opt=paths=source_relative \
		./protos/$(SERVICE)/$(VER)/$(SERVICE).proto


generate-proto-gateway:
	protoc --proto_path=./protos/ \
		--go_out=./rest-gateway/api --go_opt=paths=source_relative \
		--go-grpc_out=./rest-gateway/api --go-grpc_opt=paths=source_relative \
		./protos/$(SERVICE)/$(VER)/$(SERVICE).proto