package nfsinstance

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/kcp/nfsinstance/types"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {

		logger := composed.LoggerFromCtx(ctx)
		state, err := stateFactory.NewState(ctx, st.(types.State))
		if err != nil {
			err = fmt.Errorf("error creating new gcp nfsInstance state: %w", err)
			logger.Error(err, "Error")
			return composed.StopAndForget, nil
		}
		return composed.ComposeActions(
			"gcsNfsInstance",
			validateAlways,
			actions.AddFinalizer,
			checkGcpOperation,
			loadNfsInstance,
			validatePostCreate,
			checkNUpdateState,
			checkUpdateMask,
			syncNfsInstance,
			composed.BuildBranchingAction("RunFinalizer", StatePredicate(client.Deleted, ctx, state),
				actions.RemoveFinalizer, nil),
		)(ctx, state)
	}
}

func StatePredicate(status v1beta1.StatusState, ctx context.Context, state *State) composed.Predicate {
	return func(ctx context.Context, st composed.State) bool {
		return status == state.curState
	}
}