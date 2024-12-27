package db

import (
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/swarajkumarsingh/status-sqs-consumer/conf"
	"github.com/swarajkumarsingh/turbo-deploy/functions/logger"
	sqltrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql"
	sqlxtrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/jmoiron/sqlx"
)

type manager struct {
	DBConn *sqlx.DB
}

var Mgr manager

func init() {
	var err error
	sqltrace.Register("postgres", &pq.Driver{}, sqltrace.WithServiceName(conf.DB_URL))
	database, err := sqlxtrace.Open("postgres", conf.DB_URL)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	err = database.Ping()
	if err != nil {
		log.Fatal("Error while pinging the DB:", err)
		panic(err)
	}

	maxOpenConn := 50
	ENV_PROD := "prod"
	ENV := "dev"
	if ENV == ENV_PROD {
		maxOpenConn = 800
	}
	database.SetMaxOpenConns(maxOpenConn)
	database.SetMaxIdleConns(50)
	database.SetConnMaxIdleTime(2 * time.Minute)
	database.SetConnMaxLifetime(5 * time.Minute)
	Mgr = manager{
		DBConn: database,
	}
	logger.Log.Println("Connected to DB successfully")
}
