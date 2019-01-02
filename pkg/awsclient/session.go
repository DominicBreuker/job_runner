package awsclient

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
)

var sess *session.Session

func InitializeSession(region string) {
	conf := &aws.Config{}
	conf.Region = aws.String(region)

	sess = session.Must(session.NewSession(conf))
}

func GetSession() *session.Session {
	return sess
}

func GetDynamoDB() dynamodbiface.DynamoDBAPI {
	svc := dynamodb.New(sess)

	return svc
}

func GetSNS() snsiface.SNSAPI {
	svc := sns.New(sess)

	return svc
}
