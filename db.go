package main

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	log "github.com/Sirupsen/logrus"

	"github.com/michigan-com/chartbeat-api/timerounding"
)

func SetupMongoSession(uri string) (*mgo.Session, error) {
	session, err := mgo.Dial(uri)
	if err != nil {
		return nil, err
	}

	session.SetMode(mgo.Monotonic, true)
	return session, nil
}

func removeOldSnapshots(col *mgo.Collection) error {
	var snapshot = bson.M{
		"_id": -1,
	}
	// Remove old snapshots
	col.Find(bson.M{}).
		Select(bson.M{"_id": 1}).
		Sort("-_id").
		One(&snapshot)

	_, err := col.RemoveAll(bson.M{
		"_id": bson.M{
			"$ne": snapshot["_id"],
		},
	})
	if err != nil {
		return errors.Wrap(err, "failed to remove old snapshots")
	}

	return nil
}

// insertHistoricalValues inserts new statistical values into the given collection.
//
// The collection contains a document per domain and per documentInterval, with "_id" being derived from these values.
// The specified values are pushed into a "values" key. Additionally, the collection has static "source", "domain" keys and
// tracks the minimal and maximum timestamp of the first and the last
func insertHistoricalValues(coll *mgo.Collection, values map[string]interface{}, now time.Time, documentInterval time.Duration, coalescenceInterval time.Duration) error {
	domains := make([]string, 0, len(values))
	for domain, _ := range values {
		domains = append(domains, domain)
	}

	skippedDomains, err := findDomainsWithSnapshotsAfter(coll, domains, now.Add(-coalescenceInterval))
	if err != nil {
		return err
	}
	log.Infof("QuickStats skipped domains = %v, of all domains = %v", skippedDomains, domains)

	bulk := coll.Bulk()
	bulk.Unordered()

	skippedDomainsSet := mapStringsToTrue(skippedDomains)
	for domain, value := range values {
		if !skippedDomainsSet[domain] {
			insertHistoricalValue(bulk, domain, now, value, documentInterval)
		}
	}

	_, err = bulk.Run()
	return err
}

func insertHistoricalValue(bulk *mgo.Bulk, domain string, now time.Time, value interface{}, documentInterval time.Duration) {
	if value == nil {
		return
	}

	id := fmt.Sprintf("%s-%s", domain, timerounding.FormatRoundedToDuration(now, documentInterval))

	bulk.Upsert(bson.M{"_id": id}, bson.M{
		"$setOnInsert": bson.M{
			"source": getSourceFromDomain(domain),
			"domain": domain,
		},
		"$min": bson.M{
			"tmstart": now,
		},
		"$max": bson.M{
			"tmend": now,
		},
		"$push": bson.M{
			"values": value,
		},
	})
}

func findDomainsWithSnapshotsAfter(coll *mgo.Collection, domains []string, threshold time.Time) ([]string, error) {
	var found []string
	err := coll.Find(bson.M{"domain": bson.M{"$in": domains}, "tmend": bson.M{"$gt": threshold}}).Distinct("domain", &found)
	return found, err
}
