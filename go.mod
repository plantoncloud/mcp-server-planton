module github.com/plantoncloud-inc/mcp-server-planton

go 1.23

require (
	buf.build/gen/go/blintora/apis/grpc/go v1.5.1-20251125011413-52ef5c4f2840.2
	buf.build/gen/go/blintora/apis/protocolbuffers/go v1.36.10-20251125011413-52ef5c4f2840.1
	buf.build/gen/go/project-planton/apis/protocolbuffers/go v1.36.10-20251124125039-9c224fb3651e.1
	github.com/mark3labs/mcp-go v0.6.0
	github.com/project-planton/project-planton v0.2.245
	google.golang.org/grpc v1.75.0
	google.golang.org/protobuf v1.36.10
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.10-20250912141014-52f32327d4b0.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	golang.org/x/net v0.45.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250825161204-c5933d9347a5 // indirect
)
