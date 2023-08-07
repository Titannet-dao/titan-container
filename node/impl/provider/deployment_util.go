package provider

import (
	"fmt"
	"strings"

	"github.com/Filecoin-Titan/titan-container/api/types"
	"github.com/Filecoin-Titan/titan-container/node/impl/provider/kube/builder"
	"github.com/Filecoin-Titan/titan-container/node/impl/provider/kube/manifest"
	"github.com/google/uuid"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	podReplicas = 1
)

func ClusterDeploymentFromDeployment(deployment *types.Deployment) (builder.IClusterDeployment, error) {
	if len(deployment.ID) == 0 {
		return nil, fmt.Errorf("deployment ID can not empty")
	}

	deploymentID := manifest.DeploymentID{ID: string(deployment.ID), Owner: deployment.Owner}
	group, err := deploymentToManifestGroup(deployment)
	if err != nil {
		return nil, err
	}

	settings := builder.ClusterSettings{
		SchedulerParams: make([]*builder.SchedulerParams, len(group.Services)),
	}

	return &builder.ClusterDeployment{
		Did:     deploymentID,
		Group:   group,
		Sparams: settings,
	}, nil
}

func deploymentToManifestGroup(deployment *types.Deployment) (*manifest.Group, error) {
	if len(deployment.Services) == 0 {
		return nil, fmt.Errorf("deployment service can not empty")
	}

	services := make([]manifest.Service, 0, len(deployment.Services))
	for _, service := range deployment.Services {
		s, err := serviceToManifestService(service, deployment.Authority)
		if err != nil {
			return nil, err
		}
		services = append(services, s)
	}

	return &manifest.Group{Services: services}, nil
}

func serviceToManifestService(service *types.Service, Authority bool) (manifest.Service, error) {
	if len(service.Image) == 0 {
		return manifest.Service{}, fmt.Errorf("service image can not empty")
	}
	name := imageToServiceName(service.Image)
	resource := resourceToManifestResource(&service.ComputeResources)
	exposes, err := exposesFromPorts(service.Ports)
	if err != nil {
		return manifest.Service{}, err
	}

	s := manifest.Service{
		Name:      name,
		Image:     service.Image,
		Args:      service.Arguments,
		Env:       envToManifestEnv(service.Env),
		Resources: &resource,
		Expose:    make([]*manifest.ServiceExpose, 0),
		Count:     podReplicas,
	}

	if len(exposes) > 0 {
		s.Expose = append(s.Expose, exposes...)
	}

	return s, nil
}

func envToManifestEnv(serviceEnv types.Env) []string {
	envs := make([]string, 0, len(serviceEnv))
	for k, v := range serviceEnv {
		env := fmt.Sprintf("%s=%s", k, v)
		envs = append(envs, env)
	}
	return envs
}

func imageToServiceName(image string) string {
	names := strings.Split(image, "/")
	names = strings.Split(names[len(names)-1], ":")
	serviceName := names[0]

	uuidString := uuid.NewString()
	uuidString = strings.Replace(uuidString, "-", "", -1)

	return fmt.Sprintf("%s-%s", serviceName, uuidString)
}

func resourceToManifestResource(resource *types.ComputeResources) manifest.ResourceUnits {
	return *manifest.NewResourceUnits(uint64(resource.CPU*1000), uint64(resource.Memory*1000000), uint64(resource.Storage*1000000))
}

func serviceProto(protocol types.Protocol) (manifest.ServiceProtocol, error) {
	if len(protocol) == 0 {
		return manifest.TCP, nil
	}

	proto := strings.ToUpper(string(protocol))
	serviceProto := manifest.ServiceProtocol(proto)
	if serviceProto != manifest.TCP && serviceProto != manifest.UDP {
		return "", fmt.Errorf("it's neither tcp nor udp")
	}
	return serviceProto, nil
}

func exposesFromPorts(ports types.Ports) ([]*manifest.ServiceExpose, error) {
	if len(ports) == 0 {
		return nil, nil
	}

	serviceExposes := make([]*manifest.ServiceExpose, 0, len(ports))
	for _, port := range ports {
		proto, err := serviceProto(port.Protocol)
		if err != nil {
			return nil, err
		}
		serviceExpose := &manifest.ServiceExpose{Port: uint32(port.Port), ExternalPort: uint32(port.Port), Proto: proto, Global: true}
		serviceExposes = append(serviceExposes, serviceExpose)
	}
	return serviceExposes, nil
}

func k8sDeploymentsToServices(deploymentList *appsv1.DeploymentList) ([]*types.Service, error) {
	services := make([]*types.Service, 0, len(deploymentList.Items))

	for _, deployment := range deploymentList.Items {
		s, err := k8sDeploymentToService(&deployment)
		if err != nil {
			return nil, err
		}
		services = append(services, s)
	}

	return services, nil
}

func k8sDeploymentToService(deployment *appsv1.Deployment) (*types.Service, error) {
	if len(deployment.Spec.Template.Spec.Containers) == 0 {
		return nil, fmt.Errorf("deployment container can not empty")
	}

	container := deployment.Spec.Template.Spec.Containers[0]
	service := &types.Service{Image: container.Image, Name: container.Name}
	service.CPU = container.Resources.Limits.Cpu().AsApproximateFloat64()
	service.Memory = container.Resources.Limits.Memory().Value() / 1000000
	service.Storage = int64(container.Resources.Limits.StorageEphemeral().AsApproximateFloat64()) / 1000000

	status := types.ReplicasStatus{
		TotalReplicas:     int(deployment.Status.Replicas),
		ReadyReplicas:     int(deployment.Status.ReadyReplicas),
		AvailableReplicas: int(deployment.Status.AvailableReplicas),
	}
	service.Status = status

	return service, nil
}

func k8sServiceToPortMap(serviceList *corev1.ServiceList) (map[string]types.Ports, error) {
	portMap := make(map[string]types.Ports)
	for _, service := range serviceList.Items {
		serviceName := strings.TrimSuffix(service.Name, builder.SuffixForNodePortServiceName)

		ports := servicePortsToPortPairs(service.Spec.Ports)
		portMap[serviceName] = ports
	}
	return portMap, nil
}

func servicePortsToPortPairs(servicePorts []corev1.ServicePort) types.Ports {
	ports := make([]types.Port, 0, len(servicePorts))
	for _, servicePort := range servicePorts {
		port := types.Port{Port: int(servicePort.TargetPort.IntVal), Protocol: types.Protocol(servicePort.Protocol)}
		if servicePort.NodePort != 0 {
			port.ExposePort = int(servicePort.NodePort)
		}
		ports = append(ports, port)
	}
	return types.Ports(ports)
}
