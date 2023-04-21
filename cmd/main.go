package main

import (
	"fmt"
	"github.com/nndd91/fee-charge-example/config"
	"github.com/nndd91/fee-charge-example/workflows"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"
)

// NewTemporalWorker starts a worker process. Any err should result in a panic or fatal and crash the pods.
func NewTemporalWorker(logger *zap.Logger) *worker.Worker {
	// The client and worker are heavyweight objects that should be created once per process.
	serviceClient, err := client.NewLazyClient(client.Options{
		Namespace: config.Namespace,
		HostPort:  config.HostPort,
	})
	if err != nil {
		panic(fmt.Sprintf("unable to create temporal service client: %v", err))
	}

	temporalWorker := worker.New(serviceClient, config.TaskQueue, worker.Options{
		// Limit the number of activities that can be started per second per task queue.
		TaskQueueActivitiesPerSecond: config.TaskQueueActivitiesPerSecond,
		// Limit the number of activities that can be executed concurrently per task queue.
		MaxConcurrentActivityExecutionSize: config.MaxConcurrentActivityExecutionSize,
	})

	// Register workflows and activities
	temporalWorker.RegisterWorkflow(workflows.AccountFeeChargeWorkflow)
	temporalWorker.RegisterWorkflow(workflows.AccountFeeChargeWorkflowWithRetry)
	temporalWorker.RegisterWorkflow(workflows.InitiationWorkflow)
	temporalWorker.RegisterWorkflow(workflows.AggregateWorkflow)

	temporalWorker.RegisterActivity(workflows.FetchFeePreview)
	temporalWorker.RegisterActivity(workflows.ChargeFee)
	temporalWorker.RegisterActivity(workflows.GenerateStatement)
	temporalWorker.RegisterActivity(workflows.SendStatement)

	logger.Info("Loaded workflows and activities. Creating worker..")
	return &temporalWorker
}

func main() {
	logger, _ := zap.NewDevelopment()
	w := NewTemporalWorker(logger)

	err := (*w).Run(worker.InterruptCh())

	if err != nil {
		logger.Error("Error from worker", zap.Error(err))
	}
}
