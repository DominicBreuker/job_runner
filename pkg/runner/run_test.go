package runner

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/rs/zerolog"
)

func TestRun(t *testing.T) {
	tests := []struct {
		runInput RunInput
		logs     string // part of what to find in the logs
		err      bool
	}{
		{
			runInput: RunInput{
				JobName:      "valid_job",
				CMD:          "echo abc",
				SuccessTopic: "arn:to:success:topic",
				ErrorTopic:   "arn:to:error:topic",
			},
			logs: "abc",
			err:  false,
		},
		{
			runInput: RunInput{
				JobName:      "job_with_invalid_command",
				CMD:          "non-existing-command",
				SuccessTopic: "arn:to:success:topic",
				ErrorTopic:   "arn:to:error:topic",
			},
			logs: "command not found",
			err:  false, // error during command execution is not an error here
		},
		{
			runInput: RunInput{
				JobName:    "valid_cmd_without_success_topic",
				CMD:        "echo abc",
				ErrorTopic: "arn:to:error:topic",
			},
			logs: "abc",
			err:  true, // no success topic specified
		},
		{
			runInput: RunInput{
				JobName:      "invalid_cmd_without_error_topic",
				CMD:          "non-existing-command",
				SuccessTopic: "arn:to:error:topic",
			},
			logs: "command not found",
			err:  true, // no error topic specified
		},
	}

	for _, tt := range tests {
		var logs []string

		r, w := io.Pipe()
		done := make(chan struct{})
		go func() {
			scanner := bufio.NewScanner(r)
			for scanner.Scan() {
				logs = append(logs, scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				t.Fatalf("Error getting test logs: %v", err)
			}
			done <- struct{}{}
		}()

		logBKP := log
		log = zerolog.New(w)
		defer func() { log = logBKP }()

		snsAPIBKP := snsAPI
		mockSNSClient := mockSNSClient{}
		snsAPI = func() snsiface.SNSAPI { return mockSNSClient }
		defer func() { snsAPI = snsAPIBKP }()

		err := Run(&tt.runInput)
		if (err != nil) != tt.err {
			t.Errorf("Unexpected error result for %s: got err=%t but want err=%t | err=%v",
				tt.runInput.JobName, (err != nil), tt.err, err)
		}
		w.Close()
		<-done // wait for logs to be read completely

		if !strings.Contains(strings.Join(logs, "\n"), tt.logs) {
			t.Errorf("Unexpected logs for job %s: wanted string '%s' inside logs but did not find. Logs =\n%s\n",
				tt.runInput.JobName, tt.logs, logs)
		}
	}
}

type mockSNSClient struct {
	messages []mockSNSMessage
	snsiface.SNSAPI
}

type mockSNSMessage struct {
	subject string
	message string
}

func (svc mockSNSClient) Publish(req *sns.PublishInput) (*sns.PublishOutput, error) {
	fmt.Println("PUBLISH")
	svc.messages = append(svc.messages, mockSNSMessage{
		subject: *req.Subject,
		message: *req.Message,
	})
	resp := &sns.PublishOutput{}
	return resp, nil
}
