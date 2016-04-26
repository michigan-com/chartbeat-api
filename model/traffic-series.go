package model

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type TrafficSeriesSnapshot struct {
	Id         bson.ObjectId `bson:"_id,omitempty"`
	Created_at time.Time     `bson:"created_at"`
	Start      int           `bson:"start"`
	End        int           `bson:"end"`
	Frequency  int           `bson:"frequency"`
	Traffic    []*Traffic    `bson:"sites"`
}

func (h TrafficSeriesSnapshot) Save(session *mgo.Session) {
	// Sanity check, for when API calls fail
	if len(h.Traffic) == 0 {
		return
	}

	collection := session.DB("").C("TrafficSeries")
	err := collection.Insert(h)

	if err != nil {
		log.Errorf("Failed to insert Historical snapshot: %v", err)
		return
	}

	// Capping collections for streaming , so no longer able to delete old snapshots
	removeOldSnapshots(collection)
}

type Traffic struct {
	Source string `bson:"source"`
	Visits []int  `bson:"visits"`
}

type TrafficSeriesIn struct {
	Data struct {
		Start     int    `json:"start"`
		End       int    `json:"end"`
		Frequency int    `json:"frequency"`
		Source    string `bson:"source"`

		Freep       *TrafficSeries `json:"freep.com"`
		DetroitNews *TrafficSeries `json:"detroitnews.com"`
		BattleCreek *TrafficSeries `json:"battlecreekenquirer.com"`
		Hometown    *TrafficSeries `json:"hometownlife.com"`
		Lansing     *TrafficSeries `json:"lansingstatejournal.com"`
		Livingston  *TrafficSeries `json:"livingstondaily.com"`
		Herald      *TrafficSeries `json:"thetimesherald.com"`

		// Usat
		UsaToday *TrafficSeries `json:"usatoday.com"`

		// Tennessean
		Tennessean *TrafficSeries `json:"tennessean.com"`

		// Central Ohio omg why are there so many sites help
		Mansfield        *TrafficSeries `json:"mansfieldnewsjournal.com"`
		Newark           *TrafficSeries `json:"newarkadvocate.com"`
		Zanesville       *TrafficSeries `json:"zanesvilletimesrecorder.com"`
		Chillicothe      *TrafficSeries `json:"chillicothegazette.com"`
		Lancaster        *TrafficSeries `json:"lancastereaglegazette.com"`
		Marion           *TrafficSeries `json:"marionstar.com"`
		TheNewsMessenger *TrafficSeries `json:"thenews-messenger.com"`
		Coshocton        *TrafficSeries `json:"coshoctontribune.com"`
		Bucyrus          *TrafficSeries `json:"bucyrustelegraphforum.com"`
		PortClinton      *TrafficSeries `json:"portclintonnewsherald.com"`

		// Central Ohio omg why are there so many sites help
		DesMoines    *TrafficSeries `json:"desmoinesregister.com"`
		PressCitizen *TrafficSeries `json:"press-citizen.com"`
		Juice        *TrafficSeries `json:"dmjuice.com"`
		HawkCentral  *TrafficSeries `json:"hawkcentral.com"`
	} `json:"data"`
}

func (h *TrafficSeriesIn) GetSeries() *TrafficSeries {
	if h.Data.Freep != nil {
		return h.Data.Freep
	} else if h.Data.DetroitNews != nil {
		return h.Data.DetroitNews
	} else if h.Data.BattleCreek != nil {
		return h.Data.BattleCreek
	} else if h.Data.Hometown != nil {
		return h.Data.Hometown
	} else if h.Data.Lansing != nil {
		return h.Data.Lansing
	} else if h.Data.Livingston != nil {
		return h.Data.Livingston
	} else if h.Data.Herald != nil {
		return h.Data.Herald
	} else if h.Data.UsaToday != nil {
		return h.Data.UsaToday
	} else if h.Data.Tennessean != nil {
		return h.Data.Tennessean
	} else if h.Data.Mansfield != nil {
		return h.Data.Mansfield
	} else if h.Data.Newark != nil {
		return h.Data.Newark
	} else if h.Data.Zanesville != nil {
		return h.Data.Zanesville
	} else if h.Data.Chillicothe != nil {
		return h.Data.Chillicothe
	} else if h.Data.Lancaster != nil {
		return h.Data.Lancaster
	} else if h.Data.Marion != nil {
		return h.Data.Marion
	} else if h.Data.TheNewsMessenger != nil {
		return h.Data.TheNewsMessenger
	} else if h.Data.Coshocton != nil {
		return h.Data.Coshocton
	} else if h.Data.Bucyrus != nil {
		return h.Data.Bucyrus
	} else if h.Data.PortClinton != nil {
		return h.Data.PortClinton
	} else if h.Data.DesMoines != nil {
		return h.Data.DesMoines
	} else if h.Data.PressCitizen != nil {
		return h.Data.PressCitizen
	} else if h.Data.Juice != nil {
		return h.Data.Juice
	} else if h.Data.HawkCentral != nil {
		return h.Data.HawkCentral
	}
	return nil
}

type TrafficSeries struct {
	Series *struct {
		People []int `json:"people"`
	} `json:"series"`
}
