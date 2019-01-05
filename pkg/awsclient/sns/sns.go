package sns

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
)

type Client struct {
	SVC snsiface.SNSAPI
}

func GetClient(svc snsiface.SNSAPI) *Client {
	return &Client{
		SVC: svc,
	}
}

func (c *Client) Publish(subject, message, topic string) error {
	if topic == "" {
		return fmt.Errorf("topic cannot be empty")
	}

	req := &sns.PublishInput{}
	req.SetSubject(subject)
	req.SetMessage(message)
	req.SetTopicArn(topic)

	_, err := c.SVC.Publish(req)
	if err != nil {
		return fmt.Errorf("publishing message %s to topic %s: %v", subject, topic, err)
	}

	return nil
}
