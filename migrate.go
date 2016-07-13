package main

import (
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2"
)

func migrate(db *mgo.Database) error {
	err := db.C("PlatformStatsDaily").EnsureIndexKey("domain", "tmstart")
	if err != nil {
		return errors.Wrap(err, "PlatformStatsDaily.EnsureIndex")
	}

	err = db.C("PlatformStatsDaily").EnsureIndexKey("domain", "tmend")
	if err != nil {
		return errors.Wrap(err, "PlatformStatsDaily.EnsureIndex")
	}

	return nil
}
