package internal

import (
	"encoding/json"
	"fmt"

	apiresource "buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/commons/apiresource"
	cloudresourcev1 "buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/infrahub/cloudresource/v1"
	cloudresourcekind "buf.build/gen/go/project-planton/apis/protocolbuffers/go/org/project_planton/shared/cloudresourcekind"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

// WrapCloudResource wraps spec data into a CloudResource message.
// This is the reverse operation of UnwrapCloudResource.
//
// The function:
//  1. Takes the CloudResourceKind to determine which resource type to create
//  2. Uses the CloudObject's oneof descriptor to find the correct field
//  3. Converts the spec data (map) to the specific protobuf message using JSON marshaling
//  4. Sets the oneof field in CloudObject
//  5. Wraps everything in a CloudResource with the provided metadata
//
// Args:
//   - kind: The CloudResourceKind enum value
//   - specData: Map containing the resource-specific spec fields
//   - metadata: ApiResourceMetadata for the cloud resource
//
// Returns:
//   - CloudResource wrapper with the spec data properly set
//   - Error if wrapping fails
func WrapCloudResource(
	kind cloudresourcekind.CloudResourceKind,
	specData map[string]interface{},
	metadata *apiresource.ApiResourceMetadata,
) (*cloudresourcev1.CloudResource, error) {
	if kind == cloudresourcekind.CloudResourceKind_unspecified {
		return nil, fmt.Errorf("cloud resource kind is unspecified")
	}

	// Create a CloudObject instance
	cloudObject := &cloudresourcev1.CloudObject{}
	cloudObjectReflect := cloudObject.ProtoReflect()
	cloudObjectDescriptor := cloudObjectReflect.Descriptor()

	// Find the "object" oneof field
	oneofDescriptor := cloudObjectDescriptor.Oneofs().ByName("object")
	if oneofDescriptor == nil {
		return nil, fmt.Errorf("object oneof not found in CloudObject")
	}

	// Find the field descriptor for this kind
	fieldName := kindToFieldName(kind.String())
	var targetField protoreflect.FieldDescriptor
	fields := oneofDescriptor.Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		if string(field.Name()) == fieldName {
			targetField = field
			break
		}
	}

	if targetField == nil {
		return nil, fmt.Errorf("field %s not found in oneof object for kind %s", fieldName, kind.String())
	}

	// Get the message descriptor for this field
	messageDescriptor := targetField.Message()
	if messageDescriptor == nil {
		return nil, fmt.Errorf("field %s is not a message type", fieldName)
	}

	// Create a new dynamic message instance from the descriptor
	resourceMessage := dynamicpb.NewMessage(messageDescriptor)

	// Convert the spec data map to JSON, then unmarshal into the protobuf message
	// This approach handles type conversions automatically
	specJSON, err := json.Marshal(specData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal spec data to JSON: %w", err)
	}

	// Use protojson to unmarshal into the message
	if err := protojson.Unmarshal(specJSON, resourceMessage); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spec data into %s: %w", messageDescriptor.Name(), err)
	}

	// Set the field value in the CloudObject oneof
	cloudObjectReflect.Set(targetField, protoreflect.ValueOfMessage(resourceMessage.ProtoReflect()))

	// Create the CloudResource wrapper
	cloudResource := &cloudresourcev1.CloudResource{
		ApiVersion: "infra-hub.planton.ai/v1",
		Kind:       "CloudResource",
		Metadata:   metadata,
		Spec: &cloudresourcev1.CloudResourceSpec{
			Kind:        kind,
			CloudObject: cloudObject,
		},
	}

	return cloudResource, nil
}

