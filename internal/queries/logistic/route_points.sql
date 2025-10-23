-- name: InsertRoutePoint :exec
INSERT INTO route_points (shipment_id,point_id,planned_arrive_at,ordinal)
VALUES ($1,$2,$3,$4);
