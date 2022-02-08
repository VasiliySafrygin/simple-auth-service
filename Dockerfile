FROM golang
RUN mkdir /code/
ADD . /code/
WORKDIR /code/
RUN apt-get update && apt-get install -y protobuf-compiler
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1
ENV PATH="$PATH:$(go env GOPATH)/bin"
RUN protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./rpcpb/grpc.proto
RUN go mod download
RUN go build main.go
EXPOSE 8080
ENTRYPOINT ["./main"]

