package eventstore

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	. "github.com/gucumber/gucumber"
	_ "github.com/mattn/go-oci8"
	"github.com/stretchr/testify/assert"
	. "github.com/xtracdev/goes/sample/testagg"
	"github.com/xtracdev/oraeventstore"
)

func init() {
	var eventStore *oraeventstore.OraEventStore
	var testAgg, testAgg2 *TestAgg
	var eventCount int

	Given(`^an evironment with event publishing disabled$`, func() {
		if len(configErrors) != 0 {
			assert.Fail(T, strings.Join(configErrors, "\n"))
			return
		}

		os.Setenv(oraeventstore.EventPublishEnvVar, "0")
	})

	When(`^I store an aggregate$`, func() {
		var err error
		var connectStr = fmt.Sprintf("%s/%s@//%s:%s/%s", DBUser, DBPassword, DBHost, DBPort, DBSvc)
		db, err := sql.Open("oci8", connectStr)
		if !assert.Nil(T, err) {
			return
		}
		defer db.Close()
		eventStore, err = oraeventstore.NewOraEventStore(db)
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

			log.Infof("Stored aggregate %s", testAgg.AggregateID)
		}
	})

	Then(`^no events are written to the publish table$`, func() {
		var connectStr = fmt.Sprintf("%s/%s@//%s:%s/%s", DBUser, DBPassword, DBHost, DBPort, DBSvc)
		db, err := sql.Open("oci8", connectStr)
		if err != nil {
			log.Infof("Error connecting to oracle: %s", err.Error())
		}
		if !assert.Nil(T, err) {
			return
		}
		defer db.Close()

		var count int = -1
		log.Infof("looking for publish of agg %s version %d", testAgg.AggregateID, testAgg.Version)
		err = db.QueryRow("select count(*) from t_aepb_publish where aggregate_id = :1 and version = :2", testAgg.AggregateID, testAgg.Version).Scan(&count)
		if err != nil {
			log.Infof("Error querying for published events: %s", err.Error())
		}

		assert.Nil(T, err)
		assert.Equal(T, 0, count)
	})

	Given(`^an environment with event publishing enabled$`, func() {
		if len(configErrors) != 0 {
			assert.Fail(T, strings.Join(configErrors, "\n"))
			return
		}

		os.Setenv(oraeventstore.EventPublishEnvVar, "1")
	})

	When(`^I store a new aggregate$`, func() {
		var connectStr = fmt.Sprintf("%s/%s@//%s:%s/%s", DBUser, DBPassword, DBHost, DBPort, DBSvc)
		db, err := sql.Open("oci8", connectStr)
		if !assert.Nil(T, err) {
			return
		}
		defer db.Close()
		eventStore, err := oraeventstore.NewOraEventStore(db)
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
		var connectStr = fmt.Sprintf("%s/%s@//%s:%s/%s", DBUser, DBPassword, DBHost, DBPort, DBSvc)
		db, err := sql.Open("oci8", connectStr)
		if !assert.Nil(T, err) {
			return
		}
		defer db.Close()

		var count int = -1
		err = db.QueryRow("select count(*) from t_aepb_publish where aggregate_id = :1 and version = :2", testAgg2.AggregateID, testAgg2.Version).Scan(&count)
		if err != nil {
			log.Infof("Error selecting from publish table: %s", err.Error())
		}
		assert.Nil(T, err)
		assert.Equal(T, 1, count)
	})

	When(`^I republish the events$`, func() {
		var connectStr = fmt.Sprintf("%s/%s@//%s:%s/%s", DBUser, DBPassword, DBHost, DBPort, DBSvc)
		db, err := sql.Open("oci8", connectStr)
		if !assert.Nil(T, err) {
			return
		}
		defer db.Close()

		var eventRecords int
		err = db.QueryRow("select  count(*) from t_aeev_events").Scan(&eventRecords)

		if !assert.Nil(T, err) {
			return
		}

		eventCount = eventRecords

		log.Info("republish the events")
		eventStore, err := oraeventstore.NewOraEventStore(db)
		err = eventStore.RepublishAllEvents()
		if !assert.Nil(T, err) {
			return
		}
	})

	Then(`^all the events are written to the publish table$`, func() {
		var connectStr = fmt.Sprintf("%s/%s@//%s:%s/%s", DBUser, DBPassword, DBHost, DBPort, DBSvc)
		db, err := sql.Open("oci8", connectStr)
		if !assert.Nil(T, err) {
			return
		}
		defer db.Close()

		var publishRecords int
		err = db.QueryRow("select  count(*) from t_aepb_publish").Scan(&publishRecords)

		if !assert.Nil(T, err) {
			return
		}

		assert.Equal(T, eventCount, publishRecords)
	})

}
