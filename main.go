package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	lru "github.com/hashicorp/golang-lru"
	pb "github.com/mik3y/flightmon/proto"

	logging "github.com/ipfs/go-log"
)

var Version = "0.1.1"

var dumpHost = flag.String("dump_host", "localhost:30003", "dump1090 SBS1 stream address (required)")
var debug = flag.Bool("debug", false, "Enable debug output")
var showUI = flag.Bool("show_ui", true, "Enable the UI")
var maxAge = flag.Int("max_age", 60, "Stop showing aircraft older than this many seconds")

var logger = logging.Logger("main")

func getAgeOfUpdateInSeconds(update *pb.PositionUpdate) int {
	now := int64(time.Now().UnixNano() / int64(time.Millisecond))
	ageSeconds := (now - *update.Timestamp) / 1000
	return int(ageSeconds)
}

func startAging(cache *lru.Cache) {
	for {
		for _, key := range cache.Keys() {
			val, _ := cache.Peek(key)
			record := val.(*pb.PositionUpdate)
			age := getAgeOfUpdateInSeconds(record)
			if age >= *maxAge {
				logger.Debugf("Removing expired flight: %s", record.IcaoId)
				cache.Remove(key)
			}
		}
		time.Sleep(time.Second)
	}

}

func main() {
	flag.Parse()

	if *debug {
		logging.SetLogLevel("main", "debug")
	} else {
		if *showUI {
			logging.SetLogLevel("main", "warning")
		} else {
			logging.SetLogLevel("main", "info")
		}
	}

	trackedCache, err := lru.New(128)
	if err != nil {
		panic(err)
	}

	updates := make(chan *pb.PositionUpdate, 10)
	go startSbsTracking(trackedCache, updates)
	go startAging(trackedCache)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT)

	donec := make(chan struct{}, 1)

	if *showUI {
		go ShowUI(trackedCache, donec)
	}

	for {
		select {
		case <-stop:
			os.Exit(0)
		case <-donec:
			os.Exit(0)
		case record := <-updates:
			if record == nil {
				os.Exit(0)
			}
			logger.Debugf("PositionUpdate: %v", record)
		}
	}
}
