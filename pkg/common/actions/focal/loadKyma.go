package focal

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func loadKyma(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(State)
	logger := composed.LoggerFromCtx(ctx)

	if state.Scope() == nil || len(state.Scope().Name) == 0 {
		return nil, nil
	}

	kyma := util.NewKymaUnstructured()
	kymaUnstructured := util.NewKymaUnstructured()
	err := state.Cluster().K8sClient().Get(ctx, state.Name(), kymaUnstructured)
	if apierrors.IsNotFound(err) {
		logger.Info("Kyma CR does not exist")
		return nil, nil
	}
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading KCP Kyma CR", composed.StopWithRequeue, ctx)
	}

	state.SetKyma(kyma)

	return nil, nil
}
