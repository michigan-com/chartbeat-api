package main

import (
	"strings"
	"time"

	"github.com/michigan-com/chartbeat-api/chartbeat"
	"github.com/pkg/errors"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type TrafficSeriesSnapshot struct {
	Id        bson.ObjectId `bson:"_id,omitempty"`
	CreatedAt time.Time     `bson:"created_at"`
	Start     float64       `bson:"start"`
	End       float64       `bson:"end"`
	Frequency float64       `bson:"frequency"`
	Traffic   []*Traffic    `bson:"sites"`
}

type Traffic struct {
	Visits interface{} `bson:"visits"`
	Source string      `bson:"source"`
	Domain string      `bson:"domain"`
}

func saveTrafficSeries(trafficSeries map[string]*chartbeat.TrafficSeriesResp, db *mgo.Database) error {
	// Sanity check, for when API calls fail
	now := time.Now()
	collection := db.C("TrafficSeries")
	var t = TrafficSeriesSnapshot{
		CreatedAt: now,
	}

	for domain, traffic := range trafficSeries {
		t.Start = traffic.Data["start"].(float64)
		t.End = traffic.Data["end"].(float64)
		t.Frequency = traffic.Data["frequency"].(float64)

		trafficSeries := traffic.Data[domain].(map[string]interface{})["series"].(map[string]interface{})["people"]
		trafficVisits := Traffic{
			Visits: trafficSeries,
			Domain: domain,
			Source: strings.Replace(domain, ".com", "", 1),
		}
		t.Traffic = append(t.Traffic, &trafficVisits)
	}

	err := collection.Insert(t)

	if err != nil {
		return errors.Wrap(err, "failed to save traffic series snapshot")
	}

	// Capping collections for streaming , so no longer able to delete old snapshots
	return removeOldSnapshots(collection)
}
