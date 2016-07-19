package main

import (
	"strings"
	"time"

	"github.com/michigan-com/chartbeat-api/chartbeat"
	"github.com/pkg/errors"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type TopGeoSnapshot struct {
	Id        bson.ObjectId `bson:"_id,omitempty"`
	CreatedAt time.Time     `bson:"created_at"`
	Cities    []*TopGeo     `bson:"cities"`
}

type TopGeo struct {
	Domain string `bson:"domain"`
	Source string `bson:"source"`
	Cities bson.M `bson:"cities"`
}

func saveTopGeo(topGeoResps map[string]*chartbeat.TopGeoResp, db *mgo.Database) error {
	now := time.Now()
	var t = TopGeoSnapshot{
		CreatedAt: now,
	}
	collection := db.C("Topgeo")

	for domain, geo := range topGeoResps {
		var topGeo = TopGeo{
			domain,
			strings.Replace(domain, ".com", "", 1),
			geo.Geo.Cities,
		}
		t.Cities = append(t.Cities, &topGeo)
	}

	err := collection.Insert(t)

	if err != nil {
		return errors.Wrap(err, "failed to save TopGeo snapshot")
	}

	// Capping collections for streaming , so no longer able to delete old snapshots
	return removeOldSnapshots(collection)
}
