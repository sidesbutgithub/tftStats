-- +goose Up
CREATE TABLE units(
    id serial PRIMARY KEY,
    unitName varchar(256) NOT NULL,
    starLevel smallint NOT NULL,
    items varchar(256)[],
    placement smallint NOT NULL
);


-- +goose Down
DROP TABLE units;