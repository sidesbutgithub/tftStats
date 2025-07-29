#!/bin/bash

DBSTRING="host=db port=5432 password=password user=postgres dbname=postgres sslmode=disable"

sleep 5

cd migrations

goose postgres "$DBSTRING" up