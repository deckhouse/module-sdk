// Package openapi holds the CRD validation schema type that the installer applies
// to the cluster.
//
// It is a fork of apiextensionsv1.JSONSchemaProps. The fork exists for exactly one
// reason: the Deckhouse kube-apiserver carries 010-x-kubernetes-sensitive-data.patch
// and runs with CRDSensitiveData=true, so it understands one schema field that the
// stock apiextensions-apiserver Go types do not. Decoding a CRD through the upstream
// type would drop that field before it ever reached the cluster.
//
// The fork is also the strict contract in the other direction: any key that is not a
// field here — x-doc-examples, x-examples, x-description, x-kubernetes-immutable and
// friends — is dropped on decode instead of being sent to the apiserver, which would
// prune it anyway and log an "unknown field" warning for every occurrence.
//
// The contract holds only on the Deckhouse kube-apiserver. On a stock one — a managed
// control plane, a dev cluster, CRDSensitiveData off — x-kubernetes-sensitive-data is
// pruned server-side, so a CRD that carries it is updated on every reconcile. That is
// accepted: the field is meaningless there anyway, and detecting it would mean probing
// the apiserver build for every install.
//
// TestForkCoversUpstreamFields guards the copy against drift; read it before bumping
// k8s.io/apiextensions-apiserver.
package openapi

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// JSONSchemaProps is a JSON-Schema following Specification Draft 4 (http://json-schema.org/).
//
// Field-for-field mirror of apiextensionsv1.JSONSchemaProps, with the nested schema
// positions retargeted at this package and XSensitiveData added.
type JSONSchemaProps struct {
	ID          string                        `json:"id,omitempty"`
	Schema      apiextensionsv1.JSONSchemaURL `json:"$schema,omitempty"`
	Ref         *string                       `json:"$ref,omitempty"`
	Description string                        `json:"description,omitempty"`
	Type        string                        `json:"type,omitempty"`
	Format      string                        `json:"format,omitempty"`

	Title            string                `json:"title,omitempty"`
	Default          *apiextensionsv1.JSON `json:"default,omitempty"`
	Maximum          *float64              `json:"maximum,omitempty"`
	ExclusiveMaximum bool                  `json:"exclusiveMaximum,omitempty"`
	Minimum          *float64              `json:"minimum,omitempty"`
	ExclusiveMinimum bool                  `json:"exclusiveMinimum,omitempty"`
	MaxLength        *int64                `json:"maxLength,omitempty"`
	MinLength        *int64                `json:"minLength,omitempty"`
	Pattern          string                `json:"pattern,omitempty"`
	MaxItems         *int64                `json:"maxItems,omitempty"`
	MinItems         *int64                `json:"minItems,omitempty"`
	UniqueItems      bool                  `json:"uniqueItems,omitempty"`
	MultipleOf       *float64              `json:"multipleOf,omitempty"`

	Enum          []apiextensionsv1.JSON `json:"enum,omitempty"`
	MaxProperties *int64                 `json:"maxProperties,omitempty"`
	MinProperties *int64                 `json:"minProperties,omitempty"`

	Required []string                `json:"required,omitempty"`
	Items    *JSONSchemaPropsOrArray `json:"items,omitempty"`

	AllOf                []JSONSchemaProps                      `json:"allOf,omitempty"`
	OneOf                []JSONSchemaProps                      `json:"oneOf,omitempty"`
	AnyOf                []JSONSchemaProps                      `json:"anyOf,omitempty"`
	Not                  *JSONSchemaProps                       `json:"not,omitempty"`
	Properties           map[string]JSONSchemaProps             `json:"properties,omitempty"`
	AdditionalProperties *JSONSchemaPropsOrBool                 `json:"additionalProperties,omitempty"`
	PatternProperties    map[string]JSONSchemaProps             `json:"patternProperties,omitempty"`
	Dependencies         JSONSchemaDependencies                 `json:"dependencies,omitempty"`
	AdditionalItems      *JSONSchemaPropsOrBool                 `json:"additionalItems,omitempty"`
	Definitions          JSONSchemaDefinitions                  `json:"definitions,omitempty"`
	ExternalDocs         *apiextensionsv1.ExternalDocumentation `json:"externalDocs,omitempty"`
	Example              *apiextensionsv1.JSON                  `json:"example,omitempty"`
	Nullable             bool                                   `json:"nullable,omitempty"`

	XPreserveUnknownFields *bool                           `json:"x-kubernetes-preserve-unknown-fields,omitempty"`
	XEmbeddedResource      bool                            `json:"x-kubernetes-embedded-resource,omitempty"`
	XIntOrString           bool                            `json:"x-kubernetes-int-or-string,omitempty"`
	XListMapKeys           []string                        `json:"x-kubernetes-list-map-keys,omitempty"`
	XListType              *string                         `json:"x-kubernetes-list-type,omitempty"`
	XMapType               *string                         `json:"x-kubernetes-map-type,omitempty"`
	XValidations           apiextensionsv1.ValidationRules `json:"x-kubernetes-validations,omitempty"`

	// XSensitiveData marks a field (or an object/array subtree) as sensitive: the
	// apiserver encrypts it in etcd, filters it by RBAC through the <resource>/sensitive
	// subresource, and masks it in audit logs.
	//
	// This is NOT an upstream Kubernetes field. It only exists on the Deckhouse
	// kube-apiserver, and it is the sole reason this package forks JSONSchemaProps.
	// Keep it listed in forkOnlyFields in the test when adding others.
	XSensitiveData bool `json:"x-kubernetes-sensitive-data,omitempty"`
}

// JSONSchemaPropsOrArray represents a value that can either be a JSONSchemaProps
// or an array of JSONSchemaProps. Mainly here for serialization purposes.
type JSONSchemaPropsOrArray struct {
	Schema      *JSONSchemaProps  `json:"-"`
	JSONSchemas []JSONSchemaProps `json:"-"`
}

// JSONSchemaPropsOrBool represents JSONSchemaProps or a boolean value.
// Defaults to true for the boolean property.
type JSONSchemaPropsOrBool struct {
	Allows bool             `json:"-"`
	Schema *JSONSchemaProps `json:"-"`
}

// JSONSchemaPropsOrStringArray represents a JSONSchemaProps or a string array.
type JSONSchemaPropsOrStringArray struct {
	Schema   *JSONSchemaProps `json:"-"`
	Property []string         `json:"-"`
}

// JSONSchemaDependencies represent a dependencies property.
type JSONSchemaDependencies map[string]JSONSchemaPropsOrStringArray

// JSONSchemaDefinitions contains the models explicitly defined in this spec.
type JSONSchemaDefinitions map[string]JSONSchemaProps
