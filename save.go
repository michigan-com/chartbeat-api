package main

import (
	"sort"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/michigan-com/chartbeat-api/chartbeat"
)

type QuickStatsSnapshot struct {
	Id         bson.ObjectId      `bson:"_id,omitempty"`
	Created_at time.Time          `bson:"created_at"`
	Stats      []*QuickStatsEntry `bson:"stats"`
}

type QuickStatsEntry struct {
	Source string `bson:source`

	chartbeat.QuickStats `bson:,inline`
}

type quickStatsSort []*QuickStatsEntry

func (q quickStatsSort) Len() int           { return len(q) }
func (q quickStatsSort) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }
func (q quickStatsSort) Less(i, j int) bool { return q[i].Visits > q[j].Visits }

func saveQuickStats(stats map[string]*chartbeat.QuickStats, db *mgo.Database) error {
	snapshot := &QuickStatsSnapshot{}
	snapshot.Created_at = time.Now()
	for domain, result := range stats {
		if result != nil {
			snapshot.Stats = append(snapshot.Stats, &QuickStatsEntry{getSourceFromDomain(domain), *result})
		}
	}
	sort.Sort(quickStatsSort(snapshot.Stats))

	log.Infof("Saving quickstats for %d domains", len(snapshot.Stats))

	coll := db.C("Quickstats")
	err := coll.Insert(snapshot)
	if err != nil {
		return errors.Wrap(err, "failed to save quick stats")
	}

	return removeOldSnapshots(coll)
}

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

func saveTopPages(t *chartbeat.TopPagesSnapshot, session *mgo.Session) error {
	// Sanity check, for when API calls fail
	if len(t.Articles) == 0 {
		return nil
	}

	snapshotCollection := session.DB("").C("Toppages")
	err := snapshotCollection.Insert(t)

	if err != nil {
		return errors.Wrap(err, "failed to save Top Pages snapshot")
	}

	// Capping collections for streaming , so no longer able to delete old snapshots
	err = removeOldSnapshots(snapshotCollection)
	if err != nil {
		return errors.Wrap(err, "failed to remove old Top Pages snapshots")
	}

	return saveTopPagesArticlesToScrape(t, session)
}

func saveTopPagesArticlesToScrape(t *chartbeat.TopPagesSnapshot, session *mgo.Session) error {
	articleCollection := session.DB("").C("Article")
	toScrape := make([]interface{}, 0, len(t.Articles))

	log.Info("Determining if there's articles that we need to scrape")

	for _, topArticle := range t.Articles {
		var article newsfetchArticle
		articleId := topArticle.ArticleId
		articleIdQuery := bson.M{"article_id": articleId}
		articleCollection.Find(articleIdQuery).One(&article)

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
			return err
		}
	}

	return nil
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
