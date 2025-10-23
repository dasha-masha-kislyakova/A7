-- name: InsertRecipient :one
INSERT INTO recipients (name_company, address, fio, telephone)
VALUES ($1, $2, $3, $4) RETURNING id;
