package iprange

import (
	"context"
	"fmt"
	"github.com/elliotchance/pie/v2"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func preventDeleteOnAwsNfsVolumeUsage(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !composed.MarkedForDeletionPredicate(ctx, st) {
		// SKR IpRange is NOT marked for deletion, do not delete mirror in KCP
		return nil, nil
	}

	awsNfsVolumesUsingThisIpRange := &cloudresourcesv1beta1.AwsNfsVolumeList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(cloudresourcesv1beta1.IpRangeField, st.Name().String()),
	}
	err := state.Cluster().K8sClient().List(ctx, awsNfsVolumesUsingThisIpRange, listOps)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error listing AwsNfsVolumes using IpRange", composed.StopWithRequeue, ctx)
	}

	if len(awsNfsVolumesUsingThisIpRange.Items) == 0 {
		return nil, nil
	}

	usedByAwsNfsVolumes := fmt.Sprintf("%v", pie.Map(awsNfsVolumesUsingThisIpRange.Items, func(x cloudresourcesv1beta1.AwsNfsVolume) string {
		return fmt.Sprintf("%s/%s", x.Namespace, x.Name)
	}))

	logger.
		WithValues("usedByAwsNfsVolumes", usedByAwsNfsVolumes).
		Info("IpRange marked for deleting used by AwsNfsVolume")

	state.ObjAsIpRange().Status.State = cloudresourcesv1beta1.StateWarning
	return composed.UpdateStatus(state.ObjAsIpRange()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeWarning,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionTypeDeleteWhileUsed,
			Message: fmt.Sprintf("Can not be deleted while used by: %s", usedByAwsNfsVolumes),
		}).
		ErrorLogMessage("Error updating IpRange status with Warning condition for delete while in use").
		SuccessLogMsg("Forgetting SKR IpRange marked for deleting that is in use").
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
