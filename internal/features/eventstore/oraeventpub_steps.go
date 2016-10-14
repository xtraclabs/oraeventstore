package eventstore

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	. "github.com/gucumber/gucumber"
	"github.com/stretchr/testify/assert"
	. "github.com/xtracdev/goes/sample/testagg"
	"github.com/xtracdev/oraeventstore"
)

func init() {
	var eventStore *oraeventstore.OraEventStore
	var testAgg, testAgg2 *TestAgg

	Given(`^an evironment with event publishing disabled$`, func() {
		if len(configErrors) != 0 {
			assert.Fail(T, strings.Join(configErrors, "\n"))
			return
		}

		os.Setenv(oraeventstore.EventPublishEnvVar, "0")
	})

	When(`^I store an aggregate$`, func() {
		var err error
		connectStr := fmt.Sprintf("%s/%s@//%s:%s/%s", DBUser, DBPassword, DBHost, DBPort, DBSvc)
		db, err := sql.Open("oci8", connectStr)
		if err != nil {
			assert.Fail(T, "Error connecting to oracle")
			return
		}
		err = db.Ping()
		if err != nil {
			assert.Fail(T, "Error connecting to oracle")
			return
		}
		eventStore, err := oraeventstore.NewOraEventStore(db)
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
		log.Infof("looking for publish of agg %s version %d", testAgg.ID, testAgg.Version)
		err = db.QueryRow("select count(*) from publish where aggregate_id = :1 and version = :2", testAgg.ID, testAgg.Version).Scan(&count)
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
		connectStr := fmt.Sprintf("%s/%s@//%s:%s/%s", DBUser, DBPassword, DBHost, DBPort, DBSvc)
		db, err := sql.Open("oci8", connectStr)
		if err != nil {
			assert.Fail(T, "Error connecting to oracle")
			return
		}
		err = db.Ping()
		if err != nil {
			assert.Fail(T, "Error connecting to oracle")
			return
		}
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
		err = db.QueryRow("select count(*) from publish where aggregate_id = :1 and version = :2", testAgg2.ID, testAgg2.Version).Scan(&count)
		if err != nil {
			log.Infof("Error selecting from publish table: %s", err.Error())
		}
		assert.Nil(T, err)
		assert.Equal(T, 1, count)
	})

}
