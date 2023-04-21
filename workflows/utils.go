package workflows

import (
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

const signalValApprove = "approve"
const signalValRetry = "retry"
const signalValSkip = "skip"

func ExecuteWithRetry(ctx workflow.Context, valuePtr interface{}, signalName string, activity interface{}, args ...interface{}) error {
	logger := workflow.GetLogger(ctx)
	var signalVal string
	waitForApprovalSelector := workflow.NewSelector(ctx)
	waitForApprovalSelector.AddReceive(workflow.GetSignalChannel(ctx, signalName), func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &signalVal)
	})

	for {
		err := workflow.ExecuteActivity(ctx, activity, args...).Get(ctx, &valuePtr)
		if err == nil {
			return nil // Activity completed successfully
		}

	waitForSignal:
		for {
			waitForApprovalSelector.Select(ctx)

			switch signalVal {
			case signalValRetry:
				break waitForSignal // Will break out of the inner for loop and rerun activity
			case signalValSkip:
				// Will break out of the outer for loop and return error
				return temporal.NewApplicationError("Received signal to skip", "Skipped")
			default:
				logger.Warn("invalid signal value",
					zap.String("actualValue", signalVal),
					zap.Any("expectedValues: ", []string{signalValRetry, signalValSkip}),
				)
			}
		}
	}
}

func ExecuteWithApproval(ctx workflow.Context, valuePtr interface{}, signalName string, activity interface{}, args ...interface{}) error {
	logger := workflow.GetLogger(ctx)
	var signalVal string
	waitForApprovalSelector := workflow.NewSelector(ctx)
	waitForApprovalSelector.AddReceive(workflow.GetSignalChannel(ctx, signalName), func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &signalVal)
	})

	for {
	waitForSignal:
		for {
			waitForApprovalSelector.Select(ctx)

			switch signalVal {
			case signalValRetry, signalValApprove:
				break waitForSignal // Will break out of the inner for loop and run/rerun activity
			case signalValSkip:
				// Will break out of the outer for loop and return error
				return temporal.NewApplicationError("Received signal to skip", "Skipped")
			default:
				logger.Warn("invalid signal value",
					zap.String("actualValue", signalVal),
					zap.Any("expectedValues: ", []string{signalValRetry, signalValSkip}),
				)
			}
		}

		err := workflow.ExecuteActivity(ctx, activity, args...).Get(ctx, &valuePtr)
		if err == nil {
			return nil // Activity completed successfully
		}
	}
}
