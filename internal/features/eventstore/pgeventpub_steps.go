package eventstore

import (
	"database/sql"
	"fmt"
	log "github.com/Sirupsen/logrus"
	. "github.com/lsegal/gucumber"
	"github.com/stretchr/testify/assert"
	. "github.com/xtraclabs/goessample/testagg"
	"github.com/xtraclabs/oraeventstore"
	"os"
)

func init() {
	var eventStore *oraeventstore.OraEventStore
	var testAgg, testAgg2 *TestAgg
	var connectStr = fmt.Sprintf("%s/%s@//%s:%s/%s", "esusr", "password", "localhost", "1521", "xe.oracle.docker")

	Given(`^an evironment with event publishing disabled$`, func() {
		os.Setenv(oraeventstore.EventPublishEnvVar, "0")
	})

	When(`^I store an aggregate$`, func() {
		var err error
		eventStore, err = oraeventstore.NewPGEventStore("esusr", "password", "xe.oracle.docker", "localhost", "1521")
		if err != nil {
			log.Infof("Error connecting to oracle: %s", err.Error())
		}
		assert.NotNil(T, eventStore)
		assert.Nil(T, err)
		if assert.NotNil(T, eventStore) {
			var err error
			testAgg, err = NewTestAgg("new foo", "new bar", "new baz")
			assert.Nil(T, err)
			assert.NotNil(T, testAgg)

			err = testAgg.Store(eventStore)
			if err != nil {
				log.Infof("Error storing aggregate: %s", err.Error())
			}

			log.Infof("Stored aggregate %s", testAgg.ID)
		}
	})

	Then(`^no events are written to the publish table$`, func() {

		log.Infof("Connect using %s", connectStr)
		db, err := sql.Open("oci8", connectStr)
		if err != nil {
			log.Infof("Error connecting to oracle: %s", err.Error())
		}
		if !assert.Nil(T, err) {
			return
		}
		defer db.Close()

		var count int = -1
		log.Infof("looking for publish of agg %s version %d", testAgg.ID, testAgg.Version)
		err = db.QueryRow("select count(*) from publish where aggregate_id = :1 and version = :2", testAgg.ID, testAgg.Version).Scan(&count)
		if err != nil {
			log.Infof("Error querying for published events: %s", err.Error())
		}

		assert.Nil(T, err)
		assert.Equal(T, 0, count)
	})

	Given(`^an environment with event publishing enabled$`, func() {
		os.Setenv(oraeventstore.EventPublishEnvVar, "1")
	})

	When(`^I store a new aggregate$`, func() {
		eventStore, err := oraeventstore.NewPGEventStore("esusr", "password", "xe.oracle.docker", "localhost", "1521")
		if err != nil {
			log.Infof("Error creating event store: %s", err.Error())
		}
		assert.Nil(T, err)

		if assert.NotNil(T, eventStore) {
			var err error
			testAgg2, err = NewTestAgg("new foo", "new bar", "new baz")
			assert.Nil(T, err)
			assert.NotNil(T, testAgg2)

			testAgg2.Store(eventStore)
		}
	})

	Then(`^the events are written to the publish table$`, func() {
		log.Infof("Connect using %s", connectStr)
		db, err := sql.Open("oci8", connectStr)
		if !assert.Nil(T, err) {
			return
		}
		defer db.Close()

		var count int = -1
		err = db.QueryRow("select count(*) from publish where aggregate_id = :1 and version = :2", testAgg2.ID, testAgg2.Version).Scan(&count)
		if err != nil {
			log.Infof("Error selecting from publish table: %s", err.Error())
		}
		assert.Nil(T, err)
		assert.Equal(T, 1, count)
	})

}
