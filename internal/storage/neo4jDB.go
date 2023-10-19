package storage

import (
	"context"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func BootstrapNeo4j(dbUri string, dbName string, dbUser string, dbPassword string, timeout time.Duration) (neo4j.DriverWithContext, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	driver, err := neo4j.NewDriverWithContext(
		dbUri,
		neo4j.BasicAuth(dbUser, dbPassword, ""))

	if err != nil {
		return nil, err
	}
	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		return nil, err
	}

	println("Connected to Neo4j ... :)")
	return driver, nil
}

func CloseNeo4j(driver neo4j.DriverWithContext) error {
	return driver.Close(context.Background())
}
