// Package openapi holds the CRD validation schema type that the installer applies
// to the cluster.
//
// JSONSchemaProps embeds apiextensionsv1.JSONSchemaProps and adds the one field the
// stock type cannot hold: the Deckhouse kube-apiserver carries
// 010-x-kubernetes-sensitive-data.patch and runs with CRDSensitiveData=true, so it
// understands x-kubernetes-sensitive-data. Decoding a CRD through the upstream type
// alone would drop that field before it ever reached the cluster.
//
// Only the positions that carry a nested schema are redeclared here, so the extension
// survives at any depth. Every other schema field — its type, its json tag, its
// omitempty — is inherited from upstream and cannot drift.
//
// The type is also the strict contract in the other direction: any key that is not a
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
// Being the allowlist cuts the other way too: a schema field the cluster's apiserver
// understands and this build does not is dropped silently, and TestForkCoversUpstreamFields
// only fires when this module bumps k8s.io/apiextensions-apiserver — never when the cluster
// moves ahead of it. Keep the dependency in step with the apiserver Deckhouse ships. A
// schema this package cannot decode at all is not dropped: the installer sends that
// document as it came and reports the error (see sanitize).
//
// TestForkCoversUpstreamFields and TestForkMarshalsLikeUpstream guard the type against
// drift; read them before bumping k8s.io/apiextensions-apiserver.
package openapi

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// JSONSchemaProps is a JSON-Schema following Specification Draft 4 (http://json-schema.org/).
//
// It is apiextensionsv1.JSONSchemaProps with the nested schema positions retargeted at
// this type and XSensitiveData added.
type JSONSchemaProps struct {
	// every field that holds no nested schema, verbatim from upstream
	apiextensionsv1.JSONSchemaProps `json:",inline"`

	// The positions that carry a nested schema, retargeted at this type: upstream
	// declares them through its own JSONSchemaProps, which cannot hold XSensitiveData.
	//
	// Each one shadows the embedded field of the same json name, and json resolves that
	// by depth: these are the ones serialized, the upstream ones are dropped from the
	// field set entirely and stay nil. Never read a nested schema through the embedded
	// struct. TestForkCoversUpstreamFields is what keeps the list complete.
	Items                *JSONSchemaPropsOrArray    `json:"items,omitempty"`
	AllOf                []JSONSchemaProps          `json:"allOf,omitempty"`
	OneOf                []JSONSchemaProps          `json:"oneOf,omitempty"`
	AnyOf                []JSONSchemaProps          `json:"anyOf,omitempty"`
	Not                  *JSONSchemaProps           `json:"not,omitempty"`
	Properties           map[string]JSONSchemaProps `json:"properties,omitempty"`
	AdditionalProperties *JSONSchemaPropsOrBool     `json:"additionalProperties,omitempty"`
	PatternProperties    map[string]JSONSchemaProps `json:"patternProperties,omitempty"`
	Dependencies         JSONSchemaDependencies     `json:"dependencies,omitempty"`
	AdditionalItems      *JSONSchemaPropsOrBool     `json:"additionalItems,omitempty"`
	Definitions          JSONSchemaDefinitions      `json:"definitions,omitempty"`

	// XSensitiveData marks a field (or an object/array subtree) as sensitive: the
	// apiserver encrypts it in etcd, filters it by RBAC through the <resource>/sensitive
	// subresource, and masks it in audit logs.
	//
	// This is NOT an upstream Kubernetes field. It only exists on the Deckhouse
	// kube-apiserver, and it is the sole reason this package declares its own type.
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
