package database

import (
	"context"
	"database/sql"
)

type PostgresDB struct {
	Client  *Queries
	Context context.Context
}

func (db *PostgresDB) ConnectPostgres(postgresURI string) error {
	conn, err := sql.Open("postgres", postgresURI)
	if err != nil {
		return err
	}
	db.Client = New(conn)
	db.Context = context.Background()
	return nil
}

func (db *PostgresDB) InsertUnit(unitName string, starLevel int16, items []string, placement int16) (Unit, error) {
	return db.Client.InsertUnit(db.Context, InsertUnitParams{
		Unitname:  unitName,
		Starlevel: starLevel,
		Items:     items,
		Placement: placement,
	})
}
