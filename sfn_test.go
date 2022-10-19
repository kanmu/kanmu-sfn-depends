package sfndepents_test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
	sfndepents "github.com/kanmu/sfn-depends"
	"github.com/stretchr/testify/assert"
)

type MockSfnAPI struct {
	MockListExecutions func(ctx context.Context, params *sfn.ListExecutionsInput, optFns ...func(*sfn.Options)) (*sfn.ListExecutionsOutput, error)
}

func (api *MockSfnAPI) ListExecutions(ctx context.Context, params *sfn.ListExecutionsInput, optFns ...func(*sfn.Options)) (*sfn.ListExecutionsOutput, error) {
	return api.MockListExecutions(ctx, params, optFns...)
}

func TestSuccess(t *testing.T) {
	assert := assert.New(t)

	client := &sfndepents.Client{
		Sfn: &MockSfnAPI{
			MockListExecutions: func(ctx context.Context, params *sfn.ListExecutionsInput, optFns ...func(*sfn.Options)) (*sfn.ListExecutionsOutput, error) {
				now := time.Now()
				output := &sfn.ListExecutionsOutput{
					Executions: []types.ExecutionListItem{
						{ExecutionArn: aws.String("exec1"), Status: types.ExecutionStatusSucceeded, StartDate: aws.Time(now)},
						{ExecutionArn: aws.String("exec2"), Status: types.ExecutionStatusSucceeded, StartDate: aws.Time(now.Add(-time.Second))},
					},
				}

				return output, nil
			},
		},
	}

	err := client.Validate([]string{"state1"}, time.Hour)
	assert.NoError(err)
}

func TestSuccessLastExecution(t *testing.T) {
	assert := assert.New(t)

	client := &sfndepents.Client{
		Sfn: &MockSfnAPI{
			MockListExecutions: func(ctx context.Context, params *sfn.ListExecutionsInput, optFns ...func(*sfn.Options)) (*sfn.ListExecutionsOutput, error) {
				now := time.Now()
				output := &sfn.ListExecutionsOutput{
					Executions: []types.ExecutionListItem{
						{ExecutionArn: aws.String("exec1"), Status: types.ExecutionStatusSucceeded, StartDate: aws.Time(now)},
						{ExecutionArn: aws.String("exec2"), Status: types.ExecutionStatusFailed, StartDate: aws.Time(now.Add(-time.Second))},
					},
				}

				return output, nil
			},
		},
	}

	err := client.Validate([]string{"state1"}, time.Hour)
	assert.NoError(err)
}

func TestNotSuccess(t *testing.T) {
	assert := assert.New(t)

	client := &sfndepents.Client{
		Sfn: &MockSfnAPI{
			MockListExecutions: func(ctx context.Context, params *sfn.ListExecutionsInput, optFns ...func(*sfn.Options)) (*sfn.ListExecutionsOutput, error) {
				now := time.Now()
				output := &sfn.ListExecutionsOutput{
					Executions: []types.ExecutionListItem{
						{ExecutionArn: aws.String("exec1"), Status: types.ExecutionStatusFailed, StartDate: aws.Time(now)},
						{ExecutionArn: aws.String("exec2"), Status: types.ExecutionStatusSucceeded, StartDate: aws.Time(now.Add(-time.Second))},
					},
				}

				return output, nil
			},
		},
	}

	err := client.Validate([]string{"state1"}, time.Hour)
	assert.EqualError(err, "last execution state is FAILED: exec1")
}

func TestExecutionNotFound(t *testing.T) {
	assert := assert.New(t)

	client := &sfndepents.Client{
		Sfn: &MockSfnAPI{
			MockListExecutions: func(ctx context.Context, params *sfn.ListExecutionsInput, optFns ...func(*sfn.Options)) (*sfn.ListExecutionsOutput, error) {
				output := &sfn.ListExecutionsOutput{
					Executions: []types.ExecutionListItem{},
				}

				return output, nil
			},
		},
		Region:    "ap-northeast-1",
		AccountId: "123456789012",
	}

	err := client.Validate([]string{"state1"}, time.Hour)
	assert.EqualError(err, "execution not found: arn:aws:states:ap-northeast-1:123456789012:stateMachine:state1")
}
