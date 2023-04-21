package workflows

import (
	"context"
	"fmt"
	"github.com/nndd91/fee-charge-example/config"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
	"go.temporal.io/server/common/payload"
	"go.uber.org/zap"
	"time"
)

var accountIDs = []string{"account-1", "account-2"}

func newClient() (client.Client, error) {
	return client.NewLazyClient(client.Options{
		Namespace: config.Namespace,
		HostPort:  config.HostPort,
	})
}

func InitiateAccountFeeCharge(ctx context.Context, accountIDs []string, batchID string, month time.Month, year int) (int, error) {
	serviceClient, err := newClient()
	count := 0
	if err != nil {
		return count, err
	}
	// We can also add heartbeat to this workflow to track progress / resume from failures
	for _, accountID := range accountIDs {
		_, err := serviceClient.ExecuteWorkflow(ctx,
			client.StartWorkflowOptions{
				ID: "AccountFeeCharge_" + accountID,
			},
			AccountFeeChargeWorkflow, accountID, batchID, month, year)
		if err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

type AggregateResult struct {
	TotalFeeAmount float64
	Count          int
}

func AggregateWorkflow(ctx context.Context, batchID string) (AggregateResult, error) {
	logger := activity.GetLogger(ctx)
	var aggResult AggregateResult
	serviceClient, err := newClient()
	if err != nil {
		return aggResult, err
	}
	var nextPage []byte

	for {
		res, err := serviceClient.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
			Namespace:     config.Namespace,
			Query:         fmt.Sprintf("BatchId = '%s'and ExecutionStatus in (1)", batchID),
			NextPageToken: nextPage,
		})
		if err != nil {
			return aggResult, err
		}

		for _, wf := range res.Executions {
			v, ok := wf.SearchAttributes.GetIndexedFields()["FeeData"]
			if !ok {
				logger.Warn("wf does not have FeeData search attribute",
					zap.String("workflowID", wf.GetExecution().GetWorkflowId()))
				continue
			}
			var result Result
			if err := payload.Decode(v, &result); err != nil {
				return aggResult, fmt.Errorf("failed to decode search attr: %v for %s", err, wf.GetExecution().GetWorkflowId())
			}
			aggResult.TotalFeeAmount += result.TotalFeeAmount
			aggResult.Count++
		}
		nextPage = res.GetNextPageToken()
		if nextPage == nil {
			break
		}
	}
	return aggResult, nil
}

func InitiationWorkflow(ctx workflow.Context, month time.Month, year int) error {
	batchID := fmt.Sprintf("%d-%d", month, year)
	var count int
	err := workflow.ExecuteActivity(ctx, InitiateAccountFeeCharge, accountIDs, batchID, month, year).Get(ctx, &count)
	if err != nil {
		return err
	}

	return nil
}
