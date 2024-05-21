package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	cftype "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/aws/aws-sdk-go-v2/service/codepipeline"
	cptype "github.com/aws/aws-sdk-go-v2/service/codepipeline/types"
	"github.com/aws/aws-sdk-go/aws"
)

// Load the default AWS SDK configuration.
var sdkConfig, _ = config.LoadDefaultConfig(context.TODO())

// This function sends a message to the CodePipeline job whether the Lambda function was successful or not.
func sendResults(success bool, jobId string, pipelineErr error) {
	codePipelineClient := codepipeline.NewFromConfig(sdkConfig)

	// If the job was not successful, send a failure message to the CodePipeline job.
	if !success {
		codePipelineClient.PutJobFailureResult(context.TODO(), &codepipeline.PutJobFailureResultInput{
			FailureDetails: &cptype.FailureDetails{
				Message:             aws.String(pipelineErr.Error()),
				Type:                "JobFailed",
				ExternalExecutionId: aws.String(fmt.Sprintf("lambda-%d", time.Now().Unix())),
			},
			JobId: aws.String(jobId),
		})
	}

	// Else, send a success message to the CodePipeline job.
	codePipelineClient.PutJobSuccessResult(context.TODO(), &codepipeline.PutJobSuccessResultInput{
		JobId: aws.String(jobId),
	})
}

func handler(ctx context.Context, evt events.CodePipelineJobEvent) {
	// This section is for invalidating the CloudFront cache.
	cloudFrontClient := cloudfront.NewFromConfig(sdkConfig)
	res, err := cloudFrontClient.CreateInvalidation(context.TODO(), &cloudfront.CreateInvalidationInput{
		DistributionId: aws.String(os.Getenv("DISTRIBUTION_ID")),
		InvalidationBatch: &cftype.InvalidationBatch{
			CallerReference: aws.String(fmt.Sprintf("invalidation-%d", time.Now().Unix())),
			Paths: &cftype.Paths{
				Quantity: aws.Int32(1),
				Items: []string{
					*aws.String("/*"),
				},
			},
		},
	})

	// If there was an error, call the sendResults function with the error.
	if err != nil {
		log.Fatalln(err)
		sendResults(false, evt.CodePipelineJob.ID, err)
	}

	// Else, log the invalidation details and call the sendResults function with a success message.
	log.Printf("Invalidation ID: %v\n", *res.Invalidation.Id)
	sendResults(true, evt.CodePipelineJob.ID, nil)
}

func main() {
	lambda.Start(handler)
}
