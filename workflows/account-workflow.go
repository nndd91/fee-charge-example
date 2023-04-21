package workflows

import (
	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"time"
)
import "go.temporal.io/sdk/workflow"

const ErrFeePreviewFailed = "FeePreviewFailed"

func AccountFeeChargeWorkflow(ctx workflow.Context, batchID string, accountID string, month time.Month, year int) error {
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
			InitialInterval:        time.Second * 30,
			BackoffCoefficient:     2,
			MaximumInterval:        time.Minute * 15,
			NonRetryableErrorTypes: []string{ErrFeePreviewFailed},
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	if err := workflow.SetQueryHandler(ctx, "feeChargeResult", func() (Result, error) {
		return feeChargeResult, nil
	}); err != nil {
		workflow.GetLogger(ctx).Error("failed to add batchID to search attrs")
	}

	err := workflow.ExecuteActivity(ctx, FetchFeePreview, accountID, month, year).Get(ctx, &feeChargeResult)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, ChargeFee, accountID, month, year).Get(ctx, &feeChargeResult)
	if err != nil {
		return err
	}

	var statementId uuid.UUID
	err = workflow.ExecuteActivity(ctx, GenerateStatement, accountID, feeChargeResult).Get(ctx, &statementId)
	if err != nil {
		return err
	}

	var sendingResult bool
	err = workflow.ExecuteActivity(ctx, SendStatement, statementId).Get(ctx, sendingResult)
	if err != nil {
		return err
	}

	return nil
}
