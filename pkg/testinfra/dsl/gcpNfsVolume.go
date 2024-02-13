package dsl

import (
	"context"
	"errors"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func WithGcpNfsVolumeIpRange(ipRangeName string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.GcpNfsVolume); ok {
				x.Spec.IpRange.Name = ipRangeName
				if x.Spec.IpRange.Namespace == "" {
					x.Spec.IpRange.Namespace = DefaultSkrNamespace
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithNfsVolumeIpRange", obj))
		},
	}
}

func CreateGcpNfsVolume(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.GcpNfsVolume, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudresourcesv1beta1.GcpNfsVolume{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
			WithGcpNfsValues(),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the SKR GcpNfsVolume must have name set")
	}
	if obj.Spec.IpRange.Name == "" {
		return errors.New("the SKR GcpNfsVolume must have spec.iprange.name set")
	}
	if obj.Spec.IpRange.Namespace == "" {
		obj.Spec.IpRange.Namespace = DefaultSkrNamespace
	}

	err := clnt.Create(ctx, obj)
	return err
}

func WithGcpNfsValues() ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.GcpNfsVolume); ok {
				x.Spec.Location = "us-west1"
				x.Spec.Tier = "BASIC_HDD"
				x.Spec.CapacityGb = 1024
				x.Spec.FileShareName = "test01"
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithGcpNfsValues", obj))
		},
	}
}

func AssertGcpNfsVolumeHasId() ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.GcpNfsVolume)
		if !ok {
			return fmt.Errorf("the object %T is not GcpNfsVolume", obj)
		}
		if x.Status.Id == "" {
			return errors.New("the GcpNfsVolume ID not set")
		}
		return nil
	}
}

func WithKcpNfsStatusState(state cloudcontrolv1beta1.StatusState) ObjStatusAction {
	return &objStatusAction{
		f: func(obj client.Object) {
			x := obj.(*cloudcontrolv1beta1.NfsInstance)
			x.Status.State = state
		},
	}
}
func WithKcpNfsStatusHost(host string) ObjStatusAction {
	return &objStatusAction{
		f: func(obj client.Object) {
			x := obj.(*cloudcontrolv1beta1.NfsInstance)
			x.Status.Hosts = []string{host}
		},
	}
}

func WithKcpNfsStatusCapacity(capacity int) ObjStatusAction {
	return &objStatusAction{
		f: func(obj client.Object) {
			x := obj.(*cloudcontrolv1beta1.NfsInstance)
			x.Status.CapacityGb = capacity
		},
	}
}

func AssertKcpStatusHosts(host string) ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudcontrolv1beta1.NfsInstance)
		if !ok {
			return fmt.Errorf("expected *cloudcontrolv1beta1.NfsInstance, but got %T", obj)
		}
		if len(x.Status.Hosts) > 0 && x.Status.Hosts[0] == host {
			return nil
		}
		return fmt.Errorf("the KCP NfsInstance %s/%s expected host is %s, but it has %s",
			x.Namespace, x.Name, host, x.Status.Hosts)
	}
}
