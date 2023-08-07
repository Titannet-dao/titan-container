package builder

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Deployment interface {
	workloadBase
	Create() (*appsv1.Deployment, error)
	Update(obj *appsv1.Deployment) (*appsv1.Deployment, error)
}

type deployment struct {
	Workload
}

func NewDeployment(workload Workload) Deployment {
	d := &deployment{
		Workload: workload,
	}
	return d
}

func (b *deployment) Create() (*appsv1.Deployment, error) { // nolint:golint,unparam
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: b.Name(),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: b.labels(),
			},
			Replicas: b.replicas(),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: b.labels(),
				},
				Spec: corev1.PodSpec{
					Containers:       []corev1.Container{b.container()},
					ImagePullSecrets: b.imagePullSecrets(),
				},
			},
		},
	}

	return deployment, nil
}

func (b *deployment) Update(obj *appsv1.Deployment) (*appsv1.Deployment, error) { // nolint:golint,unparam
	obj.Labels = b.labels()
	obj.Spec.Selector.MatchLabels = b.labels()
	obj.Spec.Replicas = b.replicas()
	obj.Spec.Template.Labels = b.labels()
	obj.Spec.Template.Spec.Containers = []corev1.Container{b.container()}
	obj.Spec.Template.Spec.ImagePullSecrets = b.imagePullSecrets()

	return obj, nil
}
