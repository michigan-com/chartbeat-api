package commands

import (
  "fmt"
  "strings"

  log "github.com/Sirupsen/logrus"
  "github.com/spf13/cobra"

  "github.com/michigan-com/chartbeat-api/config"
)

var (
  VERSION    string
  COMMITHASH string
)

var ChartbeatCommand = &cobra.Command{
  Use: "chartbeat",
  Short: "Hit Chartbeat API and save data",
  Run: runChartbeat, // see ./chartbeat.go
}

func Run(version, commit string) {
  VERSION = version
  COMMITHASH = commit
  AddCommands()
  PrepareEnvironment()
  ChartbeatCommand.Execute()
}

/*
  Add all necessary command line commands
*/
func AddCommands() {
  ChartbeatCommand.AddCommand()
}

/*
  Prepare the environemtn for newsfetch. Read in the env variables, doing some
  basic env var checking to make sure they're set.
*/
func PrepareEnvironment() {
  env, _ := config.GetEnv()

  domainsSplit := strings.Split(env.Domains, ",")
  log.Info(env)
  domains := make([]string, 0, len(domainsSplit))
  for _, domain := range domainsSplit {
    if domain != "" {
      domains = append(domains, domain)
    }
  }

  if len(domains) == 0 {
    log.Fatal("No domains input, please set the DOMAINS env variable")
  }

  log.Info(fmt.Sprintf(`

  Running Chartbeat API

    Site Codes: %v

  `, domainsSplit))
}
