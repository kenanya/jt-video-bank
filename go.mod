module github.com/kenanya/jt-video-bank

require (
	github.com/amsokol/mongo-go-driver-protobuf v1.0.0-rc5
	github.com/amsokol/protoc-gen-gotag v0.2.1
	github.com/fatih/structtag v1.0.0 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/golang/protobuf v1.3.1
	github.com/golang/snappy v0.0.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/lyft/protoc-gen-star v0.4.11 // indirect
	github.com/spf13/afero v1.2.2 // indirect
	github.com/xdg/scram v0.0.0-20180814205039-7eeb5667e42c // indirect
	github.com/xdg/stringprep v1.0.0 // indirect
	go.mongodb.org/mongo-driver v1.0.0-rc1
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.10.0
	google.golang.org/grpc v1.21.0
	gopkg.in/yaml.v2 v2.2.2
)

replace github.com/kenanya/jt-video-bank => ../jt-video-bank

go 1.13
