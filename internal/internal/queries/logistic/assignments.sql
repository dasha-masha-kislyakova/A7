-- name: ListAssignments :many
SELECT id,shipment_id,application_external_id,weight_kg,volume_m3,created_at FROM assignments ORDER BY id DESC;

-- name: SumLoadByShipment :one
SELECT COALESCE(SUM(weight_kg),0), COALESCE(SUM(volume_m3),0) FROM assignments WHERE shipment_id=$1;

-- name: InsertAssignment :exec
INSERT INTO assignments(shipment_id,application_external_id,weight_kg,volume_m3,created_at,pickup_point_external_id,dropoff_point_external_id)
VALUES ($1,$2,$3,$4,now(),$5,$6);
