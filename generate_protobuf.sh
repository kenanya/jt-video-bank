#go get -u github.com/amsokol/protoc-gen-gotag;go get -u github.com/golang/protobuf/protoc-gen-go
protoc --proto_path=api/proto/v1 --proto_path=third_party --go_out=plugins=grpc:pkg/api/v1 video_bank_service.proto;
protoc --proto_path=api/proto/v1 --proto_path=third_party --gotag_out=xxx="bson+\"-\"",output_path=pkg/api/v1:. video_bank_service.proto
