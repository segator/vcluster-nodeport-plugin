package hook

import (
	"context"
	"fmt"

	"github.com/loft-sh/vcluster-sdk/plugin"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewServiceHook(nodePorts map[string]int32, labelSelector map[string]string) plugin.ClientHook {
	return &serviceHook{
		nodePorts:     nodePorts,
		labelSelector: labelSelector,
	}
}

type serviceHook struct {
	nodePorts     map[string]int32
	labelSelector map[string]string
}

func (s *serviceHook) Name() string {
	return "service-hook"
}

func (s *serviceHook) Resource() client.Object {
	return &corev1.Service{}
}

var _ plugin.MutateCreatePhysical = &serviceHook{}

func (s *serviceHook) MutateCreatePhysical(ctx context.Context, obj client.Object) (client.Object, error) {
	return s.MutateUpdatePhysical(ctx, obj)
}
func (s *serviceHook) MutateUpdatePhysical(ctx context.Context, obj client.Object) (client.Object, error) {
	service, ok := obj.(*corev1.Service)
	if !ok {
		return nil, fmt.Errorf("object %v is not a service", obj)
	}

	s.nodePort(service)
	return service, nil
}

var _ plugin.MutateGetVirtual = &serviceHook{}

// MutateGetVirtual fakes the service vcluster "sees" so that it is not trying to update the
// ports all the time
func (s *serviceHook) MutateGetVirtual(ctx context.Context, obj client.Object) (client.Object, error) {
	service, ok := obj.(*corev1.Service)
	if !ok {
		return nil, fmt.Errorf("object %v is not a service", obj)
	}

	s.nodePort(service)
	return service, nil
}

func (s *serviceHook) nodePort(service *corev1.Service) {
	if service.Spec.Type != corev1.ServiceTypeNodePort && service.Spec.Type != corev1.ServiceTypeLoadBalancer {
		return
	}

	// Check if label selector is specified and if the service matches the labels
	if len(s.labelSelector) > 0 {
		match := true
		for key, value := range s.labelSelector {
			if service.Labels == nil || service.Labels[key] != value {
				match = false
				break
			}
		}

		if !match {
			return
		}
	}

	for i := range service.Spec.Ports {
		expectedPort := s.nodePorts[service.Spec.Ports[i].Name]
		if expectedPort != 0 {
			service.Spec.Ports[i].NodePort = expectedPort
		}
	}
}
