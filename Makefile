PROTO_DIR=proto
GO_OUT=server/GORPC/proto
PY_OUT=server/PyRPC/proto

ENV_FILE=".env"
PROTO_FILES=$(PROTO_DIR)/*.proto

.PHONY: all  clean compile server

all: build compile  

compile : _go_proto _py_proto

user : 
	uv run client/main.py user 

server : 
	cd server ; uv run main.py

# Compiling for Go
# TODO :: Need to modify to regex matching dependecy ( forgot how to do that need to  see) 
_go_proto:
	protoc \
		--go_out=$(GO_OUT) \
		--go_opt=paths=source_relative\
		--go-grpc_out=$(GO_OUT) --go-grpc_opt=paths=source_relative \
		-I proto/ \
		$(PROTO_FILES)

# Compiling For Python
# TODO :: Need to modify to regex matching dependecy ( forgot how to do that need to  see) 
_py_proto:
# example [for reference]
# python -m grpc_tools.protoc -I../../protos --python_out=. --pyi_out=. --grpc_python_out=. ../../protos/helloworld.proto
	uv run -m grpc_tools.protoc \
		-I proto/ \
		--python_out=$(PY_OUT) \
		--pyi_out=$(PY_OUT) \
		--grpc_python_out=$(PY_OUT) \
		$(PROTO_FILES)

# 	touch $(PY_OUT)/__init__.py

clean:
	rm -rf $(PY_OUT)/*_pb2.py
	rm -rf $(PY_OUT)/*_pb2_grpc.py
	rm -rf $(GO_OUT)/*.pb.go
	rm -f $(PY_OUT)/__init__.py


build: 

	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

	go get google.golang.org/grpc
	go get google.golang.org/protobuf

	pip install uv 
	uv add grpcio grpcio-tools

	touch $(ENV_FILE)





