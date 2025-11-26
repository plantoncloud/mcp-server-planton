package cloudresource

import (
	"fmt"
	"strings"
	"unicode"

	cloudresourcev1 "buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/infrahub/cloudresource/v1"
	cloudresourcekind "buf.build/gen/go/project-planton/apis/protocolbuffers/go/org/project_planton/shared/cloudresourcekind"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// CloudResourceSchema represents the extracted schema for a cloud resource type
type CloudResourceSchema struct {
	Kind        string                 `json:"kind"`
	Description string                 `json:"description,omitempty"`
	Fields      []SchemaField          `json:"fields"`
	Examples    map[string]interface{} `json:"examples,omitempty"`
}

// SchemaField represents a single field in the schema
type SchemaField struct {
	Name         string        `json:"name"`
	Type         string        `json:"type"`
	Required     bool          `json:"required"`
	Description  string        `json:"description,omitempty"`
	Validation   *Validation   `json:"validation,omitempty"`
	EnumValues   []string      `json:"enum_values,omitempty"`
	NestedFields []SchemaField `json:"nested_fields,omitempty"`
	IsRepeated   bool          `json:"is_repeated,omitempty"`
	IsMap        bool          `json:"is_map,omitempty"`
	MapKeyType   string        `json:"map_key_type,omitempty"`
	MapValueType string        `json:"map_value_type,omitempty"`
}

// Validation represents validation rules for a field
type Validation struct {
	MinLength *int64   `json:"min_length,omitempty"`
	MaxLength *int64   `json:"max_length,omitempty"`
	Min       *float64 `json:"min,omitempty"`
	Max       *float64 `json:"max,omitempty"`
	Pattern   string   `json:"pattern,omitempty"`
}

// ExtractCloudResourceSchema extracts the schema for a given CloudResourceKind using protobuf reflection.
// This function uses the CloudObject's oneof descriptor to find the message descriptor for the given kind,
// then inspects its fields to build a comprehensive schema that agents can use to understand required inputs.
func ExtractCloudResourceSchema(kind cloudresourcekind.CloudResourceKind) (*CloudResourceSchema, error) {
	if kind == cloudresourcekind.CloudResourceKind_unspecified {
		return nil, fmt.Errorf("cloud resource kind is unspecified")
	}

	// Get the CloudObject message descriptor
	cloudObject := &cloudresourcev1.CloudObject{}
	cloudObjectReflect := cloudObject.ProtoReflect()
	cloudObjectDescriptor := cloudObjectReflect.Descriptor()

	// Find the "object" oneof field
	oneofDescriptor := cloudObjectDescriptor.Oneofs().ByName("object")
	if oneofDescriptor == nil {
		return nil, fmt.Errorf("object oneof not found in CloudObject")
	}

	// Find the field descriptor for this kind by matching the field name
	// Field names in the oneof follow a pattern based on the kind name
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

	// Extract the kind metadata from the enum value descriptor
	description := extractKindDescription(kind)

	// Extract fields from the message descriptor
	schemaFields, err := extractFieldsFromDescriptor(messageDescriptor)
	if err != nil {
		return nil, fmt.Errorf("failed to extract fields: %w", err)
	}

	return &CloudResourceSchema{
		Kind:        kind.String(),
		Description: description,
		Fields:      schemaFields,
	}, nil
}

// kindToFieldName converts a CloudResourceKind string to the protobuf field name format
// E.g., "aws_rds_instance" stays "aws_rds_instance" (already snake_case)
// "AwsRdsInstance" converts to "aws_rds_instance" (PascalCase to snake_case)
func kindToFieldName(kindStr string) string {
	// If already snake_case (from agent), return as-is
	if strings.Contains(kindStr, "_") {
		return strings.ToLower(kindStr)
	}
	// If PascalCase (from enum key), convert to snake_case
	return pascalToSnakeCase(kindStr)
}

// pascalToSnakeCase converts PascalCase to snake_case
// Examples: "AwsRdsInstance" → "aws_rds_instance", "GcpGkeCluster" → "gcp_gke_cluster"
func pascalToSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

// extractKindDescription extracts the description from CloudResourceKind enum value options
func extractKindDescription(kind cloudresourcekind.CloudResourceKind) string {
	// Get the descriptor for the enum value
	enumValueDescriptor := kind.Descriptor().Values().ByNumber(kind.Number())
	if enumValueDescriptor == nil {
		return ""
	}

	// For now, return the enum name as description
	// Full implementation would extract from proto options if available
	return fmt.Sprintf("Cloud resource of type %s", kind.String())
}

// extractFieldsFromDescriptor extracts fields from a message descriptor
func extractFieldsFromDescriptor(descriptor protoreflect.MessageDescriptor) ([]SchemaField, error) {
	fields := descriptor.Fields()
	var schemaFields []SchemaField

	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)

		// Skip internal/system fields
		if shouldSkipField(field) {
			continue
		}

		schemaField := extractFieldInfo(field)
		schemaFields = append(schemaFields, schemaField)
	}

	return schemaFields, nil
}

