package iprange

import (
	"context"
	"fmt"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	logger.WithValues("kcpIpRangeConditions", pie.Map(state.KcpIpRange.Status.Conditions, func(c metav1.Condition) string {
		return fmt.Sprintf("%s:%s", c.Type, c.Status)
	})).Info("Updating SKR IpRange status from KCP IpRange conditions")

	condErr := meta.FindStatusCondition(state.KcpIpRange.Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	condReady := meta.FindStatusCondition(state.KcpIpRange.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)
	kcpMarkedForDeletion := composed.IsMarkedForDeletion(state.KcpIpRange)

	if condErr != nil {
		logger.Info("Updating IpRange status with Error condition")
		return composed.UpdateStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: condErr.Message,
			}).
			DeriveStateFromConditions(state.MapConditionToState()).
			ErrorLogMessage("Error updating IpRange status with not ready condition due to KCP error").
			Run(ctx, state)
	}

	if condReady != nil && !kcpMarkedForDeletion {
		logger.Info("Updating IpRange status with Ready condition")
		return composed.UpdateStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeReady,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionTypeReady,
				Message: condReady.Message,
			}).
			DeriveStateFromConditions(state.MapConditionToState()).
			ErrorLogMessage("Error updating IpRange status with ready condition").
			Run(ctx, state)
	}

	// keep looping until KCP IpRange gets some condition
	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
