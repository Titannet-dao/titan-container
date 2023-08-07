package builder

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/Filecoin-Titan/titan-container/node/impl/provider/kube/manifest"
	sdk "github.com/cosmos/cosmos-sdk/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	StorageAttributePersistent = "persistent"
	StorageClassDefault        = "default"
)

type workloadBase interface {
	builderBase
	Name() string
}

type Workload struct {
	builder
	serviceIdx int
}

var _ workloadBase = (*Workload)(nil)

func NewWorkload(settings Settings, deployment IClusterDeployment, serviceIdx int) Workload {
	return Workload{
		builder: builder{
			settings:   settings,
			deployment: deployment,
		},
		serviceIdx: serviceIdx,
	}
}

func (b *Workload) Name() string {
	return b.deployment.ManifestGroup().Services[b.serviceIdx].Name
}

func (b *Workload) replicas() *int32 {
	replicas := new(int32)
	*replicas = int32(b.deployment.ManifestGroup().Services[b.serviceIdx].Count)

	return replicas
}

func (b *Workload) labels() map[string]string {
	obj := b.builder.labels()
	obj[TitanManifestServiceLabelName] = b.deployment.ManifestGroup().Services[b.serviceIdx].Name
	return obj
}

func (b *Workload) imagePullSecrets() []corev1.LocalObjectReference {
	if b.settings.DockerImagePullSecretsName == "" {
		return nil
	}

	return []corev1.LocalObjectReference{{Name: b.settings.DockerImagePullSecretsName}}
}

func (b *Workload) container() corev1.Container {
	// return corev1.Container{}
	falseValue := false

	service := &b.deployment.ManifestGroup().Services[b.serviceIdx]
	// sparams := b.deployment.ClusterParams().SchedulerParams[b.serviceIdx]

	kcontainer := corev1.Container{
		Name:    service.Name,
		Image:   service.Image,
		Command: service.Command,
		Args:    service.Args,
		Resources: corev1.ResourceRequirements{
			Limits:   make(corev1.ResourceList),
			Requests: make(corev1.ResourceList),
		},
		ImagePullPolicy: corev1.PullIfNotPresent,
		SecurityContext: &corev1.SecurityContext{
			RunAsNonRoot:             &falseValue,
			Privileged:               &falseValue,
			AllowPrivilegeEscalation: &falseValue,
		},
	}

	if cpu := service.Resources.CPU; cpu != nil {
		requestedCPU := computeCommittedResources(b.settings.CPUCommitLevel, cpu.Units)
		kcontainer.Resources.Requests[corev1.ResourceCPU] = resource.NewScaledQuantity(int64(requestedCPU.Val.Uint64()), resource.Milli).DeepCopy()
		kcontainer.Resources.Limits[corev1.ResourceCPU] = resource.NewScaledQuantity(int64(cpu.Units.Val.Uint64()), resource.Milli).DeepCopy()
	}

	if mem := service.Resources.Memory; mem != nil {
		requestedMem := computeCommittedResources(b.settings.MemoryCommitLevel, mem.Quantity)
		kcontainer.Resources.Requests[corev1.ResourceMemory] = resource.NewQuantity(int64(requestedMem.Val.Uint64()), resource.DecimalSI).DeepCopy()
		kcontainer.Resources.Limits[corev1.ResourceMemory] = resource.NewQuantity(int64(mem.Quantity.Val.Uint64()), resource.DecimalSI).DeepCopy()
	}

	for _, ephemeral := range service.Resources.Storage {
		attr := ephemeral.Attributes.Find(StorageAttributePersistent)
		if persistent, _ := attr.AsBool(); !persistent {
			requestedStorage := computeCommittedResources(b.settings.StorageCommitLevel, ephemeral.Quantity)
			kcontainer.Resources.Requests[corev1.ResourceEphemeralStorage] = resource.NewQuantity(int64(requestedStorage.Val.Uint64()), resource.DecimalSI).DeepCopy()
			kcontainer.Resources.Limits[corev1.ResourceEphemeralStorage] = resource.NewQuantity(int64(ephemeral.Quantity.Val.Uint64()), resource.DecimalSI).DeepCopy()

			break
		}
	}

	if service.Params != nil {
		for _, params := range service.Params.Storage {
			kcontainer.VolumeMounts = append(kcontainer.VolumeMounts, corev1.VolumeMount{
				// matches VolumeName in persistentVolumeClaims below
				Name:      fmt.Sprintf("%s-%s", service.Name, params.Name),
				ReadOnly:  params.ReadOnly,
				MountPath: params.Mount,
			})
		}
	}

	envVarsAdded := make(map[string]int)
	for _, env := range service.Env {
		parts := strings.SplitN(env, "=", 2)
		switch len(parts) {
		case 2:
			kcontainer.Env = append(kcontainer.Env, corev1.EnvVar{Name: parts[0], Value: parts[1]})
		case 1:
			kcontainer.Env = append(kcontainer.Env, corev1.EnvVar{Name: parts[0]})
		}
		envVarsAdded[parts[0]] = 0
	}
	kcontainer.Env = b.addEnvVarsForDeployment(envVarsAdded, kcontainer.Env)

	for _, expose := range service.Expose {
		kcontainer.Ports = append(kcontainer.Ports, corev1.ContainerPort{
			ContainerPort: int32(expose.Port),
		})
	}

	buf, err := json.Marshal(kcontainer)
	if err != nil {
		fmt.Printf("Marshal err %s", err.Error())
	}
	fmt.Printf("deployment %#v", string(buf))

	return kcontainer
}

