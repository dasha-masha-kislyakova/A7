-- name: GetClientByInnName :one
SELECT id FROM clients WHERE inn = $1 AND name = $2;

-- name: InsertClient :one
INSERT INTO clients (name, inn, fio, tel, email)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;
