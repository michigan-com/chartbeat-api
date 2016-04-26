package model

import (
	"errors"
	"time"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type ReferrersSnapshot struct {
	Id         bson.ObjectId `bson:"_id,omitempty"`
	Created_at time.Time     `bson:"created_at"`
	Referrers  []*Referrers  `bson:"referrers"`
}

func (r ReferrersSnapshot) Save(session *mgo.Session) {
	// Sanity check, for when API calls fail
	if len(r.Referrers) == 0 {
		return
	}

	realtimeCollection := session.DB("").C("Referrers")
	historyCollection := session.DB("").C("ReferrerHistory")

	shortIndex := mgo.Index{
		Key:         []string{"created_at"},
		ExpireAfter: 30 * time.Second,
	}
	longIndex := mgo.Index{
		Key:         []string{"created_at"},
		ExpireAfter: 24 * 90 * time.Hour,
	}

	err := realtimeCollection.EnsureIndex(shortIndex)

	if err != nil {
		log.Errorf("Failed to ensure Index on Referrers collection: %v", err)
		return
	}

	err = historyCollection.EnsureIndex(longIndex)

	if err != nil {
		log.Errorf("Failed to ensure Index on ReferrerHistory collection: %v", err)
		return
	}

	err = realtimeCollection.Insert(r)
	if err != nil {
		log.Errorf("Failed to insert Referrers snapshot: %v", err)
		return
	}

	latest := &ReferrersSnapshot{}

	fiveMinutesAgo := time.Now().Add(-time.Duration(5) * time.Minute)

	err = historyCollection.Find(bson.M{}).Sort("-created_at").One(latest)

	if err == errors.New("not found") || latest.Created_at.Before(fiveMinutesAgo) {
		historyCollection.Insert(r)
	}
}

type Referrers struct {
	Source    string `json:"source"`
	Referrers bson.M `bson:"referrers" json:"referrers"`
}
