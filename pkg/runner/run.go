package runner

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"syscall"

	"github.com/dominicbreuker/job_runner/pkg/awsclient/sns"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

var log = zlog.With().Logger()

type RunInput struct {
	jobName string
	cmd     string

	successTopic string // SNS topic for success notifications
	errorTopic   string // SNS topic for error notifications
	historyTable string // DynamoDB table for job history
}

func Run(cfg *RunInput) error {
	log = log.With().Str("cmd", cfg.cmd).Str("job", cfg.jobName).Logger()

	waitStatus, err := execute(cfg.cmd, &log)
	if err != nil {
		return fmt.Errorf("executing command: %v", err)
	}

	success := waitStatus == 0
	if err := publishFinalStatus(success, cfg); err != nil {
		return fmt.Errorf("running job: %v", err)
	}

	return nil
}

func execute(cmd string, log *zerolog.Logger) (int, error) {
	command := exec.Command("/bin/sh", "-c", cmd)

	stdout, err := command.StdoutPipe()
	if err != nil {
		return -1, fmt.Errorf("attaching pipe to stdout: %v", err)
	}

	stderr, err := command.StderrPipe()
	if err != nil {
		return -1, fmt.Errorf("attaching pipe to stderr: %v", err)
	}

	if err := command.Start(); err != nil {
		return -1, fmt.Errorf("starting command %s: %v", cmd, err)
	}

	stdoutLog := log.With().Str("fd", "stdout").Logger()
	go logOutput(stdout, &stdoutLog)

	stderrLog := log.With().Str("fd", "stderr").Logger()
	go logOutput(stderr, &stderrLog)

	if err := command.Wait(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus := exitError.Sys().(syscall.WaitStatus)
			return int(waitStatus), nil
		}
		return -1, fmt.Errorf("waiting for command %s: %v", cmd, err)
	}

	return 0, nil
}

func logOutput(r io.Reader, log *zerolog.Logger) {
	reader := bufio.NewReaderSize(r, 65536)

	line, isPrefix, err := reader.ReadLine()
	if err != nil {
		log.Error().Err(err).Msg("Error reading shell output from job runner")
		return
	}
	if isPrefix {
		line = append(line, byte('.'), byte('.'), byte('.'))
	}

	log.Info().Msg(string(line))
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
