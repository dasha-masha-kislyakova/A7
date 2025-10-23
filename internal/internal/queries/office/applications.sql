-- name: InsertApplication :one
INSERT INTO applications (pickup_point_id, dropoff_point_id, client_id, cargo_id, recipient_id, status)
VALUES ($1,$2,$3,$4,$5,'NEW') RETURNING id;

-- name: UpdateApplicationStatus :exec
UPDATE applications
SET status=$2,
    accepted_at = CASE WHEN $2='IN_WORK' THEN now() ELSE accepted_at END,
    dispatched_at = CASE WHEN $2='IN_TRANSIT' THEN now() ELSE dispatched_at END,
    delivered_at = CASE WHEN $2='DELIVERED' THEN now() ELSE delivered_at END
WHERE id=$1;

-- name: ListApplications :many
SELECT id, pickup_point_id, dropoff_point_id, status, created_at
FROM applications
ORDER BY id DESC;

-- name: ListApplicationsByStatus :many
SELECT id, pickup_point_id, dropoff_point_id, status, created_at
FROM applications
WHERE status=$1
ORDER BY id DESC;
