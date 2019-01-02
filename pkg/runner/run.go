package runner

import (
	"fmt"

	"github.com/dominicbreuker/job_runner/pkg/awsclient/sns"
)

type RunInput struct {
	jobName string
	cmd     string

	successTopic string // SNS topic for success notifications
	errorTopic   string // SNS topic for error notifications
	historyTable string // DynamoDB table for job history
}

func Run(cfg *RunInput) error {

	success := true
	if err := publishFinalStatus(success, cfg); err != nil {
		return fmt.Errorf("running job: %v", err)
	}

	return nil
}

func publishFinalStatus(success bool, cfg *RunInput) error {
	subject := fmt.Sprintf("Job '%s': success = %t", cfg.jobName, success)
	message := "..."
	topic := cfg.successTopic
	if !success {
		topic = cfg.errorTopic
	}

	if err := sns.Publish(subject, message, topic); err != nil {
		return fmt.Errorf("publishing final status notification: %v", err)
	}

	return nil
}
