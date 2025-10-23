package repo

import (
	"database/sql"
	"time"

	"a7/internal/logistic"
)

type Repo struct{ DB *sql.DB }

func New(db *sql.DB) *Repo { return &Repo{DB: db} }

func (r *Repo) CreateLogisticPoint(title, address string, lat, lon *float64) (int64, error) {
	var id int64
	row := r.DB.QueryRow(`INSERT INTO logistic_points(title,address,lat,lon,created_at)
		VALUES($1,$2,$3,$4,now()) RETURNING id`, title, address, lat, lon)
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repo) ListLogisticPoints() ([]logistic.LogisticPoint, error) {
	rows, err := r.DB.Query(`SELECT id,title,address,lat,lon,created_at FROM logistic_points ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []logistic.LogisticPoint{}
	for rows.Next() {
		var lp logistic.LogisticPoint
		var addr sql.NullString
		var lat, lon sql.NullFloat64
		if err := rows.Scan(&lp.ID, &lp.Title, &addr, &lat, &lon, &lp.CreatedAt); err != nil {
			return nil, err
		}
		if addr.Valid && addr.String != "" {
			s := addr.String
			lp.Address = &s
		}
		if lat.Valid {
			v := lat.Float64
			lp.Lat = &v
		}
		if lon.Valid {
			v := lon.Float64
			lp.Lon = &v
		}
		out = append(out, lp)
	}
	return out, nil
}

func (r *Repo) GetPoint(id int64) (logistic.LogisticPoint, error) {
	row := r.DB.QueryRow(`SELECT id,title,address,lat,lon,created_at FROM logistic_points WHERE id=$1`, id)
	var lp logistic.LogisticPoint
	var addr sql.NullString
	var lat, lon sql.NullFloat64
	if err := row.Scan(&lp.ID, &lp.Title, &addr, &lat, &lon, &lp.CreatedAt); err != nil {
		return logistic.LogisticPoint{}, err
	}
	if addr.Valid && addr.String != "" {
		s := addr.String
		lp.Address = &s
	}
	if lat.Valid {
		v := lat.Float64
		lp.Lat = &v
	}
	if lon.Valid {
		v := lon.Float64
		lp.Lon = &v
	}
	return lp, nil
}

func (r *Repo) CreateShipment(maxW, maxV float64, departure time.Time) (int64, error) {
	var id int64
	row := r.DB.QueryRow(`INSERT INTO shipments(max_weight_kg,max_volume_m3,departure_at,status,created_at)
		VALUES($1,$2,$3,'PLANNED',now()) RETURNING id`, maxW, maxV, departure)
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repo) InsertRoutePoint(shipmentID, pointID int64, at time.Time, ord int) error {
	_, err := r.DB.Exec(`INSERT INTO route_points(shipment_id,point_id,planned_arrive_at,ordinal)
		VALUES($1,$2,$3,$4)`, shipmentID, pointID, at, ord)
	return err
}

func (r *Repo) ListShipments() ([]logistic.Shipment, error) {
	rows, err := r.DB.Query(`SELECT id,max_weight_kg,max_volume_m3,departure_at,status,created_at FROM shipments ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []logistic.Shipment{}
	for rows.Next() {
		var sh logistic.Shipment
		var st string
		if err := rows.Scan(&sh.ID, &sh.MaxWeightKg, &sh.MaxVolumeM3, &sh.DepartureAt, &st, &sh.CreatedAt); err != nil {
			return nil, err
		}
		sh.Status = logistic.ShipmentStatus(st)
		out = append(out, sh)
	}
	return out, nil
}

func (r *Repo) GetShipment(id int64) (logistic.Shipment, []logistic.RoutePoint, error) {
	row := r.DB.QueryRow(`SELECT id,max_weight_kg,max_volume_m3,departure_at,status,created_at FROM shipments WHERE id=$1`, id)
	var sh logistic.Shipment
	var st string
	if err := row.Scan(&sh.ID, &sh.MaxWeightKg, &sh.MaxVolumeM3, &sh.DepartureAt, &st, &sh.CreatedAt); err != nil {
		return logistic.Shipment{}, nil, err
	}
	sh.Status = logistic.ShipmentStatus(st)

	rs, err := r.DB.Query(`SELECT id,point_id,planned_arrive_at,ordinal FROM route_points WHERE shipment_id=$1 ORDER BY ordinal`, sh.ID)
	if err != nil {
		return logistic.Shipment{}, nil, err
	}
	defer rs.Close()

	route := []logistic.RoutePoint{}
	for rs.Next() {
		var rp logistic.RoutePoint
		if err := rs.Scan(&rp.ID, &rp.PointID, &rp.PlannedArriveAt, &rp.Ordinal); err != nil {
			return logistic.Shipment{}, nil, err
		}
		rp.ShipmentID = sh.ID
		route = append(route, rp)
	}
	return sh, route, nil
}

type OpenShipment struct {
	ID          int64
	MaxWeightKg float64
	MaxVolumeM3 float64
	DepartureAt time.Time
	P1          int64
	P2          int64
}

func (r *Repo) GetOpenShipments() ([]OpenShipment, error) {
	rows, err := r.DB.Query(`
		SELECT s.id,s.max_weight_kg,s.max_volume_m3,s.departure_at,
			(SELECT point_id FROM route_points WHERE shipment_id=s.id AND ordinal=1 LIMIT 1) AS p1,
			(SELECT point_id FROM route_points WHERE shipment_id=s.id AND ordinal=2 LIMIT 1) AS p2
		FROM shipments s
		WHERE s.status='PLANNED'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []OpenShipment{}
	for rows.Next() {
		var o OpenShipment
		if err := rows.Scan(&o.ID, &o.MaxWeightKg, &o.MaxVolumeM3, &o.DepartureAt, &o.P1, &o.P2); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, nil
}

func (r *Repo) MarkShipmentDeparted(id int64) error {
	_, err := r.DB.Exec(`UPDATE shipments SET status='DEPARTED' WHERE id=$1`, id)
	return err
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
	_, err := r.DB.Exec(`INSERT INTO assignments(shipment_id,application_external_id,weight_kg,volume_m3,created_at,pickup_point_external_id,dropoff_point_external_id)
		VALUES ($1,$2,$3,$4,now(),$5,$6)`, shipmentID, appID, w, v, pp, dp)
	return err
}

func (r *Repo) ListAssignments() ([]logistic.Assignment, error) {
	rows, err := r.DB.Query(`SELECT id,shipment_id,application_external_id,pickup_point_external_id,dropoff_point_external_id,weight_kg,volume_m3,created_at
		FROM assignments ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []logistic.Assignment{}
	for rows.Next() {
		var a logistic.Assignment
		var pp, dp sql.NullInt64
		if err := rows.Scan(&a.ID, &a.ShipmentID, &a.ApplicationExternalID, &pp, &dp, &a.WeightKg, &a.VolumeM3, &a.CreatedAt); err != nil {
			return nil, err
		}
		if pp.Valid {
			v := pp.Int64
			a.PickupPointExternalID = &v
		}
		if dp.Valid {
			v := dp.Int64
			a.DropoffPointExternalID = &v
		}
		out = append(out, a)
	}
	return out, nil
}

func (r *Repo) ListAssignmentsByShipment(shipmentID int64) ([]logistic.Assignment, error) {
	rows, err := r.DB.Query(`SELECT id,shipment_id,application_external_id,pickup_point_external_id,dropoff_point_external_id,weight_kg,volume_m3,created_at
		FROM assignments WHERE shipment_id=$1 ORDER BY id DESC`, shipmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []logistic.Assignment{}
	for rows.Next() {
		var a logistic.Assignment
		var pp, dp sql.NullInt64
		if err := rows.Scan(&a.ID, &a.ShipmentID, &a.ApplicationExternalID, &pp, &dp, &a.WeightKg, &a.VolumeM3, &a.CreatedAt); err != nil {
			return nil, err
		}
		if pp.Valid {
			v := pp.Int64
			a.PickupPointExternalID = &v
		}
		if dp.Valid {
			v := dp.Int64
			a.DropoffPointExternalID = &v
		}
		out = append(out, a)
	}
	return out, nil
}
