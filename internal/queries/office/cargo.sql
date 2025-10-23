-- name: InsertCargo :one
INSERT INTO cargos (name, boxes, total_weight, total_volume, special_requirements)
VALUES ($1, $2, $3, $4, $5) RETURNING id;
