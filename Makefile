grpc-server:
	go run server/main.go

grpc-client:
	go run client/main.go

proto-generate:
	protoc -I ./proto \
	--go_out ./proto --go_opt paths=source_relative \
	--go-grpc_out ./proto --go-grpc_opt paths=source_relative \
	--grpc-gateway_out ./proto --grpc-gateway_opt paths=source_relative \
	./proto/user/user.proto

environment-variable:
	export PATH=$PATH:$(go env GOPATH)/bin