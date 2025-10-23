-- name: InsertLogisticPoint :one
INSERT INTO logistic_points (title,address,lat,lon,created_at)
VALUES ($1,$2,$3,$4,now()) RETURNING id;

-- name: ListLogisticPoints :many
SELECT id,title,address,lat,lon,created_at FROM logistic_points ORDER BY id;
