package repo

import (
	"database/sql"
	"time"
)

type Repo struct{ DB *sql.DB }

func New(db *sql.DB) *Repo { return &Repo{DB: db} }

func (r *Repo) CreateLogisticPoint(title, address string, lat, lon *float64) (int64, error) {
	var id int64
	row := r.DB.QueryRow(`INSERT INTO logistic_points(title,address,lat,lon,created_at) VALUES($1,$2,$3,$4,now()) RETURNING id`, title, address, lat, lon)
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repo) ListLogisticPoints() ([]map[string]any, error) {
	rows, err := r.DB.Query(`SELECT id,title,address,lat,lon,created_at FROM logistic_points ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id int64
		var title, address string
		var lat, lon *float64
		var created time.Time
		if err := rows.Scan(&id, &title, &address, &lat, &lon, &created); err == nil {
			out = append(out, map[string]any{"id": id, "title": title, "address": address, "lat": lat, "lon": lon, "created_at": created})
		}
	}
	return out, nil
}

func (r *Repo) GetPoint(id int64) (title string, address string, lat, lon *float64, err error) {
	row := r.DB.QueryRow(`SELECT title,address,lat,lon FROM logistic_points WHERE id=$1`, id)
	err = row.Scan(&title, &address, &lat, &lon)
	return
}

func (r *Repo) CreateShipment(maxW, maxV float64, departure time.Time) (int64, error) {
	var id int64
	row := r.DB.QueryRow(`INSERT INTO shipments(max_weight_kg,max_volume_m3,departure_at,status,created_at) VALUES($1,$2,$3,'PLANNED',now()) RETURNING id`,
		maxW, maxV, departure)
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repo) InsertRoutePoint(shipmentID, pointID int64, at time.Time, ord int) error {
	_, err := r.DB.Exec(`INSERT INTO route_points(shipment_id,point_id,planned_arrive_at,ordinal) VALUES($1,$2,$3,$4)`, shipmentID, pointID, at, ord)
	return err
}

func (r *Repo) ListShipments() ([]map[string]any, error) {
	rows, err := r.DB.Query(`SELECT id,max_weight_kg,max_volume_m3,departure_at,status,created_at FROM shipments ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id int64
		var mw, mv float64
		var dep time.Time
		var st string
		var created time.Time
		if err := rows.Scan(&id, &mw, &mv, &dep, &st, &created); err == nil {
			out = append(out, map[string]any{"id": id, "max_weight_kg": mw, "max_volume_m3": mv, "departure_at": dep, "status": st, "created_at": created})
		}
	}
	return out, nil
}

func (r *Repo) GetShipment(id int64) (map[string]any, error) {
	row := r.DB.QueryRow(`SELECT id,max_weight_kg,max_volume_m3,departure_at,status,created_at FROM shipments WHERE id=$1`, id)
	var sid int64
	var mw, mv float64
	var dep time.Time
	var st string
	var created time.Time
	if err := row.Scan(&sid, &mw, &mv, &dep, &st, &created); err != nil {
		return nil, err
	}
	rows, err := r.DB.Query(`SELECT id,point_id,planned_arrive_at,ordinal FROM route_points WHERE shipment_id=$1 ORDER BY ordinal`, sid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	route := []map[string]any{}
	for rows.Next() {
		var id, pid int64
		var at time.Time
		var ord int16
		if err := rows.Scan(&id, &pid, &at, &ord); err == nil {
			route = append(route, map[string]any{"id": id, "point_id": pid, "planned_arrive_at": at, "ordinal": ord})
		}
	}
	return map[string]any{"id": sid, "max_weight_kg": mw, "max_volume_m3": mv, "departure_at": dep, "status": st, "created_at": created, "route": route}, nil
}

func (r *Repo) GetOpenShipments() ([]map[string]any, error) {
	rows, err := r.DB.Query(`SELECT s.id,s.max_weight_kg,s.max_volume_m3,s.departure_at, p1.point_id, p2.point_id 
		FROM shipments s 
		JOIN LATERAL (SELECT point_id FROM route_points WHERE shipment_id=s.id AND ordinal=1 LIMIT 1) p1 ON true
		JOIN LATERAL (SELECT point_id FROM route_points WHERE shipment_id=s.id AND ordinal=2 LIMIT 1) p2 ON true
		WHERE s.status='PLANNED'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id int64
		var mw, mv float64
		var dep time.Time
		var p1, p2 int64
		if err := rows.Scan(&id, &mw, &mv, &dep, &p1, &p2); err == nil {
			out = append(out, map[string]any{"id": id, "max_weight_kg": mw, "max_volume_m3": mv, "departure_at": dep, "p1": p1, "p2": p2})
		}
	}
	return out, nil
}

func (r *Repo) CurrentLoad(shipmentID int64) (float64, float64, error) {
	row := r.DB.QueryRow(`SELECT COALESCE(SUM(weight_kg),0), COALESCE(SUM(volume_m3),0) FROM assignments WHERE shipment_id=$1`, shipmentID)
	var w, v float64
	if err := row.Scan(&w, &v); err != nil {
		return 0, 0, err
	}
	return w, v, nil
}

func (r *Repo) AddAssignment(shipmentID int64, appID int64, w, v float64, pp, dp int64) error {
	_, err := r.DB.Exec(`INSERT INTO assignments(shipment_id,application_external_id,weight_kg,volume_m3,created_at,pickup_point_external_id,dropoff_point_external_id) VALUES($1,$2,$3,$4,now(),$5,$6)`,
		shipmentID, appID, w, v, pp, dp)
	return err
}

func (r *Repo) MarkShipmentDeparted(id int64) error {
	_, err := r.DB.Exec(`UPDATE shipments SET status='DEPARTED' WHERE id=$1`, id)
	return err
}
