package myaws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/marksauter/markus-ninja-images/pkg/util"
)

var AWSRegion = util.GetOptionalEnv("AWS_REGION", "us-east-1")

func NewSession() *session.Session {
	sess, err := session.NewSession(&aws.Config{
		// LogLevel:                      aws.LogLevel(aws.LogDebugWithHTTPBody),
		// CredentialsChainVerboseErrors: aws.Bool(true),
		Region: aws.String(AWSRegion),
	})

	if err != nil {
		panic(err)
	}

	return sess
}
