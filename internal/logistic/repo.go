package logistic

import (
	"context"
	"database/sql"
)

type Repo interface {
	EnsureSchema(ctx context.Context) error

	GetLogApp(ctx context.Context, id int64) (LogisticApplication, error)
	FindOrCreateLogApp(ctx context.Context, originalID int64) (int64, error)
	UpdateLogAppStatus(ctx context.Context, id int64, status ApplicationStatus) error
	ListLogApps(ctx context.Context, status *string) ([]LogisticApplication, error)

	InsertRoute(ctx context.Context, r CreateRouteRequest) (Route, error)
	InsertRoutePoint(ctx context.Context, routeID int64, p RoutePointInput) error
	AssignRouteApp(ctx context.Context, routeID, originalAppID, logAppID int64) error
	RouteAppPairs(ctx context.Context, routeID int64) ([][2]int64, error)
	SetRouteInProgress(ctx context.Context, routeID int64) error
}

type pgRepo struct{ db *sql.DB }

func NewRepo(db *sql.DB) Repo { return &pgRepo{db: db} }

func (r *pgRepo) EnsureSchema(ctx context.Context) error {
	ddl := `
CREATE TYPE IF NOT EXISTS application_status AS ENUM ('NEW','IN_PROGRESS','SHIPPED','DELIVERED','CANCELLED');
CREATE TYPE IF NOT EXISTS route_status AS ENUM ('DRAFT','SCHEDULED','IN_PROGRESS','COMPLETED');

CREATE TABLE IF NOT EXISTS logistics_applications (
  id BIGSERIAL PRIMARY KEY,
  original_application_id BIGINT NOT NULL,
  status application_status NOT NULL DEFAULT 'NEW',
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_log_app_original ON logistics_applications(original_application_id);

CREATE TABLE IF NOT EXISTS routes (
  id BIGSERIAL PRIMARY KEY,
  truck_volume NUMERIC(10,2) NOT NULL,
  truck_max_weight NUMERIC(10,2) NOT NULL,
  departure_date TIMESTAMP NOT NULL,
  status route_status NOT NULL DEFAULT 'DRAFT',
  created_by_manager_id BIGINT NOT NULL DEFAULT 1,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS route_points (
  id BIGSERIAL PRIMARY KEY,
  route_id BIGINT NOT NULL REFERENCES routes(id) ON DELETE CASCADE,
  logistics_point_id BIGINT NOT NULL,
  point_order INTEGER NOT NULL,
  planned_arrival TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS route_applications (
  id BIGSERIAL PRIMARY KEY,
  route_id BIGINT NOT NULL REFERENCES routes(id) ON DELETE CASCADE,
  application_id BIGINT NOT NULL,             -- office application id
  logistic_application_id BIGINT REFERENCES logistics_applications(id)
);

CREATE INDEX IF NOT EXISTS idx_route_points_route ON route_points(route_id);
CREATE INDEX IF NOT EXISTS idx_route_apps_route ON route_applications(route_id);
`
	_, err := r.db.ExecContext(ctx, ddl)
	return err
}

func (r *pgRepo) GetLogApp(ctx context.Context, id int64) (LogisticApplication, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, original_application_id, status, created_at, updated_at FROM logistics_applications WHERE id=$1`, id)
	var app LogisticApplication
	err := row.Scan(&app.ID, &app.OriginalApplicationID, &app.Status, &app.CreatedAt, &app.UpdatedAt)
	return app, err
}

func (r *pgRepo) FindOrCreateLogApp(ctx context.Context, originalID int64) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `SELECT id FROM logistics_applications WHERE original_application_id=$1`, originalID).Scan(&id)
	if err == sql.ErrNoRows {
		err = r.db.QueryRowContext(ctx, `INSERT INTO logistics_applications(original_application_id,status) VALUES($1,'NEW') RETURNING id`, originalID).Scan(&id)
	}
	return id, err
}

func (r *pgRepo) UpdateLogAppStatus(ctx context.Context, id int64, status ApplicationStatus) error {
	_, err := r.db.ExecContext(ctx, `UPDATE logistics_applications SET status=$1, updated_at=NOW() WHERE id=$2`, status, id)
	return err
}

func (r *pgRepo) ListLogApps(ctx context.Context, status *string) ([]LogisticApplication, error) {
	q := `SELECT id, original_application_id, status, created_at, updated_at FROM logistics_applications`
	args := []any{}
	if status != nil && *status != "" {
		q += " WHERE status=$1"
		args = append(args, *status)
	}
	q += " ORDER BY id DESC LIMIT 100"
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []LogisticApplication
	for rows.Next() {
		var a LogisticApplication
		if err := rows.Scan(&a.ID, &a.OriginalApplicationID, &a.Status, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, nil
}

func (r *pgRepo) InsertRoute(ctx context.Context, req CreateRouteRequest) (Route, error) {
	row := r.db.QueryRowContext(ctx, `INSERT INTO routes(truck_volume,truck_max_weight,departure_date,status) VALUES($1,$2,$3,'DRAFT')
	RETURNING id,truck_volume,truck_max_weight,departure_date,status,created_by_manager_id,created_at,updated_at`,
		req.TruckVolume, req.TruckMaxWeight, req.DepartureDate)
	var rt Route
	err := row.Scan(&rt.ID, &rt.TruckVolume, &rt.TruckMaxWeight, &rt.DepartureDate, &rt.Status, &rt.CreatedByManager, &rt.CreatedAt, &rt.UpdatedAt)
	return rt, err
}

func (r *pgRepo) InsertRoutePoint(ctx context.Context, routeID int64, p RoutePointInput) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO route_points(route_id,logistics_point_id,point_order,planned_arrival) VALUES($1,$2,$3,$4)`,
		routeID, p.LogisticsPointID, p.PointOrder, p.PlannedArrival)
	return err
}

func (r *pgRepo) AssignRouteApp(ctx context.Context, routeID, originalAppID, logAppID int64) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO route_applications(route_id,application_id,logistic_application_id) VALUES($1,$2,$3)`,
		routeID, originalAppID, logAppID)
	return err
}

func (r *pgRepo) RouteAppPairs(ctx context.Context, routeID int64) ([][2]int64, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT logistic_application_id, application_id FROM route_applications WHERE route_id=$1`, routeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out [][2]int64
	for rows.Next() {
		var a, b int64
		if err := rows.Scan(&a, &b); err != nil {
			return nil, err
		}
		out = append(out, [2]int64{a, b})
	}
	return out, nil
}

func (r *pgRepo) SetRouteInProgress(ctx context.Context, routeID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE routes SET status='IN_PROGRESS', updated_at=NOW() WHERE id=$1`, routeID)
	return err
}
