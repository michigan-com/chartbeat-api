package api

import (
  m "github.com/michigan-com/chartbeat-api/model"
)

type ChartbeatApi interface {
  // Domains, api key
  Fetch([]string, string) m.Snapshot
}