package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/michigan-com/chartbeat-api/chartbeat"
)

const (
	ExitCodeErrProcessing   = 1
	ExitCodeErrDependencies = 2
	ExitCodeErrConfig       = 3
)

func main() {
	var loopSec int

	runtime.GOMAXPROCS(runtime.NumCPU())

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		invalidUsage("missing MONGO_URI env variable")
	}

	gnapiDomain := os.Getenv("GNAPI_DOMAIN")

	domstr := os.Getenv("DOMAINS")
	if len(domstr) == 0 {
		invalidUsage("missing DOMAINS env variable")
	}
	domains := strings.Split(domstr, ",")

	chartbeatAPIKey := os.Getenv("CHARTBEAT_API_KEY")
	if chartbeatAPIKey == "" {
		invalidUsage("missing CHARTBEAT_API_KEY env variable")
	}

	session, err := SetupMongoSession(mongoURI)
	if err != nil {
		log.Fatalf("Failed to connect to '%s': %v", mongoURI, err)
		os.Exit(ExitCodeErrDependencies)
	}
	defer session.Close()

	flag.IntVar(&loopSec, "l", 0, "Time in seconds to sleep before looping and hitting the apis again")
	flag.Parse()

	log.Infof(`Running Chartbeat Fetcher for domains: %v`, domains)

	chartb := &chartbeat.Client{
		APIKey: chartbeatAPIKey,
	}

	for {
		var startTime time.Time = time.Now()

		fetch(session, chartb, domains, gnapiDomain)

		endTime := time.Now()
		log.Infof("Elapsed time: %v", endTime.Sub(startTime))

		if loopSec > 0 {
			log.Infof("Sleeping for %d seconds...", loopSec)
			time.Sleep(time.Duration(loopSec) * time.Second)
			log.Info("...and now I'm awake!")
		} else {
			break
		}
	}
}

func invalidUsage(message string) {
	fmt.Fprintf(os.Stderr, "** %v\n", message)
	os.Exit(ExitCodeErrConfig)
}
