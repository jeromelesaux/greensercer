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
	AppleKey           string `json:"_"`
}

var (
	doOnce              sync.Once
	GlobalConfiguration = &Configuration{}
	AppleCertificatEnv  = "APPLE_CERTIFICAT"
	AppleKeyEnv         = "APPLE_KEY"
	AppleEnvError       = fmt.Errorf("Apple Key or Certificat Env is not set, export content file in " + AppleCertificatEnv + " and " + AppleKeyEnv)
)

func LoadConfiguration(configurationFilePath string) error {
	var err error
	appleKey := os.Getenv(AppleCertificatEnv)
	appleCertificat := os.Getenv(AppleKeyEnv)
	if appleKey == "" || appleCertificat == "" {
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
				GlobalConfiguration.AppleKey = appleKey
			})
		return err
	}
	fmt.Fprintf(os.Stdout, "configuration file is not set http port default value (8080)\n")
	doOnce.Do(
		func() {
			GlobalConfiguration = &Configuration{
				Port:               "8080",
				AppleCertification: appleCertificat,
				AppleKey:           appleKey,
			}
		})

	return nil
}
