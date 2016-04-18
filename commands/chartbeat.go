package commands

import (
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/mgo.v2"

	a "github.com/michigan-com/chartbeat-api/api"
	fetch "github.com/michigan-com/chartbeat-api/api/fetch"
	"github.com/michigan-com/chartbeat-api/config"
	lib "github.com/michigan-com/chartbeat-api/lib"
)

var endPoints = []a.ChartbeatApi{
	fetch.TopPages{},
	fetch.QuickStats{},
	fetch.Recent{},
	fetch.Referrers{},
	fetch.TopGeo{},
	fetch.TrafficSeries{},
}

func runChartbeat(command *cobra.Command, args []string) {
	var startTime time.Time = time.Now()
	var envConfig, _ = config.GetEnv()
	var apiConfig, _ = config.GetApiConfig()
	var wait sync.WaitGroup

	var session *mgo.Session
	if envConfig.MongoUri != "" {
		session = lib.DBConnect(envConfig.MongoUri)
		defer session.Close()
	}

	for _, endPoint := range endPoints {
		wait.Add(1)
		go func(endPoint a.ChartbeatApi) {
			sessionCopy := session.Copy()
			defer sessionCopy.Close()

			snapshot := endPoint.Fetch(apiConfig.Domains, apiConfig.ChartbeatApiKey)
			snapshot.Save(sessionCopy)
			wait.Done()
		}(endPoint)
	}
	wait.Wait()

	log.Info(startTime)
}
