package main

import (
	"strings"
	"time"

	"github.com/pkg/errors"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/michigan-com/chartbeat-api/chartbeat"
)

type ReferrersSnapshot struct {
	Id         bson.ObjectId `bson:"_id,omitempty"`
	Created_at time.Time     `bson:"created_at"`
	Referrers  []*Referrers  `bson:"referrers"`
}

type Referrers struct {
	Source    string `bson:"source"`
	Domain    string `bson:"domain"`
	Referrers bson.M `bson:"referrers"`
}

func saveReferrers(referrers map[string]*chartbeat.Referrers, db *mgo.Database) error {
	// Sanity check, for when API calls fail
	var r = ReferrersSnapshot{}
	realtimeCollection := db.C("Referrers")
	historyCollection := db.C("ReferrerHistory")

	for domain, ref := range referrers {
		var refs = Referrers{
			domain,
			strings.Replace(domain, ".com", "", 1),
			ref.Referrers,
		}

		r.Referrers = append(r.Referrers, &refs)
	}

	err := realtimeCollection.Insert(r)
	if err != nil {
		return errors.Wrap(err, "failed to save referrers snapshot")
	} else {
		err = removeOldSnapshots(realtimeCollection)
		if err != nil {
			return errors.Wrap(err, "failed to remove old recent snapshots")
		}
	}

	shortIndex := mgo.Index{
		Key:         []string{"created_at"},
		ExpireAfter: 30 * time.Second,
	}
	longIndex := mgo.Index{
		Key:         []string{"created_at"},
		ExpireAfter: 24 * 90 * time.Hour,
	}

	err = realtimeCollection.EnsureIndex(shortIndex)
	if err != nil {
		return errors.Wrap(err, "failed to create index on Referrers collection")
	}

	err = historyCollection.EnsureIndex(longIndex)
	if err != nil {
		return errors.Wrap(err, "failed to create index on ReferrerHistory collection")
	}

	err = realtimeCollection.Insert(r)
	if err != nil {
		return errors.Wrap(err, "failed to insert Referrers snapshot")
	}

	var latest ReferrersSnapshot

	fiveMinutesAgo := time.Now().Add(-time.Duration(5) * time.Minute)

	err = historyCollection.Find(bson.M{}).Sort("-created_at").One(&latest)
	if err != nil && err != mgo.ErrNotFound {
		return errors.Wrap(err, "failed to load latest doc from ReferrerHistory")
	}

	if err == mgo.ErrNotFound || latest.Created_at.Before(fiveMinutesAgo) {
		historyCollection.Insert(r)
	}

	return nil
}
