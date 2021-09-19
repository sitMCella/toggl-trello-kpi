// Package storage provides the services for storing and retrieving data from the database and CSV files.
package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/sitMCella/toggl-trello-kpi/configuration"
)

// PostgresqlConnection implements the PostgreSQL client.
type PostgresqlConnection struct {
	Db *sql.DB
}

// NewPostgresConnection creates a new connection to PostgreSQL.
func NewPostgresConnection(dbConfiguration configuration.DBConfiguration) (postgresqlConnection PostgresqlConnection, err error) {
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", dbConfiguration.Username, dbConfiguration.Password, dbConfiguration.Host, dbConfiguration.Port, dbConfiguration.Name)
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		return
	}
	db.SetMaxOpenConns(dbConfiguration.MaxOpenConnections)
	db.SetMaxIdleConns(dbConfiguration.MaxIdleConnections)
	db.SetConnMaxLifetime(time.Duration(dbConfiguration.ConnectionMaxLifeTimeInMinutes) * time.Minute)
	postgresqlConnection = PostgresqlConnection{
		Db: db,
	}
	return
}

// InitDB creates the "toggl_time" and "trello_card" tables if these don't exist.
func (pc PostgresqlConnection) InitDatabase() error {
	err := pc.createTogglTimeTable()
	if err != nil {
		return err
	}
	return pc.createTrelloCardTable()
}

// Close closes the PostgreSQL connection.
func (pc PostgresqlConnection) Close() error {
	return pc.Db.Close()
}

// GetDb retrieves the database connection.
func (pc PostgresqlConnection) GetDb() *sql.DB {
	return pc.Db
}

func (pc PostgresqlConnection) createTogglTimeTable() (err error) {
	tx, err := pc.Db.Begin()
	if err != nil {
		return
	}
	defer func() {
		switch err {
		case nil:
			sqlerr := tx.Commit()
			if err == nil {
				err = sqlerr
			}
		default:
			sqlerr := tx.Rollback()
			if err == nil {
				err = sqlerr
			}
		}
	}()
	sqlStmt := `CREATE TABLE IF NOT EXISTS toggl_time
				(
					id              varchar(255) NOT NULL,
					description     text NOT NULL,
					start           timestamp NOT NULL,
					stop            timestamp NOT NULL,
					duration        integer NOT NULL,
					billable        boolean NOT NULL,
					workspace_id    integer NOT NULL,
					project_id      integer NOT NULL,
					tags            varchar(255)[] NOT NULL DEFAULT array[]::varchar(255)[],
					trello_card_id  varchar(255) NOT NULL DEFAULT '',
					PRIMARY KEY(id)
				);`
	_, err = pc.Db.Exec(sqlStmt)
	return
}

func (pc PostgresqlConnection) createTrelloCardTable() (err error) {
	tx, err := pc.Db.Begin()
	if err != nil {
		return
	}
	defer func() {
		switch err {
		case nil:
			sqlerr := tx.Commit()
			if err == nil {
				err = sqlerr
			}
		default:
			sqlerr := tx.Rollback()
			if err == nil {
				err = sqlerr
			}
		}
	}()
	sqlStmt := `CREATE TABLE IF NOT EXISTS trello_card
				(
					id              varchar(255) NOT NULL,
					name            varchar(255) NOT NULL,
					closed          boolean NOT NULL,
					labels          varchar(255)[] NOT NULL DEFAULT array[]::varchar(255)[],
					project         varchar(255) NOT NULL DEFAULT '',
					customer        varchar(255) NOT NULL DEFAULT '',
					team            varchar(255) NOT NULL DEFAULT '',
					type            varchar(255) NOT NULL DEFAULT '',
					PRIMARY KEY(id)
				);`
	_, err = pc.Db.Exec(sqlStmt)
	return
}
