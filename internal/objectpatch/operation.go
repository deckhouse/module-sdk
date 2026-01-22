package objectpatch

// CreateOperation defines object creation strategies.
type CreateOperation string

const (
	Create            CreateOperation = "Create"            // Always create (fails if exists)
	CreateOrUpdate    CreateOperation = "CreateOrUpdate"    // Create or update if exists
	CreateIfNotExists CreateOperation = "CreateIfNotExists" // Create only if not exists
)

// DeleteOperation defines object deletion propagation policies.
type DeleteOperation string

const (
	// Delete uses foreground propagation: waits for dependents to be deleted first.
	Delete DeleteOperation = "Delete"
	// DeleteInBackground uses background propagation: deletes object immediately,
	// garbage collector handles dependents asynchronously.
	DeleteInBackground DeleteOperation = "DeleteInBackground"
	// DeleteNonCascading orphans dependents: removes owner references without deletion.
	DeleteNonCascading DeleteOperation = "DeleteNonCascading"
)

// PatchOperation defines how patches are applied to objects.
type PatchOperation string

const (
	MergePatch PatchOperation = "MergePatch" // RFC7396 JSON Merge Patch
	JQPatch    PatchOperation = "JQPatch"    // Mutate object with jq expression
	JSONPatch  PatchOperation = "JSONPatch"  // RFC6902 JSON Patch (op/path/value)
)
