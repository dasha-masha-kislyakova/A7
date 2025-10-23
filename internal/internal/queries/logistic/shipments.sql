-- name: InsertShipment :one
INSERT INTO shipments (max_weight_kg,max_volume_m3,departure_at,status,created_at)
VALUES ($1,$2,$3,'PLANNED',now()) RETURNING id;

-- name: ListShipments :many
SELECT id,max_weight_kg,max_volume_m3,departure_at,status,created_at FROM shipments ORDER BY id DESC;

-- name: GetShipment :one
SELECT id,max_weight_kg,max_volume_m3,departure_at,status,created_at FROM shipments WHERE id=$1;

-- name: UpdateShipmentDeparted :exec
UPDATE shipments SET status='DEPARTED' WHERE id=$1;

-- name: GetOpenShipments :many
SELECT s.id,s.max_weight_kg,s.max_volume_m3,s.departure_at, p1.point_id, p2.point_id 
FROM shipments s 
JOIN LATERAL (SELECT point_id FROM route_points WHERE shipment_id=s.id AND ordinal=1 LIMIT 1) p1 ON true
JOIN LATERAL (SELECT point_id FROM route_points WHERE shipment_id=s.id AND ordinal=2 LIMIT 1) p2 ON true
WHERE s.status='PLANNED';
