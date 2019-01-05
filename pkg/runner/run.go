package runner

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"syscall"

	"github.com/dominicbreuker/job_runner/pkg/awsclient"
	"github.com/dominicbreuker/job_runner/pkg/awsclient/sns"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

var log = zlog.With().Logger()
var snsAPI = awsclient.GetSNS

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

	stdoutR, stdoutW := io.Pipe()
	command.Stdout = stdoutW
	defer stdoutW.Close()

	stderrR, stderrW := io.Pipe()
	command.Stderr = stderrW
	defer stderrW.Close()

	if err := command.Start(); err != nil {
		return -1, fmt.Errorf("starting command %s: %v", cmd, err)
	}

	wg := sync.WaitGroup{}

	stdoutLogger := log.With().Str("fd", "stdout").Logger()
	go logOutput(stdoutR, &stdoutLogger, &wg)

	stderrLogger := log.With().Str("fd", "stderr").Logger()
	go logOutput(stderrR, &stderrLogger, &wg)

	wg.Wait()

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

	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		log.Info().Msg(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Error().Err(err).Msg("Error reading shell output from job runner")
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
