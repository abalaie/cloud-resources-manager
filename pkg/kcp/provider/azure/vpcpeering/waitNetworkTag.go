package vpcpeering

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"time"
)

func waitNetworkTag(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	if state.remotePeering != nil {
		return nil, nil
	}

	// If VpcNetwork is found but tags don't match user can recover by adding tag to remote VPC network so, we are
	// adding stop with requeue delay of one minute.
	if pointer.StringDeref(state.remoteVpc.Tags["shoot-name"], "") != state.Scope().Spec.ShootName {

		var kv []any

		for k, v := range state.remoteVpc.Tags {
			kv = append(kv, k, v)
		}

		logger.Info("Loaded remote VPC Network have no matching tags", kv...)

		return composed.UpdateStatus(obj).
			SetCondition(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  cloudcontrolv1beta1.ReasonFailedLoadingRemoteVpcNetwork,
				Message: "Loaded remote Vpc network has no matching tags",
			}).
			ErrorLogMessage("Error updating VpcPeering status due to remote vpc network tag mismatch").
			FailedError(composed.StopWithRequeue).
			SuccessError(composed.StopWithRequeueDelay(time.Minute)).
			Run(ctx, state)
	}

	return nil, nil
}
