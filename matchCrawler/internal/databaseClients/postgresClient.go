package databaseClients

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/sidesbutgithub/tftStats/matchCrawler/internal/database"
)

type PostgresDB struct {
	Client            *database.Queries
	Context           context.Context
	initialConnection *pgx.Conn
}

func (db *PostgresDB) ConnectPostgres(postgresURI string) error {
	db.Context = context.Background()
	conn, err := pgx.Connect(db.Context, postgresURI)
	if err != nil {
		return err
	}
	db.initialConnection = conn
	db.Client = database.New(conn)
	return nil
}

func (db *PostgresDB) InsertUnit(unitName string, starLevel int16, items []string, placement int16) error {
	return db.Client.InsertUnit(db.Context, database.InsertUnitParams{
		Unitname:  unitName,
		Starlevel: starLevel,
		Items:     items,
		Placement: placement,
	})
}

func (db *PostgresDB) CloseConn() error {
	return db.initialConnection.Close(db.Context)
}
