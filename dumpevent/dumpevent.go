package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/xtraclabs/oraeventstore"
	"os"
	"github.com/xtraclabs/goessample/testagg"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usaged: go run dumpevent.go <aggregate id>")
	}

	eventStore, err := oraeventstore.NewOraEventStore("replicantusr", "password", "xe.oracle.docker", "localhost", "1521")
	if err != nil {
		log.Fatalf("Error instantiating oracle event store")
	}

	events, err := eventStore.RetrieveEvents(os.Args[1])
	if err != nil {
		log.Fatalf("Error loading events: %s",err.Error())
	}

	testAgg := testagg.NewTestAggFromHistory(events)
	if testAgg == nil {
		log.Infof("No event history for the given aggregate")
		return
	}

	log.Infof("Your aggregate:")
	log.Infof("\taggregate id: %s", testAgg.Aggregate.ID)
	log.Infof("\tversion: %d",testAgg.Version)
	log.Infof("\tfoo: %s",testAgg.Foo)
	log.Infof("\tbar: %s",testAgg.Bar)
	log.Infof("\tbaz: %s",testAgg.Baz)
}
