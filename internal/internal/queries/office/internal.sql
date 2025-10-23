-- name: ListAvailableForRoute :many
SELECT a.id, g.total_weight, g.total_volume, a.pickup_point_id, a.dropoff_point_id, a.created_at
FROM applications a
JOIN cargos g ON g.id=a.cargo_id
WHERE a.status='NEW' AND a.pickup_point_id=$1 AND a.dropoff_point_id=$2 AND a.created_at <= $3
ORDER BY a.created_at ASC;

-- name: GetApplicationStatusBulk :many
SELECT id, status FROM applications WHERE id = ANY($1::bigint[]) ORDER BY id;
