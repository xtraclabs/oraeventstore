package eventstore

import (
	. "github.com/gucumber/gucumber"
	"os"
)

var DBUser string
var DBPassword string
var DBHost string
var DBPort string
var DBSvc string
var configErrors []string

func init() {
	Given(`^some tests to run$`, func() {
	})

	Then(`^the database connection configuration is read from the environment$`, func() {
	})

	GlobalContext.BeforeAll(func() {
		DBUser = os.Getenv("FEED_DB_USER")
		if DBUser == "" {
			configErrors = append(configErrors, "Configuration missing FEED_DB_USER env variable")
		}

		DBPassword = os.Getenv("FEED_DB_PASSWORD")
		if DBPassword == "" {
			configErrors = append(configErrors, "Configuration missing FEED_DB_PASSWORD env variable")
		}

		DBHost = os.Getenv("FEED_DB_HOST")
		if DBHost == "" {
			configErrors = append(configErrors, "Configuration missing FEED_DB_HOST env variable")
		}

		DBPort = os.Getenv("FEED_DB_PORT")
		if DBPort == "" {
			configErrors = append(configErrors, "Configuration missing FEED_DB_PORT env variable")
		}

		DBSvc = os.Getenv("FEED_DB_SVC")
		if DBSvc == "" {
			configErrors = append(configErrors, "Configuration missing FEED_DB_SVC env variable")
		}

	})

}
