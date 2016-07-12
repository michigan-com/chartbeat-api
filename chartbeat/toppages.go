package chartbeat

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"

	"github.com/michigan-com/chartbeat-api/lib"
)

/** Sorting stuff */
type ByVisits []*TopArticle

func (a ByVisits) Len() int           { return len(a) }
func (a ByVisits) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByVisits) Less(i, j int) bool { return a[i].Visits > a[j].Visits }

var toppagesEndpoint = "live/toppages/v3"

type TopPagesSnapshot struct {
	Id         bson.ObjectId `bson:"_id,omitempty"`
	Created_at time.Time     `bson:"created_at"`
	Articles   []*TopArticle `bson:"articles"`
}

type TopArticle struct {
	Id        bson.ObjectId `bson:"_id,omitempty"`
	ArticleId int           `bson:"article_id"`
	Headline  string        `bson:"headline"`
	Url       string        `bson:"url"`
	Authors   []string      `bson:"authors"`
	Source    string        `bson:"source"`
	Sections  []string      `bson:"sections"`
	Visits    int           `bson:"visits"`
	Loyalty   LoyaltyStats  `json:"loyalty"`
}

type TopPagesData struct {
	Site  string
	Pages []*ArticleContent `json:"pages"`
}

type ArticleContent struct {
	Path     string        `json:"path"`
	Sections []string      `json:"sections"`
	Stats    *ArticleStats `json: "stats"`
	Title    string        `json:"title"`
	Authors  []string      `json:"authors"`
}

type ArticleStats struct {
	Visits  int          `json:"visits"`
	Loyalty LoyaltyStats `json:"loyalty"`
}

func (cl *Client) FetchTopPages(domains []string) (*TopPagesSnapshot, error) {
	log.Info("Fetching toppages...")
	var queryParams = url.Values{}
	queryParams.Set("all_platforms", "1")
	queryParams.Set("loyalty", "1")
	queryParams.Set("limit", "100")
	queryParams.Set("apikey", cl.APIKey)

	var wait sync.WaitGroup
	topPagesChannel := make(chan *TopPagesData, len(domains))
	for _, domain := range domains {
		// TODO domain parsing for sections
		queryParams.Set("host", domain)
		url := fmt.Sprintf("%s/%s?%s", apiRoot, toppagesEndpoint, queryParams.Encode())
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

	topPages := make([]*TopPagesData, 0, len(domains))
	for result := range topPagesChannel {
		topPages = append(topPages, result)
	}

	snapshot, _ := formatTopPages(topPages)
	return snapshot, nil
}

/*
	Given a chartbeat url API, fetch it and return the dat
*/
func fetchTopPages(url string) (*TopPagesData, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Errorf("\n\n\tFailed to fetch Toppages url %v:\n\n\t\t%v", url, err)
		return nil, err
	}
	defer resp.Body.Close()

	topPages := &TopPagesData{}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(topPages)

	if err != nil {
		log.Errorf("\n\n\tFailed to decode Toppages url %v:\n\n\t\t%v", url, err)
		return nil, err
	}

	if len(topPages.Pages) == 0 {
		err = errors.New(fmt.Sprintf("No top pages returned for %s", url))
		log.Errorf("\n\n\t%s\n\n\t", err)
		return nil, err
	}

	topPages.Site, _ = lib.GetHostFromParamsAndStrip(url)

	return topPages, nil
}

func formatTopPages(topPages []*TopPagesData) (*TopPagesSnapshot, error) {
	var wait sync.WaitGroup
	topPagesChannel := make(chan *TopArticle, len(topPages)*100)

	for _, topPageData := range topPages {
		wait.Add(1)
		go func(topPageData *TopPagesData) {
			defer wait.Done()

			for _, page := range topPageData.Pages {
				article := &TopArticle{}
				articleId := lib.GetArticleId(page.Path)

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
				article.Source = topPageData.Site

				topPagesChannel <- article
			}
		}(topPageData)
	}
	wait.Wait()
	close(topPagesChannel)

	topArticles := make([]*TopArticle, 0, len(topPagesChannel))
	for article := range topPagesChannel {
		topArticles = append(topArticles, article)
	}
	sort.Sort(ByVisits(topArticles))

	snapshot := &TopPagesSnapshot{}
	snapshot.Created_at = time.Now()
	snapshot.Articles = topArticles

	return snapshot, nil
}
