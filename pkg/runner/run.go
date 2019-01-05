package runner

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/dominicbreuker/job_runner/pkg/awsclient"
	"github.com/dominicbreuker/job_runner/pkg/awsclient/sns"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

var log = zlog.With().Logger()
var snsAPI = awsclient.GetSNS
var waitTime = 500 * time.Millisecond

type RunInput struct {
	JobName string
	CMD     string

	SuccessTopic string // SNS topic for success notifications
	ErrorTopic   string // SNS topic for error notifications
	HistoryTable string // DynamoDB table for job history
}

func Run(cfg *RunInput) error {
	log = log.With().Str("cmd", cfg.CMD).Str("job", cfg.JobName).Logger()

	waitStatus, err := execute(cfg.CMD, &log)
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

	wg := sync.WaitGroup{}

	stdoutLog := log.With().Str("fd", "stdout").Logger()
	go logOutput(stdout, &stdoutLog, &wg)

	stderrLog := log.With().Str("fd", "stderr").Logger()
	go logOutput(stderr, &stderrLog, &wg)

	wg.Wait()
	time.Sleep(waitTime) // TODO: find out why we get 'file already closed' without...

	if err := command.Wait(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus := exitError.Sys().(syscall.WaitStatus)
			return int(waitStatus), nil
		}
		return -1, fmt.Errorf("waiting for command %s: %v", cmd, err)
	}

	return 0, nil
}

func logOutput(r io.Reader, log *zerolog.Logger, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	reader := bufio.NewReaderSize(r, 65536)

	for {
		line, isPrefix, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				return
			}

			log.Error().Err(err).Msg("Error reading shell output from job runner")
			return
		}

		if isPrefix {
			line = append(line, byte('.'), byte('.'), byte('.'))
		}

		log.Info().Msg(string(line))
	}
}

func publishFinalStatus(success bool, cfg *RunInput) error {
	subject := fmt.Sprintf("Job '%s': success = %t", cfg.JobName, success)
	message := "..."
	topic := cfg.SuccessTopic
	if !success {
		topic = cfg.ErrorTopic
	}

	if err := sns.GetClient(snsAPI()).Publish(subject, message, topic); err != nil {
		return fmt.Errorf("publishing final status notification: %v", err)
	}

	return nil
}
