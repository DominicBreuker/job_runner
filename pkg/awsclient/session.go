package awsclient

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

var sess *session.Session

func InitializeSession(region string) {
	conf := &aws.Config{}
	conf.Region = aws.String(region)

	sess = session.Must(session.NewSession(conf))
}

func GetDynamoDB() dynamodbiface.DynamoDBAPI {
	svc := dynamodb.New(sess)

	return svc
}
