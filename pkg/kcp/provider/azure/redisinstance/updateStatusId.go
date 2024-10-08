package redisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func updateStatusId(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	redisInstance := state.ObjAsRedisInstance()

	if redisInstance.Status.Id != "" { // already set
		return nil, nil
	}

	redisInstance.Status.Id = *(state.azureRedisInstance.Name)

	err := state.UpdateObjStatus(ctx)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating Azure RedisInstance success .status.id", composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeue, nil
}
