package model

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type TopGeoSnapshot struct {
	Id         bson.ObjectId `bson:"_id,omitempty"`
	Created_at time.Time     `bson:"created_at"`
	Cities     []*TopGeo     `bson:"cities"`
}

func (t TopGeoSnapshot) Save(session *mgo.Session) {
	// Sanity check, for when API calls fail
	if len(t.Cities) == 0 {
		return
	}

	collection := session.DB("").C("Topgeo")
	err := collection.Insert(t)

	if err != nil {
		log.Errorf("Failed to save Topgeo snapshot: %v", err)
	}

	// Capping collections for streaming , so no longer able to delete old snapshots
	removeOldSnapshots(collection)
}

type TopGeoResp struct {
	Geo TopGeo `bson:"geo:"`
}

type TopGeo struct {
	Source string `bson:"source"`
	Cities bson.M `bson:"cities"`
}
