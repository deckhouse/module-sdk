package objectpatch

type CreateOperation string

const (
	Create            CreateOperation = "Create"
	CreateOrUpdate    CreateOperation = "CreateOrUpdate"
	CreateIfNotExists CreateOperation = "CreateIfNotExists"
)

type DeleteOperation string

const (
	Delete             DeleteOperation = "Delete"
	DeleteInBackground DeleteOperation = "DeleteInBackground"
	DeleteNonCascading DeleteOperation = "DeleteNonCascading"
)

type PatchOperation string

const (
	MergePatch PatchOperation = "MergePatch"
	JQPatch    PatchOperation = "JQPatch"
	JSONPatch  PatchOperation = "JSONPatch"
)
