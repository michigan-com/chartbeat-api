package model

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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
	snapshotCollection := session.DB("").C("Toppages")
	err := snapshotCollection.Insert(t)

	if err != nil {
		log.Errorf("Failed to insert TopPages snapshot: %v", err)
		return
	}

	removeOldSnapshots(snapshotCollection)
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

type LoyaltyStats struct {
	New       int `bson:"new"`
	Loyal     int `bson:"loyal"`
	Returning int `bson:"returning"`
}
