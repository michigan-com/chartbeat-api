package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/michigan-com/chartbeat-api/lib"
	m "github.com/michigan-com/chartbeat-api/model"
)

type TrafficSeries struct{}

var trafficSeriesEndpoint = "historical/traffic/series/"

func (t TrafficSeries) Fetch(domains []string, apiKey string) m.Snapshot {
	log.Info("Fetching traffic series...")
	queryParams := url.Values{}
	queryParams.Set("apikey", apiKey)
	queryParams.Set("limit", "100")

	var start int
	var end int
	var frequency int
	var wait sync.WaitGroup
	trafficSeriesChannel := make(chan *m.Traffic, len(domains))
	for _, domain := range domains {
		queryParams.Set("host", domain)
		url := fmt.Sprintf("%s/%s?%s", ApiRoot, trafficSeriesEndpoint, queryParams.Encode())
		wait.Add(1)
		go func(url string) {
			defer wait.Done()
			resp, err := fetchTrafficSeries(url)
			if err != nil {
				return
			}

			traffic := &m.Traffic{}
			traffic.Source = resp.Data.Source

			series := resp.GetSeries()
			if series != nil {
				traffic.Visits = series.Series.People
			}

			start = resp.Data.Start
			end = resp.Data.End
			frequency = resp.Data.Frequency

			trafficSeriesChannel <- traffic
		}(url)
	}
	wait.Wait()
	close(trafficSeriesChannel)

	log.Info("...fetched traffic series")

	trafficSeriesArray := make([]*m.Traffic, 0, len(domains))
	for traffic := range trafficSeriesChannel {
		trafficSeriesArray = append(trafficSeriesArray, traffic)
	}

	snapshot := m.TrafficSeriesSnapshot{}
	snapshot.Start = start
	snapshot.End = end
	snapshot.Frequency = frequency
	snapshot.Traffic = trafficSeriesArray
	snapshot.Created_at = time.Now()
	return snapshot
}

func fetchTrafficSeries(url string) (*m.TrafficSeriesIn, error) {

	resp, err := http.Get(url)
	if err != nil {
		log.Errorf("\n\n\tFailed to get traffic series url %s: \n\n\t%v\n", url, err)
		return nil, err
	}
	defer resp.Body.Close()

	// They use variable keys, based on the host, for each return
	// Decode, once for the data with static keys, and one for the data with variable keys
	trafficSeriesIn := &m.TrafficSeriesIn{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(trafficSeriesIn)
	if err != nil {
		log.Errorf("\n\n\tFailed to decode url (using TrafficSeriesIn): %s\n\n\t%v\n", url, err)
		return nil, err
	}

	trafficSeriesIn.Data.Source, _ = lib.GetHostFromParamsAndStrip(url)

	return trafficSeriesIn, nil
}
