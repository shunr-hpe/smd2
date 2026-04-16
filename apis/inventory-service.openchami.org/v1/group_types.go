package v1

import (
	"context"

	"github.com/openchami/fabrica/pkg/fabrica"
)

type Group struct {
	APIVersion string           `json:"apiVersion"`
	Kind       string           `json:"kind"`
	Metadata   fabrica.Metadata `json:"metadata"`
	ID         string           `json:"id,omitempty"`
	Spec       GroupSpec        `json:"spec" validate:"required"`
	Status     GroupStatus      `json:"status,omitempty"`
}

type GroupSpec struct {
	Description    string `json:"description,omitempty" validate:"max=200"`
	Label          string `json:"label"`
	ExclusiveGroup string `json:"exclusiveGroup,omitempty"`

	Tags    []string `json:"tags,omitempty"`
	Members Members  `json:"members"`
}

type GroupStatus struct {
	Phase   string `json:"phase,omitempty"`
	Message string `json:"message,omitempty"`
	Ready   bool   `json:"ready"`
}

func (r *Group) Validate(ctx context.Context) error {

	return nil
}

func (r *Group) GetKind() string {
	return "Group"
}

func (r *Group) GetName() string {
	return r.Metadata.Name
}

func (r *Group) GetUID() string {
	return r.Metadata.UID
}

func (r *Group) IsHub() {}

type Members struct {
	IDs []string `json:"ids"`
}
type Membership struct {
	ID            string   `json:"id"`
	GroupLabels   []string `json:"groupLabels"`
	PartitionName string   `json:"partitionName"`
}
