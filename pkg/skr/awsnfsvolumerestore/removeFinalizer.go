package awsnfsvolumerestore

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func removeFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if !composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	controllerutil.RemoveFinalizer(state.Obj(), cloudresourcesv1beta1.Finalizer)
	err := state.UpdateObj(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error saving SKR AwsNfsVolumeBackup after finalizer remove", composed.StopWithRequeue, ctx)
	}

	return composed.StopAndForget, nil
}
