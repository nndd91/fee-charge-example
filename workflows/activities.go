package workflows

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go.temporal.io/sdk/activity"
	"math/rand"
	"time"
)

type Result struct {
	TotalFeeAmount float64 `json:"totalFeeAmount"`
}

func FetchFeePreview(ctx context.Context, accountID string, month time.Month, year int) (Result, error) {
	logger := activity.GetLogger(ctx)
	logger.Info(fmt.Sprintf("Fetching fee preview for accountID: %s, %v-%s", accountID, year, month))

	return Result{TotalFeeAmount: rand.Float64()}, nil
}

func ChargeFee(ctx context.Context, accountID string, month time.Month, year int) (Result, error) {
	logger := activity.GetLogger(ctx)
	logger.Info(fmt.Sprintf("Charging fees for accountID: %s, %v-%s", accountID, year, month))

	return Result{TotalFeeAmount: rand.Float64()}, nil
}

func GenerateStatement(ctx context.Context, accountID string, month time.Month, year int) (uuid.UUID, error) {
	logger := activity.GetLogger(ctx)
	logger.Info(fmt.Sprintf("Generating statement for accountID: %s, %v-%s", accountID, year, month))

	return uuid.New(), nil
}

func SendStatement(ctx context.Context, statementId uuid.UUID) (bool, error) {
	logger := activity.GetLogger(ctx)
	logger.Info(fmt.Sprintf("Sending statement: %s", statementId))

	return true, nil
}
