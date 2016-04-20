package config

import (
  "strings"

  "github.com/kelseyhightower/envconfig"
)

type EnvConfig struct {
  MongoUri        string `envconfig:"mongo_uri"`
  ChartbeatApiKey string `envconfig:"chartbeat_api_key"`
  Domains         string `envconfig:"domains"`
  GnapiDomain     string `envconfig:"gnapi_domain"`
}

type ApiConfig struct {
  ChartbeatApiKey string
  Domains     []string
}

func GetApiConfig() (apiConfig ApiConfig, err error) {
  var env EnvConfig
  err = envconfig.Process("chartbeat-api.api", &env)

  apiConfig.ChartbeatApiKey = env.ChartbeatApiKey
  apiConfig.Domains = strings.Split(env.Domains, ",")

  return apiConfig, err
}

/*
  Get the current configuration from environment variables
*/
func GetEnv() (env EnvConfig, err error) {
  err = envconfig.Process("chartbeat-api.global", &env)
  return env, err
}
