package initialize

import (
	"github.com/dominicbreuker/job_runner/pkg/awsclient"
	"github.com/dominicbreuker/job_runner/pkg/config"
	"github.com/spf13/viper"
)

func initAWS() {
	awsclient.InitializeSession(viper.GetString(config.AWSRegionVar))
}
