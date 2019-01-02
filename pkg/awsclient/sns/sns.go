package sns

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/dominicbreuker/job_runner/pkg/awsclient"
)

var svc = awsclient.GetSNS

func Publish(subject, message, topic string) error {
	req := &sns.PublishInput{}
	req.SetSubject(subject)
	req.SetMessage(message)
	req.SetTopicArn(topic)

	_, err := svc().Publish(req)
	if err != nil {
		return fmt.Errorf("publishing message %s to topic %s: %v", subject, topic, err)
	}

	return nil
}
