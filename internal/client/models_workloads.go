package client

// Workload represents a FlashBlade workload from GET responses.
// Workloads organise storage resources (volumes, file systems, etc.)
// and their related configuration objects into logical groupings.
type Workload struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Context       *NamedReference `json:"context,omitempty"`
	Created       int64           `json:"created"`
	Destroyed     bool            `json:"destroyed"`
	Preset        *WorkloadPreset `json:"preset,omitempty"`
	Status        string          `json:"status"`
	StatusDetails []string        `json:"status_details"`
	TimeRemaining int64           `json:"time_remaining"`
}

// WorkloadPreset is a reference to the preset from which the workload was deployed.
// It extends NamedReference with the revision number of the deployed preset.
type WorkloadPreset struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Revision int64  `json:"revision,omitempty"`
}

// WorkloadPost contains the fields for POST /workloads.
// Name is sent via ?names= and preset via ?preset_names= query parameters.
type WorkloadPost struct {
	Parameters []WorkloadParameter `json:"parameters,omitempty"`
}

// WorkloadParameter represents a single named parameter value passed to a preset
// when creating a workload.
type WorkloadParameter struct {
	Name  string                 `json:"name"`
	Value WorkloadParameterValue `json:"value"`
}

// WorkloadParameterValue holds the typed value for a workload parameter.
// Only one field should be set per parameter; the others remain zero/nil.
type WorkloadParameterValue struct {
	Boolean           *bool                             `json:"boolean,omitempty"`
	Integer           *int64                            `json:"integer,omitempty"`
	String            *string                           `json:"string,omitempty"`
	ResourceReference *WorkloadParameterResourceRef     `json:"resource_reference,omitempty"`
}

// WorkloadParameterResourceRef is the value for a resource-reference parameter.
type WorkloadParameterResourceRef struct {
	ID           string `json:"id,omitempty"`
	Name         string `json:"name,omitempty"`
	ResourceType string `json:"resource_type,omitempty"`
}

// WorkloadPatch contains pointer fields for PATCH /workloads.
// nil = omit the field; non-nil = send the field.
type WorkloadPatch struct {
	Destroyed *bool   `json:"destroyed,omitempty"`
	Name      *string `json:"name,omitempty"`
}
