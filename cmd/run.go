package cmd

import (
	"github.com/dominicbreuker/hackcli/pkg/initialize"
	"github.com/dominicbreuker/job_runner/pkg/config"
	"github.com/dominicbreuker/job_runner/pkg/runner"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var jobName string
var command string

// startCmd represents the start command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		initialize.All()
		log.Info().Msg("Initialization complete")

		cfg := &runner.RunInput{
			JobName: jobName,
			CMD:     command,
		}
		if err := runner.Run(cfg); err != nil {
			log.Fatal().Err(err).Msg("Error executing job")
		}
		log.Info().Msg("Execution successful")
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.PersistentFlags().StringVarP(&jobName, config.RunJobName, "j", "", "Name of the job to be run")
	viper.BindPFlag(config.RunJobName, runCmd.PersistentFlags().Lookup(config.RunJobName))

	runCmd.PersistentFlags().StringVarP(&command, config.RunCMD, "c", "", "Command to execute as a job")
	viper.BindPFlag(config.RunCMD, runCmd.PersistentFlags().Lookup(config.RunCMD))
}
