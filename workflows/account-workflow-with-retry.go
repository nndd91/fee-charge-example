package workflows

import (
	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"time"
)
import "go.temporal.io/sdk/workflow"

func AccountFeeChargeWorkflowWithRetry(ctx workflow.Context, batchID string, accountID string, month time.Month, year int) error {
	if err := workflow.UpsertSearchAttributes(ctx, map[string]interface{}{
		"BatchId": batchID,
	}); err != nil {
		workflow.GetLogger(ctx).Error("failed to add batchID to search attrs")
	}

	feeChargeResult := Result{}
	ao := workflow.ActivityOptions{
		StartToCloseTimeout:    time.Minute * 5,
		ScheduleToCloseTimeout: time.Hour * 2,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second * 30,
			BackoffCoefficient: 2,
			MaximumInterval:    time.Minute * 15,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	if err := workflow.SetQueryHandler(ctx, "feeChargeResult", func() (Result, error) {
		return feeChargeResult, nil
	}); err != nil {
		workflow.GetLogger(ctx).Error("failed to add batchID to search attrs")
	}

	err := ExecuteWithApproval(ctx, &feeChargeResult, "approveFeePreview", FetchFeePreview, accountID, month, year)
	if err != nil {
		return err
	}

	err = upsertFeeData(ctx, feeChargeResult)

	err = ExecuteWithApproval(ctx, &feeChargeResult, "approveFeeCharge", ChargeFee, accountID, month, year)
	if err != nil {
		return err
	}

	err = upsertFeeData(ctx, feeChargeResult)

	var statementId uuid.UUID
	err = ExecuteWithApproval(ctx, &statementId, "approveGenerateStatement", GenerateStatement, accountID, feeChargeResult)
	if err != nil {
		return err
	}

	var sendingResult bool
	err = ExecuteWithApproval(ctx, &sendingResult, "approveSend", SendStatement, statementId)
	if err != nil {
		return err
	}

	return nil
}
