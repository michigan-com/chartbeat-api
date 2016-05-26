package model

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	m "github.com/michigan-com/gannett-newsfetch/model"
)

/*
 * DATA GOING OUT
 */
type TopPagesSnapshot struct {
	Id         bson.ObjectId `bson:"_id,omitempty"`
	Created_at time.Time     `bson:"created_at"`
	Articles   []*TopArticle `bson:"articles"`
}

func (t TopPagesSnapshot) Save(session *mgo.Session) {
	// Sanity check, for when API calls fail
	if len(t.Articles) == 0 {
		return
	}

	snapshotCollection := session.DB("").C("Toppages")
	err := snapshotCollection.Insert(t)

	if err != nil {
		log.Errorf("Failed to insert TopPages snapshot: %v", err)
		return
	}

	// Capping collections for streaming , so no longer able to delete old snapshots
	removeOldSnapshots(snapshotCollection)

	t.SaveArticlesToScrape(session)
}

func (t TopPagesSnapshot) SaveArticlesToScrape(session *mgo.Session) {
	articleCollection := session.DB("").C("Article")
	toScrape := make([]interface{}, 0, len(t.Articles))

	log.Info("Determining if there's articles that we need to scrape")

	for _, topArticle := range t.Articles {
		article := &m.Article{}
		articleId := topArticle.ArticleId
		articleIdQuery := bson.M{"article_id": articleId}
		articleCollection.Find(articleIdQuery).One(article)

		if !article.Id.Valid() || len(article.Summary) != 3 {
			toScrape = append(toScrape, articleIdQuery)
			toScrape = append(toScrape, articleIdQuery)
		}
	}

	if len(toScrape) > 0 {
		bulk := session.DB("").C("ToScrape").Bulk()
		bulk.Upsert(toScrape...)
		_, err := bulk.Run()
		if err != nil {
			log.Error(err)
		}
	}
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
