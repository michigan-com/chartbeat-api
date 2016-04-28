package commands

import (
	"sync"
	"time"
	"fmt"
	"net/http"

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
	var envConfig, _ = config.GetEnv()
	var apiConfig, _ = config.GetApiConfig()
	var wait sync.WaitGroup
	var session *mgo.Session
	if envConfig.MongoUri != "" {
		session = lib.DBConnect(envConfig.MongoUri)
		defer session.Close()
	}

	for {
		var startTime time.Time = time.Now()

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

		if envConfig.GnapiDomain != "" {
			// Now hit all the necessary endpoints to update mapi
			// TODO do something different
			mapiUrls := []string{"popular", "quickstats", "topgeo", "referrers", "recent", "traffic-series"}
			for _, url := range mapiUrls {
				wait.Add(1)

				mapiUrl := fmt.Sprintf("%s/%s/", envConfig.GnapiDomain, url)
				log.Info(mapiUrl)
				go func (mapiUrl string) {
					defer wait.Done()
					resp, err := http.Get(mapiUrl)
					if err != nil {
						return
					}
					defer resp.Body.Close()
				}(mapiUrl)
			}
			wait.Wait()
		} else {
			log.Info("No Gnapi domain specified, cannot update gnapi instance")
		}

		endTime := time.Now()
		log.Infof("Elapsed time: %v", endTime.Sub(startTime))

		if loop > 0 {
			log.Infof("Sleeping for %d seconds...", loop)
			time.Sleep(time.Duration(loop) * time.Second)
			log.Info("...and now I'm awake!")
		} else {
			break
		}
	}
}
