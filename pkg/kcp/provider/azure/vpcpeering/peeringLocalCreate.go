package vpcpeering

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func peeringLocalCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.localPeering != nil {
		return nil, nil
	}

	// params must be the same as in peeringLocalLoad()
	err := state.localClient.CreatePeering(
		ctx,
		state.localNetworkId.ResourceGroup,
		state.localNetworkId.NetworkName(),
		state.ObjAsVpcPeering().Name,
		state.remoteNetworkId.String(),
		true,
	)
	if err != nil {
		logger.Error(err, "Error creating VPC Peering")

		if azuremeta.IsTooManyRequests(err) {
			return composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx
		}

		state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.ErrorState)

		return composed.PatchStatus(state.ObjAsVpcPeering()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonFailedCreatingVpcPeeringConnection,
				Message: "Error creating local Azure peering",
			}).
			ErrorLogMessage("Error updating KCP VpcPeering status on failed creation of local vpc peering").
			FailedError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)
	}

	logger.Info("Azure local Peering is created")

	return nil, nil
}