func computeCommittedResources(factor float64, rv manifest.ResourceValue) manifest.ResourceValue {
	// If the value is less than 1, commit the original value. There is no concept of undercommit
	if factor <= 1.0 {
		return rv
	}

	v := rv.Val.Uint64()
	fraction := 1.0 / factor
	committedValue := math.Round(float64(v) * fraction)

	// Don't return a value of zero, since this is used as a resource request
	if committedValue <= 0 {
		committedValue = 1
	}

	result := manifest.ResourceValue{
		Val: sdk.NewInt(int64(committedValue)),
	}

	return result
}

func (b *Workload) addEnvVarsForDeployment(envVarsAlreadyAdded map[string]int, env []corev1.EnvVar) []corev1.EnvVar {
	// TODO: add envVarsAlreadyAdded to evn
	return env
}

func (b *Workload) persistentVolumeClaims() []corev1.PersistentVolumeClaim {
	var pvcs []corev1.PersistentVolumeClaim // nolint:prealloc

	service := &b.deployment.ManifestGroup().Services[b.serviceIdx]

	for _, storage := range service.Resources.Storage {
		attr := storage.Attributes.Find(StorageAttributePersistent)
		if persistent, valid := attr.AsBool(); !valid || !persistent {
			continue
		}

		volumeMode := corev1.PersistentVolumeFilesystem
		pvc := corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s-%s", service.Name, storage.Name),
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				Resources: corev1.ResourceRequirements{
					Limits:   make(corev1.ResourceList),
					Requests: make(corev1.ResourceList),
				},
				VolumeMode:       &volumeMode,
				StorageClassName: nil,
				DataSource:       nil, // bind to existing pvc. akash does not support it. yet
			},
		}

		pvc.Spec.Resources.Requests[corev1.ResourceStorage] = resource.NewQuantity(int64(storage.Quantity.Val.Uint64()), resource.DecimalSI).DeepCopy()

		attr = storage.Attributes.Find(StorageAttributePersistent)
		if class, valid := attr.AsString(); valid && class != StorageClassDefault {
			pvc.Spec.StorageClassName = &class
		}

		pvcs = append(pvcs, pvc)
	}

	return pvcs
}
