module github.com/plantoncloud-inc/mcp-server-planton

go 1.24.7

require (
	github.com/mark3labs/mcp-go v0.6.0
	github.com/plantoncloud-inc/planton-cloud/apis v0.0.0
	google.golang.org/grpc v1.75.0
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.9-20250912141014-52f32327d4b0.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	golang.org/x/net v0.45.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250825161204-c5933d9347a5 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

replace github.com/plantoncloud-inc/planton-cloud/apis => ../planton-cloud/apis
