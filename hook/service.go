package hook

import (
	"context"
	"fmt"

	"github.com/loft-sh/vcluster-sdk/plugin"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewServiceHook() plugin.ClientHook {
	return &serviceHook{}
}

type serviceHook struct {
	nodePorts map[string]int32
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
	for _, p := range service.Spec.Ports {
		expectedPort := s.nodePorts[p.Name]
		if expectedPort != 0 {
			p.Port = expectedPort
		}
	}
}
