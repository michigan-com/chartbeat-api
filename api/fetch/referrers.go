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

type Referrers struct{}

var referrersEndpoint = "live/referrers/v3"

func (r Referrers) Fetch(domains []string, apiKey string) m.Snapshot {
	log.Info("Fetching referrers...")
	urlParams := url.Values{}
	urlParams.Set("apikey", apiKey)

	var wait sync.WaitGroup
	referrersChannel := make(chan *m.Referrers, len(domains))
	for _, domain := range domains {
		urlParams.Set("host", domain)
		url := fmt.Sprintf("%s/%s?%s", ApiRoot, referrersEndpoint, urlParams.Encode())
		wait.Add(1)

		go func(url string) {
			defer wait.Done()

			resp, err := fetchReferrers(url)
			if err != nil {
				return
			}
			referrersChannel <- resp
		}(url)
	}
	wait.Wait()
	close(referrersChannel)

	log.Info("...referrers fetched")

	referrers := make([]*m.Referrers, 0, len(domains))
	for referrer := range referrersChannel {
		referrers = append(referrers, referrer)
	}

	snapshot := m.ReferrersSnapshot{}
	snapshot.Created_at = time.Now()
	snapshot.Referrers = referrers
	return snapshot
}

func fetchReferrers(url string) (*m.Referrers, error) {

	resp, err := http.Get(url)
	if err != nil {
		log.Errorf("\n\n\tFailed to fetch referrers Url: %s:\n\n\t%v\n", url, err)
		return nil, err
	}
	defer resp.Body.Close()

	referrers := &m.Referrers{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(referrers)
	if err != nil {
		log.Errorf("\n\n\tFailed to decode referrers Url: %s\n\n\t%v\n", url, err)
		return nil, err
	}

	referrers.Source, _ = lib.GetHostFromParamsAndStrip(url)

	return referrers, nil
}
