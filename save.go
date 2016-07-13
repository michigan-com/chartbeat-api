package main

import (
	"time"

	"github.com/pkg/errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/michigan-com/chartbeat-api/chartbeat"
)

func saveRecent(r *chartbeat.RecentSnapshot, session *mgo.Session) error {
	// Sanity check, for when API calls fail
	if len(r.Recents) == 0 {
		return nil
	}

	col := session.DB("").C("Recent")
	err := col.Insert(r)

	if err != nil {
		return errors.Wrap(err, "failed to save Recent Snapshot")
	}

	return removeOldSnapshots(col)
}

func saveReferrers(r *chartbeat.ReferrersSnapshot, session *mgo.Session) error {
	// Sanity check, for when API calls fail
	if len(r.Referrers) == 0 {
		return nil
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

	var latest chartbeat.ReferrersSnapshot

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

func saveTopGeo(t *chartbeat.TopGeoSnapshot, session *mgo.Session) error {
	// Sanity check, for when API calls fail
	if len(t.Cities) == 0 {
		return nil
	}

	collection := session.DB("").C("Topgeo")
	err := collection.Insert(t)

	if err != nil {
		return errors.Wrap(err, "failed to save TopGeo snapshot")
	}

	// Capping collections for streaming , so no longer able to delete old snapshots
	return removeOldSnapshots(collection)
}

func saveTrafficSeries(h *chartbeat.TrafficSeriesSnapshot, session *mgo.Session) error {
	// Sanity check, for when API calls fail
	if len(h.Traffic) == 0 {
		return nil
	}

	collection := session.DB("").C("TrafficSeries")
	err := collection.Insert(h)

	if err != nil {
		return errors.Wrap(err, "failed to save traffic series snapshot")
	}

	// Capping collections for streaming , so no longer able to delete old snapshots
	return removeOldSnapshots(collection)
}

type newsfetchArticle struct {
	Id      bson.ObjectId `bson:"_id,omitempty" json:"_id"`
	Summary []string      `bson"summary" json:"summary"`
}
