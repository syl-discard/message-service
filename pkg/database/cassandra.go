package database

import (
	"discard/message-service/pkg/configuration"
	logger "discard/message-service/pkg/models/logger"
	"time"

	gocqlastra "github.com/datastax/gocql-astra"
	"github.com/gocql/gocql"
)

func ConnectToDatabase(configuration configuration.Configuration) *gocql.Session {
	var session *gocql.Session

	for {
		var err error
		var cluster *gocql.ClusterConfig

		if configuration.DatabaseSettings.Provider == "astra" {
			cluster, err = gocqlastra.NewClusterFromURL(
				gocqlastra.AstraAPIURL,
				configuration.DatabaseSettings.AstraId,
				configuration.DatabaseSettings.AstraToken,
				10*time.Second,
			)

			if err != nil {
				logger.WARN.Println("Failed to create a new Astra Cluster")
			}

			cluster.Keyspace = configuration.DatabaseSettings.Keyspace
			session, err = gocql.NewSession(*cluster)

			if err != nil {
				logger.WARN.Println("Failed to create a new Astra Session")
			}
		} else {
			cluster = gocql.NewCluster(configuration.DatabaseSettings.Url)
			cluster.Keyspace = configuration.DatabaseSettings.Keyspace
			session, err = cluster.CreateSession()

			if err != nil {
				logger.WARN.Println("Failed to create a new Cassandra Session")
			}
		}

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
