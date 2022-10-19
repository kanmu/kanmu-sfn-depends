package sfndepents

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type SfnAPI interface {
	ListExecutions(ctx context.Context, params *sfn.ListExecutionsInput, optFns ...func(*sfn.Options)) (*sfn.ListExecutionsOutput, error)
}

type Client struct {
	Sfn       SfnAPI
	Region    string
	AccountId string
}

func NewClient() (*Client, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	accountId, err := getAccountId(cfg)

	if err != nil {
		return nil, fmt.Errorf("failed to get AWS account ID: %w", err)
	}

	sfnClient := sfn.NewFromConfig(cfg)

	client := &Client{
		Sfn:       sfnClient,
		Region:    cfg.Region,
		AccountId: accountId,
	}

	return client, nil
}

func getAccountId(cfg aws.Config) (string, error) {
	stsClient := sts.NewFromConfig(cfg)
	input := &sts.GetCallerIdentityInput{}
	output, err := stsClient.GetCallerIdentity(context.Background(), input)

	if err != nil {
		return "", nil
	}

	return aws.ToString(output.Account), nil
}

func (client *Client) Validate(stateMachines []string, period time.Duration) error {
	for _, stateMachine := range stateMachines {
		log.Printf("validate %s\n", stateMachine)

		arn := client.buildArn(stateMachine)
		startFrom := time.Now().Add(-period)
		executions, err := client.listExecutions(arn, startFrom)

		if err != nil {
			return fmt.Errorf("failed to list executions: %w", err)
		}

		if len(executions) == 0 {
			return fmt.Errorf("execution not found: %s", arn)
		}

		lastExecution := executions[0]

		if lastExecution.Status != types.ExecutionStatusSucceeded {
			return fmt.Errorf("last execution state is %s: %s", lastExecution.Status, aws.ToString(lastExecution.ExecutionArn))
		}

		log.Println("state machine is successfully completed.")
	}

	return nil
}

func (client *Client) listExecutions(stateMachineArn string, startFrom time.Time) ([]types.ExecutionListItem, error) {
	input := &sfn.ListExecutionsInput{
		StateMachineArn: aws.String(stateMachineArn),
	}

	output, err := client.Sfn.ListExecutions(context.Background(), input)

	if err != nil {
		return nil, err
	}

	sort.Slice(output.Executions, func(i, j int) bool {
		e1 := output.Executions[i]
		e2 := output.Executions[j]

		// sort descending
		return e1.StartDate.After(aws.ToTime(e2.StartDate))
	})

	executions := []types.ExecutionListItem{}

	for _, e := range output.Executions {
		if e.StartDate.Equal(startFrom) || e.StartDate.After(startFrom) {
			executions = append(executions, e)
		}
	}

	return executions, nil
}

func (client *Client) buildArn(stateMachine string) string {
	return fmt.Sprintf("arn:aws:states:%s:%s:stateMachine:%s", client.Region, client.AccountId, stateMachine)
}
