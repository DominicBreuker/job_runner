package cmd

import (
	"fmt"
	"os"

	"github.com/dominicbreuker/job_runner/pkg/config"
	_ "github.com/dominicbreuker/job_runner/pkg/initialize"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var awsRegion string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "job-runner",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&awsRegion, config.AWSRegionVar, "", "AWS region to use")
	viper.BindPFlag(config.AWSRegionVar, rootCmd.PersistentFlags().Lookup(config.AWSRegionVar))
}

// initConfig reads in ENV variables if set.
func initConfig() {
	viper.AutomaticEnv() // read in environment variables that match
}
