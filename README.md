# Fee Charge Sample

This is a code sample to accompany my presentation during Temporal Event on the 26 April 2023. It contains code snippets on how
to implement a fee charge service using Temporal. The code is written in Go and uses the Temporal Go SDK.

## Running
- `go run cmd/main.go` will start the worker
- `tctl --ns default wf start --tq fee-charge --wt FeeChargeWorkflow -i '1 1 4 2023'` will start fee charge workflow