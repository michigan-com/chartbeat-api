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
	Source               string `bson:"source"`
	chartbeat.QuickStats `bson:",inline"`
}

type PlatformStatsValue struct {
	Time                    time.Time `bson:"tm"`
	chartbeat.PlatformStats `bson:",inline"`
}

type PlatformStatsSnapshot struct {
	ID string `bson:"_id"`

	Source string `bson:"source"`
	Domain string `bson:"domain"`

	StartTime time.Time `bson:"tmstart"`
	EndTime   time.Time `bson:"tmend"`

	Values []PlatformStatsValue `bson:"values"`
}

type quickStatsSort []*QuickStatsEntry

func (q quickStatsSort) Len() int           { return len(q) }
func (q quickStatsSort) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }
func (q quickStatsSort) Less(i, j int) bool { return q[i].Visits > q[j].Visits }

func saveQuickStats(stats map[string]*chartbeat.QuickStats, db *mgo.Database) error {
	now := time.Now()

	snapsColl := db.C("Quickstats")
	historicalColl := db.C("PlatformStatsDaily")

	snapshot := QuickStatsSnapshot{
		Created_at: now,
	}
	for domain, st := range stats {
		if st != nil {
			snapshot.Stats = append(snapshot.Stats, &QuickStatsEntry{getSourceFromDomain(domain), *st})
		}
	}
	sort.Sort(quickStatsSort(snapshot.Stats))

	log.Infof("Saving quickstats for %d domains", len(snapshot.Stats))

	err := snapsColl.Insert(snapshot)
	if err != nil {
		return errors.Wrap(err, "failed to save quick stats")
	}

	err = removeOldSnapshots(snapsColl)
	if err != nil {
		return err
	}

	platformValues := make(map[string]interface{}, len(stats))
	for domain, st := range stats {
		if st != nil {
			platformValues[domain] = PlatformStatsValue{now, st.PlatformEngaged}
		}
	}

	err = insertHistoricalValues(historicalColl, platformValues, now, 24*time.Hour, 5*time.Minute)
	if err != nil {
		return err
	}

	return nil
}
