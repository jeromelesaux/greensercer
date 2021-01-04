package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Configuration struct {
	Port               string `json:"port"`
	DbUser             string `json:"rdsuser"`
	DbName             string `json:"rdsname"`
	DbPassword         string `json:"rdspassword"`
	DbEndpoint         string `json:"rdsendpoint"`
	AwsRegion          string `json:"awsregion"`
	AppleCertification string `json:"_"`
}

var (
	doOnce              sync.Once
	GlobalConfiguration = &Configuration{}
	AppleCertificatEnv  = "APPLE_CERTIFICAT"
	AppleEnvError       = fmt.Errorf("Apple Certificat path file Env is not set, export path file in " + AppleCertificatEnv)
)

func LoadConfiguration(configurationFilePath string) error {
	var err error
	appleCertificat := os.Getenv(AppleCertificatEnv)
	if appleCertificat == "" {
		return AppleEnvError
	}
	if configurationFilePath != "" {
		fmt.Fprintf(os.Stdout, "reading configuration from file [%s]\n", configurationFilePath)
		doOnce.Do(
			func() {
				f, err := os.Open(configurationFilePath)
				if err != nil {
					return
				}
				defer f.Close()
				err = json.NewDecoder(f).Decode(GlobalConfiguration)
				if err != nil {
					return
				}
				GlobalConfiguration.AppleCertification = appleCertificat
			})
		return err
	}
	fmt.Fprintf(os.Stdout, "configuration file is not set http port default value (8080)\n")
	doOnce.Do(
		func() {
			GlobalConfiguration = &Configuration{
				Port:               "8080",
				AppleCertification: appleCertificat,
			}
		})

	return nil
}
