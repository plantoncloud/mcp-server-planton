package internal

import (
	cloudresourcev1 "buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/infrahub/cloudresource/v1"
	cloudresourcekind "buf.build/gen/go/project-planton/apis/protocolbuffers/go/org/project_planton/shared/cloudresourcekind"
	"fmt"
	"google.golang.org/protobuf/proto"
)

// UnwrapCloudResource extracts the specific cloud resource object from a CloudResource wrapper.
//
// The CloudResource wrapper is an internal implementation detail used by the backend to unify
// all cloud resource types under a single API. Client-facing code (CLI, web console, MCP server)
// should only expose the specific resource type (e.g., AwsEksCluster, GcpGkeCluster).
//
// This function uses protobuf reflection to dynamically extract the correct resource type
// from the CloudObject's oneof field based on the CloudResourceKind.
//
// Args:
//   - cloudResource: The CloudResource wrapper object returned by the gRPC API
//
// Returns:
//   - The specific cloud resource object (e.g., AwsEksCluster, GcpGkeCluster) as proto.Message
//   - Error if the resource cannot be unwrapped
func UnwrapCloudResource(cloudResource *cloudresourcev1.CloudResource) (proto.Message, error) {
	if cloudResource == nil || cloudResource.Spec == nil {
		return nil, fmt.Errorf("cloud resource or spec is nil")
	}

	cloudResourceKind := cloudResource.Spec.Kind
	if cloudResourceKind == cloudresourcekind.CloudResourceKind_unspecified {
		return nil, fmt.Errorf("cloud resource kind is unspecified")
	}

	cloudObject := cloudResource.Spec.CloudObject
	if cloudObject == nil {
		return nil, fmt.Errorf("cloud object is nil in CloudResource spec")
	}

	// Get the protobuf reflection descriptor for cloudObject
	cloudObjectReflect := cloudObject.ProtoReflect()
	cloudObjectDescriptor := cloudObjectReflect.Descriptor()

	// Find the "object" oneof field
	oneofDescriptor := cloudObjectDescriptor.Oneofs().ByName("object")
	if oneofDescriptor == nil {
		return nil, fmt.Errorf("object oneof not found in CloudObject")
	}

	// Get which field is set in the oneof
	whichOneof := cloudObjectReflect.WhichOneof(oneofDescriptor)
	if whichOneof == nil {
		return nil, fmt.Errorf("no field is set in the object oneof")
	}

	// Get the value of the set field
	fieldValue := cloudObjectReflect.Get(whichOneof)
	if !fieldValue.IsValid() {
		return nil, fmt.Errorf("oneof field value is invalid")
	}

	// Extract the message from the field value
	messageValue := fieldValue.Message()
	if messageValue == nil {
		return nil, fmt.Errorf("oneof field does not contain a message")
	}

	// Convert to proto.Message interface
	unwrappedResource := messageValue.Interface()

	// Note: Unlike the CLI, we don't need to convert ApiResourceMetadata to CloudResourceMetadata
	// because the underlying cloud resource objects already have their own metadata field that
	// was populated by the backend. The CloudResource.Metadata is the wrapper's metadata.
	// The actual resource's metadata is already part of the unwrapped object.

	return unwrappedResource, nil
}
