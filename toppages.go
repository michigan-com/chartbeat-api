package main

import (
	"sort"
	"strings"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/michigan-com/chartbeat-api/chartbeat"
	"github.com/michigan-com/chartbeat-api/lib"
	"github.com/pkg/errors"
)

/** Sorting stuff */
type ByVisits []*TopArticle

func (a ByVisits) Len() int           { return len(a) }
func (a ByVisits) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByVisits) Less(i, j int) bool { return a[i].Visits > a[j].Visits }

type newsfetchArticle struct {
	Id      bson.ObjectId `bson:"_id,omitempty" json:"_id"`
	Summary []string      `bson:"summary" json:"summary"`
}

/*
  Chartbeat Top pages snapshot that will be saved. Article objects from chartbeat api
  will be formatted as TopArticle objects
*/
type TopPagesSnapshot struct {
	ID        bson.ObjectId `bson:"_id,omitempty"`
	CreatedAt time.Time     `bson:"created_at"`
	Articles  []*TopArticle `bson:"articles"`
}

/*
  Formatted top article from the chartbeat toppages API
*/
type TopArticle struct {
	ID        bson.ObjectId           `bson:"_id,omitempty"`
	ArticleID int                     `bson:"article_id"`
	Headline  string                  `bson:"headline"`
	URL       string                  `bson:"url"`
	Authors   []string                `bson:"authors"`
	Source    string                  `bson:"source"`
	Domain    string                  `bson:"domain"`
	Sections  []string                `bson:"sections"`
	Visits    int                     `bson:"visits"`
	Loyalty   chartbeat.LoyaltyStats  `loyalty:"loyalty"`
	Stats     *chartbeat.ArticleStats `bson:"stats"`
}

func saveTopPages(topPages map[string]*chartbeat.TopPagesData, db *mgo.Database) error {
	now := time.Now()

	// Sanity check, for when API calls fail
	snapshotCollection := db.C("Toppages")

	snapshot := TopPagesSnapshot{
		CreatedAt: now,
	}

	for _, top := range topPages {
		snapshot.Articles = append(snapshot.Articles, formatTopPages(top)...)
	}

	sort.Sort(ByVisits(snapshot.Articles))

	err := snapshotCollection.Insert(snapshot)

	if err != nil {
		return errors.Wrap(err, "failed to save Top Pages snapshot")
	}

	// Capping collections for streaming , so no longer able to delete old snapshots
	err = removeOldSnapshots(snapshotCollection)
	if err != nil {
		return errors.Wrap(err, "failed to remove old Top Pages snapshots")
	}

	return saveTopPagesArticlesToScrape(&snapshot, db)
}

func saveTopPagesArticlesToScrape(t *TopPagesSnapshot, db *mgo.Database) error {
	articleCollection := db.C("Article")
	toScrape := make([]interface{}, 0, len(t.Articles))

	for _, topArticle := range t.Articles {
		var article newsfetchArticle
		articleId := topArticle.ArticleID
		articleIdQuery := bson.M{"article_id": articleId}
		articleCollection.Find(articleIdQuery).One(&article)

		if !article.Id.Valid() || len(article.Summary) != 3 {
			toScrape = append(toScrape, articleIdQuery)
			toScrape = append(toScrape, articleIdQuery)
		}
	}

	if len(toScrape) > 0 {
		bulk := db.C("ToScrape").Bulk()
		bulk.Upsert(toScrape...)
		_, err := bulk.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func formatTopPages(topPages *chartbeat.TopPagesData) []*TopArticle {
	topArticles := make([]*TopArticle, 0, 100)

	for _, page := range topPages.Pages {
		articleID := lib.GetArticleId(page.Path)
		domain := lib.GetDomainFromURL(page.Path)

		if articleID < 0 || lib.IsBlacklisted(page.Path) {
			continue
		}

		article := TopArticle{}
		article.ArticleID = articleID
		article.Headline = page.Title
		article.URL = page.Path
		article.Sections = page.Sections
		article.Visits = page.Stats.Visits   // TODO deprecate, use .Stats
		article.Loyalty = page.Stats.Loyalty // TODO deprecate, use .Stats
		article.Authors = lib.ParseAuthors(page.Authors)
		article.Domain = domain
		article.Source = strings.Replace(domain, ".com", "", 1) // TODO deprecate
		article.Stats = page.Stats

		topArticles = append(topArticles, &article)
	}

	return topArticles
}
