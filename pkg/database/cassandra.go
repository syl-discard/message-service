package database

import (
	logger "discard/message-service/pkg/models/logger"
	"time"

	"github.com/gocql/gocql"
)

func ConnectToDatabase(url string, keyspace string) *gocql.Session {
	var session *gocql.Session

	for {
		var err error
		cluster := gocql.NewCluster("message-db")
		cluster.Keyspace = "messages"
		session, err = cluster.CreateSession()

		if err == nil {
			logger.LOG.Println("Cassandra initialization done!")
			break
		}

		logger.WARN.Println("Failed to connect to Cassandra, retrying in 5 seconds...")
		logger.WARN.Println(err.Error())

		time.Sleep(5 * time.Second)
	}

	return session
}
