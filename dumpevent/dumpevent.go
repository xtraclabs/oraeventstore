package main

//This is a little test utility used for development - updated is with your credentials and import the
//appropriate event sourced entities to dump them


import (
	log "github.com/Sirupsen/logrus"
	"github.com/xtracdev/goes/sample/testagg"
	"github.com/xtracdev/oraeventstore"
	"os"
	"fmt"
	"database/sql"
)

func fatal(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usaged: go run dumpevent.go <aggregate id>")
	}

	var connectStr = fmt.Sprintf("%s/%s@//%s:%s/%s", "replicantusr", "password", "localhost", "1521", "xe.oracle.docker")
	db, err := sql.Open("oci8", connectStr)
	fatal(err)

	eventStore, err := oraeventstore.NewOraEventStore(db)
	fatal(err)

	events, err := eventStore.RetrieveEvents(os.Args[1])
	fatal(err)

	testAgg := testagg.NewTestAggFromHistory(events)
	if testAgg == nil {
		log.Infof("No event history for the given aggregate")
		return
	}

	log.Infof("Your aggregate:")
	log.Infof("\taggregate id: %s", testAgg.Aggregate.AggregateID)
	log.Infof("\tversion: %d", testAgg.Version)
	log.Infof("\tfoo: %s", testAgg.Foo)
	log.Infof("\tbar: %s", testAgg.Bar)
	log.Infof("\tbaz: %s", testAgg.Baz)
}
