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

type TopGeo struct{}

var topGeoEndpoint = "live/top_geo/v1"

func (t TopGeo) Fetch(domains []string, apiKey string) m.Snapshot {
	log.Info("Fetching topgeo...")
	queryParams := url.Values{}
	queryParams.Set("apikey", apiKey)
	queryParams.Set("limit", "100")

	var wait sync.WaitGroup
	topGeoChannel := make(chan *m.TopGeo, len(domains))
	for _, domain := range domains {
		queryParams.Set("host", domain)
		url := fmt.Sprintf("%s/%s?%s", ApiRoot, topGeoEndpoint, queryParams.Encode())
		wait.Add(1)

		go func(url string) {
			defer wait.Done()

			resp, err := fetchTopGeo(url)
			if err != nil {
				return
			}
			topGeoChannel <- resp
		}(url)
	}
	wait.Wait()
	close(topGeoChannel)

	log.Info("...fetched topgeo")

	topGeo := make([]*m.TopGeo, 0, len(domains))
	for geo := range topGeoChannel {
		topGeo = append(topGeo, geo)
	}

	snapshot := m.TopGeoSnapshot{}
	snapshot.Created_at = time.Now()
	snapshot.Cities = topGeo
	return snapshot
}

func fetchTopGeo(url string) (*m.TopGeo, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Errorf("\n\n\tFailed to fetch Topgeo url %s\n\n\t%v\n", url, err)
		return nil, err
	}
	defer resp.Body.Close()

	topGeoResp := &m.TopGeoResp{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(topGeoResp)
	if err != nil {
		log.Errorf("\n\n\tFailed to decode Topgeo url %s\n\n\t", url, err)
		return nil, err
	}

	topGeo := &topGeoResp.Geo
	topGeo.Source, _ = lib.GetHostFromParamsAndStrip(url)

	return topGeo, nil
}
