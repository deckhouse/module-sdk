package objectpatch

type CreateOperation string

const (
	Create            CreateOperation = "Create"
	CreateOrUpdate    CreateOperation = "CreateOrUpdate"
	CreateIfNotExists CreateOperation = "CreateIfNotExists"
)

type DeleteOperation string

const (
	// DeletePropagationForeground
	Delete DeleteOperation = "Delete"
	// DeletePropagationBackground
	DeleteInBackground DeleteOperation = "DeleteInBackground"
	// DeletePropagationOrphan
	DeleteNonCascading DeleteOperation = "DeleteNonCascading"
)

type PatchOperation string

const (
	MergePatch PatchOperation = "MergePatch"
	JQPatch    PatchOperation = "JQPatch"
	JSONPatch  PatchOperation = "JSONPatch"
)
