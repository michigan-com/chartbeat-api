package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
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
	queryParams := url.Values{}
	queryParams.Set("apikey", apiKey)
	queryParams.Set("limit", "100")

	var wait sync.WaitGroup
	recentChannel := make(chan *m.RecentResp, len(domains))
	for _, domain := range domains {
		queryParams.Set("host", domain)
		url := fmt.Sprintf("%s/%s?%s", ApiRoot, recentEndpoint, queryParams.Encode())
		wait.Add(1)

		go func(url string) {
			defer wait.Done()
			recent, err := fetchRecent(url)
			if err != nil {
				return
			}
			parsedArticles := make([]m.Recent, 0, 100)
			for _, article := range recent.Recents {
				articleId := lib.GetArticleId(article.Url)

				if articleId > 0 {
					article.Host = strings.Replace(article.Host, ".com", "", -1)
					parsedArticles = append(parsedArticles, article)
				}
			}
			recent.Recents = parsedArticles
			recentChannel <- recent
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

	if len(recentArray) == 0 {
		err = errors.New(fmt.Sprintf("Recents array is 0 for url %s", url))
		log.Errorf("\n\n\t%s\n\n", err)
		return nil, err
	}

	recent := &m.RecentResp{}
	recent.Recents = recentArray
	recent.Source, _ = lib.GetHostFromParamsAndStrip(url)

	return recent, nil
}
