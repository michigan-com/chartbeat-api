package main

import (
	"strings"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/michigan-com/chartbeat-api/chartbeat"
	"github.com/michigan-com/newsfetch/lib"
	"github.com/pkg/errors"
)

type RecentSnapshot struct {
	Id        bson.ObjectId `bson:"_id,omitempty"`
	CreatedAt time.Time     `bson:"created_at"`
	Recents   []*Recents    `bson:"recents"`
}

type Recents struct {
	Source  string `bson:"source"`
	Domain  string `bson:"domain"`
	Recents []*chartbeat.Recent
}

func saveRecent(recents map[string]*chartbeat.RecentResp, db *mgo.Database) error {
	now := time.Now()

	snapshotCollection := db.C("Recent")

	snapshot := RecentSnapshot{
		CreatedAt: now,
	}

	for _, recent := range recents {
		snapshot.Recents = append(snapshot.Recents, formatRecents(recent))
	}

	err := snapshotCollection.Insert(snapshot)

	if err != nil {
		return errors.Wrap(err, "failed to save recent snapshot")
	}

	err = removeOldSnapshots(snapshotCollection)
	if err != nil {
		return errors.Wrap(err, "failed to remove old recent snapshots")
	}

	return nil
}

func formatRecents(recents *chartbeat.RecentResp) *Recents {

	parsedArticles := make([]*chartbeat.Recent, 0, 100)
	var domain string
	var source string
	for _, article := range recents.Recents {
		articleID := lib.GetArticleId(article.Url)

		if articleID > 0 {
			domain = article.Host
			article.Host = strings.Replace(article.Host, ".com", "", -1)
			parsedArticles = append(parsedArticles, &article)
		}
	}

	source = strings.Replace(domain, ".com", "", 1)

	var r = Recents{
		source,
		domain,
		parsedArticles,
	}

	return &r
}
