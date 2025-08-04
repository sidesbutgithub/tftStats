-- name: InsertUnit :exec
INSERT INTO units(unitName, starLevel, items, placement)
VALUES ($1, $2, $3, $4);


-- name: BulkInsertUnits :copyfrom
INSERT INTO units(unitName, starLevel, items, placement)
VALUES ($1, $2, $3, $4);
