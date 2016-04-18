package api

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"

	"github.com/michigan-com/chartbeat-api/lib"
	m "github.com/michigan-com/chartbeat-api/model"
)

type QuickStats struct{}

var quickstatsEndpoint = "live/quickstats/v4"

func (q QuickStats) Fetch(domains []string, apiKey string) m.Snapshot {
	log.Info("Fetching quickstats...")
	var queryParams = url.Values{}
	queryParams.Set("all_platforms", "1")
	queryParams.Set("loyalty", "1")
	queryParams.Set("limit", "100")
	queryParams.Set("apikey", apiKey)

	var wait sync.WaitGroup
	quickStatsChannel := make(chan *m.QuickStats, len(domains))
	for _, domain := range domains {
		queryParams.Set("host", domain)
		url := fmt.Sprintf("%s/%s?%s", ApiRoot, quickstatsEndpoint, queryParams.Encode())
		wait.Add(1)
		go func(url string) {
			defer wait.Done()
			resp, err := fetchQuickstats(url)
			if err != nil {
				return
			}
			quickStatsChannel <- resp
		}(url)
	}

	wait.Wait()
	close(quickStatsChannel)

	log.Info("...quickstats fetched")

	quickStats := make([]*m.QuickStats, 0, len(domains))
	for quickStatsResp := range quickStatsChannel {
		quickStats = append(quickStats, quickStatsResp)
	}

	snapshot := m.QuickStatsSnapshot{}
	snapshot.Created_at = time.Now()
	snapshot.Stats = SortQuickStats(quickStats)
	return snapshot
}

func fetchQuickstats(url string) (*m.QuickStats, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Errorf("\n\n\tFailed to fetch Quickstats url %v:\n\n\t\t%v", url, err)
		return nil, err
	}
	defer resp.Body.Close()

	quickStatsResp := &m.QuickStatsResp{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(quickStatsResp)

	if err != nil {
		log.Errorf("\n\n\tFailed to decode Quickstats url %v:\n\n\t\t%v", url, err)
		return nil, err
	} else if quickStatsResp.Data == nil || quickStatsResp.Data.Stats == nil {
		errStr := "Quickstats Data or Stats are nil"
		log.Errorf("\n\n\t%s: %v", errStr, quickStatsResp.Data)
		return nil, errors.New(errStr)
	}

	quickStats := quickStatsResp.Data.Stats
	quickStats.Source, err = lib.GetHostFromParams(url)

	return quickStats, nil
}

type QuickStatsSort []*m.QuickStats

func (q QuickStatsSort) Len() int           { return len(q) }
func (q QuickStatsSort) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }
func (q QuickStatsSort) Less(i, j int) bool { return q[i].Visits > q[j].Visits }

func SortQuickStats(quickStats []*m.QuickStats) []*m.QuickStats {
	if len(quickStats) > 0 {
		sort.Sort(QuickStatsSort(quickStats))
	}
	return quickStats
}