// extractFieldInfo extracts detailed information for a single field
func extractFieldInfo(field protoreflect.FieldDescriptor) SchemaField {
	schemaField := SchemaField{
		Name:       string(field.Name()),
		Type:       getFieldType(field),
		Required:   isFieldRequired(field),
		IsRepeated: field.Cardinality() == protoreflect.Repeated && !field.IsMap(),
		IsMap:      field.IsMap(),
	}

	// Extract field description from comments
	schemaField.Description = extractFieldDescription(field)

	// Handle enum fields
	if field.Enum() != nil {
		schemaField.EnumValues = extractEnumValues(field.Enum())
	}

	// Handle nested message fields
	if field.Message() != nil && !field.IsMap() {
		// Extract nested fields from the message descriptor
		nestedFields, _ := extractFieldsFromDescriptor(field.Message())
		schemaField.NestedFields = nestedFields
	}

	// Handle map fields
	if field.IsMap() {
		schemaField.MapKeyType = getFieldType(field.MapKey())
		schemaField.MapValueType = getFieldType(field.MapValue())
	}

	// Extract validation rules (basic support)
	schemaField.Validation = extractValidationRules(field)

	return schemaField
}

// getFieldType returns the string representation of a field's type
func getFieldType(field protoreflect.FieldDescriptor) string {
	if field.IsMap() {
		return "map"
	}

	kind := field.Kind()
	switch kind {
	case protoreflect.StringKind:
		return "string"
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "int32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "int64"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint32"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "uint64"
	case protoreflect.FloatKind:
		return "float"
	case protoreflect.DoubleKind:
		return "double"
	case protoreflect.BoolKind:
		return "bool"
	case protoreflect.BytesKind:
		return "bytes"
	case protoreflect.EnumKind:
		if field.Enum() != nil {
			return string(field.Enum().Name())
		}
		return "enum"
	case protoreflect.MessageKind:
		if field.Message() != nil {
			return string(field.Message().Name())
		}
		return "message"
	default:
		return "unknown"
	}
}

// extractEnumValues extracts all possible values for an enum field
func extractEnumValues(enum protoreflect.EnumDescriptor) []string {
	values := enum.Values()
	result := make([]string, 0, values.Len())

	for i := 0; i < values.Len(); i++ {
		val := values.Get(i)
		// Skip unspecified/default values
		if strings.HasSuffix(string(val.Name()), "unspecified") {
			continue
		}
		result = append(result, string(val.Name()))
	}

	return result
}

// isFieldRequired checks if a field is required (basic heuristic)
// Note: Full buf.validate support would require parsing field options
func isFieldRequired(field protoreflect.FieldDescriptor) bool {
	// Fields marked as required in proto2
	if field.HasOptionalKeyword() {
		return false
	}

	// In proto3, all scalar fields are technically optional (have default values)
	// We'd need to parse buf.validate.field options to determine true required status
	// For now, treat non-repeated, non-optional fields as potentially required
	return field.Cardinality() != protoreflect.Repeated && !field.HasOptionalKeyword()
}

// extractFieldDescription extracts the description from field comments
// Note: This requires the proto files to be compiled with source code info
func extractFieldDescription(field protoreflect.FieldDescriptor) string {
	// protoreflect doesn't expose comments directly in a simple way
	// For now, return empty string
	// Full implementation would require accessing SourceCodeInfo from descriptor proto
	return ""
}

// extractValidationRules extracts validation rules from field options
// Note: Full implementation would parse buf.validate.field extension
func extractValidationRules(field protoreflect.FieldDescriptor) *Validation {
	// Placeholder - full implementation would parse buf.validate options
	// This would require importing buf.validate proto extensions and using proto.GetExtension
	return nil
}

// shouldSkipField determines if a field should be skipped in schema extraction
func shouldSkipField(field protoreflect.FieldDescriptor) bool {
	fieldName := string(field.Name())

	// Skip common system/internal fields
	skipFields := []string{
		"api_version",
		"kind",
		"metadata",
		"status",
		"lifecycle",
		"audit",
	}

	for _, skip := range skipFields {
		if fieldName == skip {
			return true
		}
	}

	return false
}
