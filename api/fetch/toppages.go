package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/michigan-com/chartbeat-api/lib"
	m "github.com/michigan-com/chartbeat-api/model"
)

type TopPages struct{}

var toppagesEndpoint = "live/toppages/v3"

func (t TopPages) Fetch(domains []string, apiKey string) m.Snapshot {
	log.Info("Fetching toppages...")
	var queryParams = url.Values{}
	queryParams.Set("all_platforms", "1")
	queryParams.Set("loyalty", "1")
	queryParams.Set("limit", "100")
	queryParams.Set("apikey", apiKey)

	var wait sync.WaitGroup
	topPagesChannel := make(chan *m.TopPagesData, len(domains))
	for _, domain := range domains {
		// TODO domain parsing for sections
		queryParams.Set("host", domain)
		url := fmt.Sprintf("%s/%s?%s", ApiRoot, toppagesEndpoint, queryParams.Encode())
		wait.Add(1)
		go func(url string) {
			defer wait.Done()
			topPages, err := fetchTopPages(url)
			if err != nil {
				return
			}
			topPagesChannel <- topPages
		}(url)
	}
	wait.Wait()
	close(topPagesChannel)

	log.Info("...toppages fetched")

	topPages := make([]*m.TopPagesData, 0, len(domains))
	for result := range topPagesChannel {
		topPages = append(topPages, result)
	}

	snapshot, _ := formatTopPages(topPages)
	return snapshot
}

/*
	Given a chartbeat url API, fetch it and return the dat
*/
func fetchTopPages(url string) (*m.TopPagesData, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Errorf("\n\n\tFailed to fetch Toppages url %v:\n\n\t\t%v", url, err)
		return nil, err
	}
	defer resp.Body.Close()

	topPages := &m.TopPagesData{}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(topPages)

	if err != nil {
		log.Errorf("\n\n\tFailed to decode Toppages url %v:\n\n\t\t%v", url, err)
		return nil, err
	}

	return topPages, nil
}

func formatTopPages(topPages []*m.TopPagesData) (m.Snapshot, error) {
	var wait sync.WaitGroup
	topPagesChannel := make(chan *m.TopArticle, len(topPages)*100)

	for _, topPageData := range topPages {
		wait.Add(1)
		go func(topPageData *m.TopPagesData) {
			defer wait.Done()

			for _, page := range topPageData.Pages {
				article := &m.TopArticle{}
				articleId := lib.GetArticleId(page.Path)
				parsedUrl, _ := url.Parse(page.Path)

				if articleId < 0 || lib.IsBlacklisted(page.Path) {
					continue
				}

				article.ArticleId = articleId
				article.Headline = page.Title
				article.Url = page.Path
				article.Sections = page.Sections
				article.Visits = page.Stats.Visits
				article.Loyalty = page.Stats.Loyalty
				article.Authors = lib.ParseAuthors(page.Authors)
				article.Source = parsedUrl.Host

				topPagesChannel <- article
			}
		}(topPageData)
	}
	wait.Wait()
	close(topPagesChannel)

	topArticles := make([]*m.TopArticle, 0, len(topPagesChannel))
	for article := range topPagesChannel {
		topArticles = append(topArticles, article)
	}
	sort.Sort(ByVisits(topArticles))

	snapshot := m.TopPagesSnapshot{}
	snapshot.Created_at = time.Now()
	snapshot.Articles = topArticles

	return snapshot, nil
}

/** Sorting stuff */
type ByVisits []*m.TopArticle

func (a ByVisits) Len() int           { return len(a) }
func (a ByVisits) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByVisits) Less(i, j int) bool { return a[i].Visits > a[j].Visits }
