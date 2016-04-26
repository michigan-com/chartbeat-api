package model

import (
	"time"

	log "github.com/Sirupsen/logrus"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type RecentSnapshot struct {
	Id         bson.ObjectId `bson:"_id,omitempty"`
	Created_at time.Time     `bson:"created_at"`
	Recents    []*RecentResp `bson:"recents"`
}

func (r RecentSnapshot) Save(session *mgo.Session) {
	// Sanity check, for when API calls fail
	if len(r.Recents) == 0 {
		return
	}

	col := session.DB("").C("Recent")
	err := col.Insert(r)

	if err != nil {
		log.Errorf("Failed to insert Recent Snapshot: %v", err)
	}

	// Capping collections for streaming , so no longer able to delete old snapshots
	removeOldSnapshots(col)
}

type RecentResp struct {
	Source  string
	Recents []Recent
}

type Recent struct {
	Lat      float32 `json:"lat" bson:"lat"`
	Lng      float32 `json:"lng" bson:"lng"`
	Title    string  `json:"title" bson:"title"`
	Url      string  `json:"path" bson"url"`
	Host     string  `json:"domain" bson:"host"`
	Platform string
}
