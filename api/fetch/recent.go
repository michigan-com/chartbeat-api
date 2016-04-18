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

type Recent struct{}

var recentEndpoint = "live/recent/v3"

func (r Recent) Fetch(domains []string, apiKey string) m.Snapshot {
	log.Info("Fetching recent...")
	urlParams := url.Values{}
	urlParams.Set("apikey", apiKey)

	var wait sync.WaitGroup
	recentChannel := make(chan *m.RecentResp, len(domains))
	for _, domain := range domains {
		urlParams.Set("host", domain)
		url := fmt.Sprintf("%s/%s?%s", ApiRoot, recentEndpoint, urlParams.Encode())
		wait.Add(1)

		go func(url string) {
			defer wait.Done()
			resp, err := fetchRecent(url)
			if err != nil {
				return
			}
			recentChannel <- resp
		}(url)
	}
	wait.Wait()
	close(recentChannel)

	log.Info("...recent fetched")

	recents := make([]*m.RecentResp, 0, len(domains))
	for recent := range recentChannel {
		recents = append(recents, recent)
	}

	snapshot := m.RecentSnapshot{}
	snapshot.Created_at = time.Now()
	snapshot.Recents = recents
	return snapshot
}

func fetchRecent(url string) (*m.RecentResp, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Errorf("\n\n\tFailed to fetch url %s:\n\n\t%v\n", url, err)
		return nil, err
	}
	defer resp.Body.Close()

	recentArray := make([]m.Recent, 0, 100)
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&recentArray)
	if err != nil {
		log.Errorf("\n\n\tFailed to decode url %s:\n\n\t%v\n", url, err)
		return nil, err
	}

	recent := &m.RecentResp{}
	recent.Recents = recentArray
	recent.Source, _ = lib.GetHostFromParams(url)

	return recent, nil
}
