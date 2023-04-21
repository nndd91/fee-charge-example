package main

import (
	"context"
	"fmt"
	"github.com/nndd91/fee-charge-example/config"
	"github.com/nndd91/fee-charge-example/workflows"
	"go.temporal.io/sdk/client"
)

func main() {
	serviceClient, err := client.NewLazyClient(client.Options{
		Namespace: config.Namespace,
		HostPort:  config.HostPort,
	})
	if err != nil {
		panic(fmt.Sprintf("unable to create temporal service client: %v", err))
	}

	wfr, err := serviceClient.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
		ID: "AccountFeeCharge_account-1",
	}, workflows.AccountFeeChargeWorkflow, "123", "account-1", 1, 2021)
	if err != nil {
		panic(fmt.Sprintf("unable to start workflow: %v", err))
	}

	fmt.Println(fmt.Sprintf("Workflow started: ID: %s, RunID: %s", wfr.GetID(), wfr.GetRunID()))
}
