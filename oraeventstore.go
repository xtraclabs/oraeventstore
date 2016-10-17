package oraeventstore

import (
	"database/sql"
	"errors"
	"os"

	log "github.com/Sirupsen/logrus"
	_ "github.com/mattn/go-oci8"
	"github.com/xtracdev/goes"
)

var (
	ErrConcurrency = errors.New("Concurrency Exception")
	ErrPayloadType = errors.New("Only payloads of type []byte are allowed")
	ErrEventInsert = errors.New("Error inserting record into events table")
	ErrPubInsert   = errors.New("Error inserting record into pub table")
)

const (
	EventPublishEnvVar = "ES_PUBLISH_EVENTS"
	insertSQL          = "insert into events (aggregate_id, version, typecode, payload) values (:1, :2, :3, :4)"
)

type OraEventStore struct {
	db      *sql.DB
	publish bool
}

func (ora *OraEventStore) GetMaxVersionForAggregate(aggId string) (*int, error) {
	row, err := ora.db.Query("select max(version) from events where aggregate_id = :1", aggId)
	if err != nil {
		return nil, err
	}

	defer row.Close()

	var max int
	row.Scan(&max)

	err = row.Err()
	if err != nil {
		return nil, err
	}

	return &max, nil
}

func InsertEventFromParts(db *sql.DB, aggId string, version int, typecode string, payload []byte) error {
	_, err := db.Exec("insert into events (aggregate_id, version, typecode, payload) values (:1, :2, :3, :4)",
		aggId, version, typecode, payload)
	return err
}

func (ora *OraEventStore) writeEvents(agg *goes.Aggregate) error {

	log.Debug("start transaction")
	tx, err := ora.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	log.Debug("prepare statement")
	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		return err
	}

	var pubStmt *sql.Stmt
	if ora.publish {
		log.Debug("create publish statement")
		var pubstmtErr error
		pubStmt, pubstmtErr = tx.Prepare("insert into publish (aggregate_id, version) values (:1, :2)")
		if pubstmtErr != nil {
			return pubstmtErr
		}
	}

	for _, e := range agg.Events {
		log.Debug("process event %v\n", e)
		eventBytes, ok := e.Payload.([]byte)
		if !ok {
			stmt.Close()
			return ErrPayloadType
		}

		log.Debug("execute statement")
		_, execerr := stmt.Exec(agg.ID, e.Version, e.TypeCode, eventBytes)
		if execerr != nil {
			stmt.Close()
			log.Warn(execerr.Error())
			return ErrEventInsert
		}

		if ora.publish {
			log.Debug("execute publish statement")
			_, puberr := pubStmt.Exec(agg.ID, e.Version)
			if puberr != nil {
				log.Warn(puberr.Error())
				return ErrPubInsert
			}
		}
	}

	stmt.Close()
	if pubStmt != nil {
		pubStmt.Close()
	}

	log.Debug("commit transaction")
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (ora *OraEventStore) StoreEvents(agg *goes.Aggregate) error {
	//Select max for the aggregate id
	max, err := ora.GetMaxVersionForAggregate(agg.ID)
	if err != nil {
		return err
	}

	//If the stored version is not smaller than the agg version then
	//its a concurrency exception. Note we'll have a null max if no record
	//exists
	if !(*max < agg.Version) {
		return ErrConcurrency
	}

	//Store the events
	return ora.writeEvents(agg)
}

func (ora *OraEventStore) RetrieveEvents(aggID string) ([]goes.Event, error) {
	var events []goes.Event

	//Select the events, ordered by version
	rows, err := ora.db.Query(`select version, typecode, payload from events where aggregate_id = :1 order by version`, aggID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var version int
	var typecode string
	var payload []byte

	for rows.Next() {
		rows.Scan(&version, &typecode, &payload)
		event := goes.Event{
			Source:   aggID,
			Version:  version,
			TypeCode: typecode,
			Payload:  payload,
		}

		events = append(events, event)

	}

	err = rows.Err()

	return events, err
}

func NewOraEventStore(db *sql.DB) (*OraEventStore, error) {
	log.Infof("Creating event store...")
	publishEvents := os.Getenv(EventPublishEnvVar)
	switch publishEvents {
	case "1":
		log.Info("Event store configured to write records to publish table")
	default:
		log.Info("Event store will not write records to publish table - export ",
			EventPublishEnvVar, "= 1 to enable writing to publish table")

	}

	return &OraEventStore{
		db:      db,
		publish: publishEvents == "1",
	}, nil
}
