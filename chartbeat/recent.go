package chartbeat

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
	"gopkg.in/mgo.v2/bson"

	"github.com/michigan-com/chartbeat-api/lib"
)

var recentEndpoint = "live/recent/v3"

type RecentSnapshot struct {
	Id         bson.ObjectId `bson:"_id,omitempty"`
	Created_at time.Time     `bson:"created_at"`
	Recents    []*RecentResp `bson:"recents"`
}
type RecentResp struct {
	Source  string
	Recents []Recent
}

type Recent struct {
	Lat      float32 `json:"lat" bson:"lat"`
	Lng      float32 `json:"lng" bson:"lng"`
	Title    string  `json:"title" bson:"title"`
	Url      string  `json:"path" bson"url"`
	Host     string  `json:"domain" bson:"host"`
	Platform string
}

func (cl *Client) FetchRecent(domains []string) (*RecentSnapshot, error) {
	log.Info("Fetching recent...")
	queryParams := url.Values{}
	queryParams.Set("apikey", cl.APIKey)
	queryParams.Set("limit", "100")

	var wait sync.WaitGroup
	recentChannel := make(chan *RecentResp, len(domains))
	for _, domain := range domains {
		queryParams.Set("host", domain)
		url := fmt.Sprintf("%s/%s?%s", apiRoot, recentEndpoint, queryParams.Encode())
		wait.Add(1)

		go func(url string) {
			defer wait.Done()
			recent, err := fetchRecent(url)
			if err != nil {
				return
			}
			parsedArticles := make([]Recent, 0, 100)
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

	recents := make([]*RecentResp, 0, len(domains))
	for recent := range recentChannel {
		recents = append(recents, recent)
	}

	snapshot := &RecentSnapshot{}
	snapshot.Created_at = time.Now()
	snapshot.Recents = recents
	return snapshot, nil
}

func fetchRecent(url string) (*RecentResp, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Errorf("\n\n\tFailed to fetch url %s:\n\n\t%v\n", url, err)
		return nil, err
	}
	defer resp.Body.Close()

	recentArray := make([]Recent, 0, 100)
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

	recent := &RecentResp{}
	recent.Recents = recentArray
	recent.Source, _ = lib.GetHostFromParamsAndStrip(url)

	return recent, nil
}
