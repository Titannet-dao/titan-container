package kube

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type resourceItem struct {
	Capacity    resource.Quantity
	Allocatable resource.Quantity
	Allocated   resource.Quantity
}

func newResourceItem(capacity, allocatable, allocated resource.Quantity) resourceItem {
	rp := resourceItem{
		Capacity:    capacity,
		Allocatable: allocatable,
		Allocated:   allocated,
	}

	return rp
}

type nodeResource struct {
	CPU              resourceItem
	Memory           resourceItem
	EphemeralStorage resourceItem
}

func newNodeResource(nodeStatus *corev1.NodeStatus) *nodeResource {
	mzero := resource.NewMilliQuantity(0, resource.DecimalSI)
	zero := resource.NewQuantity(0, resource.DecimalSI)

	capacity := nodeStatus.Capacity
	allocatable := nodeStatus.Allocatable

	nr := &nodeResource{
		CPU:              newResourceItem(capacity.Cpu().DeepCopy(), allocatable.Cpu().DeepCopy(), mzero.DeepCopy()),
		Memory:           newResourceItem(capacity.Memory().DeepCopy(), allocatable.Memory().DeepCopy(), zero.DeepCopy()),
		EphemeralStorage: newResourceItem(capacity.StorageEphemeral().DeepCopy(), allocatable.StorageEphemeral().DeepCopy(), zero.DeepCopy()),
	}

	return nr
}

func (nr *nodeResource) addAllocatedResources(rl corev1.ResourceList) {
	for name, quantity := range rl {
		switch name {
		case corev1.ResourceCPU:
			nr.CPU.Allocated.Add(quantity)
		case corev1.ResourceMemory:
			nr.Memory.Allocated.Add(quantity)
		case corev1.ResourceEphemeralStorage:
			nr.EphemeralStorage.Allocated.Add(quantity)
		}
	}
}
