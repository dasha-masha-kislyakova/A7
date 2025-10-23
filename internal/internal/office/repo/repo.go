package repo

import (
	"a7/internal/office"
	"database/sql"
	"time"
)

type Repo struct{ DB *sql.DB }

func New(db *sql.DB) *Repo { return &Repo{DB: db} }

func (r *Repo) UpsertClient(c *office.Client) (int64, error) {
	var id int64
	err := r.DB.QueryRow(`SELECT id FROM clients WHERE inn=$1 AND name=$2`, c.INN, c.Name).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	returning := r.DB.QueryRow(`INSERT INTO clients(name,inn,fio,tel,email) VALUES($1,$2,$3,$4,$5) RETURNING id`,
		c.Name, c.INN, c.FIO, c.Tel, c.Email)
	if err := returning.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repo) InsertCargo(ca *office.Cargo) (int64, error) {
	var id int64
	returning := r.DB.QueryRow(`INSERT INTO cargos(name,boxes,total_weight,total_volume,special_requirements) VALUES($1,$2,$3,$4,$5) RETURNING id`,
		ca.Name, ca.Boxes, ca.TotalWeight, ca.TotalVolume, ca.SpecialReq)
	if err := returning.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repo) InsertRecipient(re *office.Recipient) (int64, error) {
	var id int64
	returning := r.DB.QueryRow(`INSERT INTO recipients(name_company,address,fio,telephone) VALUES($1,$2,$3,$4) RETURNING id`,
		re.Name, re.Address, re.FIO, re.Telephone)
	if err := returning.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repo) InsertApplication(app *office.ApplicationRow) (int64, error) {
	var id int64
	returning := r.DB.QueryRow(`INSERT INTO applications(pickup_point_id,dropoff_point_id,client_id,cargo_id,recipient_id,status,created_at) VALUES($1,$2,$3,$4,$5,$6,now()) RETURNING id`,
		app.PickupPointID, app.DropoffPointID, app.ClientID, app.CargoID, app.RecipientID, "NEW")
	if err := returning.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repo) ChangeStatus(id int64, newStatus string) error {
	_, err := r.DB.Exec(`UPDATE applications SET status=$2, 
		accepted_at = CASE WHEN $2='IN_WORK' THEN now() ELSE accepted_at END,
		dispatched_at = CASE WHEN $2='IN_TRANSIT' THEN now() ELSE dispatched_at END,
		delivered_at = CASE WHEN $2='DELIVERED' THEN now() ELSE delivered_at END
		WHERE id=$1`, id, newStatus)
	return err
}

func (r *Repo) GetApplication(id int64) (map[string]any, error) {
	row := r.DB.QueryRow(`
		SELECT a.id, a.pickup_point_id, a.dropoff_point_id, a.status, a.created_at, a.accepted_at, a.dispatched_at, a.delivered_at,
			c.id, c.name, c.inn, c.fio, c.tel, c.email,
			g.id, g.name, g.boxes, g.total_weight, g.total_volume, g.special_requirements,
			r.id, r.name_company, r.address, r.fio, r.telephone
		FROM applications a
		JOIN clients c ON c.id=a.client_id
		JOIN cargos g ON g.id=a.cargo_id
		JOIN recipients r ON r.id=a.recipient_id
		WHERE a.id=$1`, id)
	var (
		appID, pp, dp, cid, gid, rid    int64
		status                          string
		created                         time.Time
		acc, disp, deliv                *time.Time
		cName, cInn, cFio, cTel, cEmail string
		gName                           string
		gBoxes                          int64
		gW, gV                          float64
		gSpec                           *string
		rName, rAddr, rFio, rTel        string
	)
	if err := row.Scan(&appID, &pp, &dp, &status, &created, &acc, &disp, &deliv, &cid, &cName, &cInn, &cFio, &cTel, &cEmail, &gid, &gName, &gBoxes, &gW, &gV, &gSpec, &rid, &rName, &rAddr, &rFio, &rTel); err != nil {
		return nil, err
	}
	m := map[string]any{
		"id": appID, "pickup_point_id": pp, "dropoff_point_id": dp, "status": status,
		"created_at": created, "accepted_at": acc, "dispatched_at": disp, "delivered_at": deliv,
		"client":    map[string]any{"id": cid, "name": cName, "inn": cInn, "fio": cFio, "tel": cTel, "email": cEmail},
		"cargo":     map[string]any{"id": gid, "name": gName, "boxes": gBoxes, "total_weight": gW, "total_volume": gV, "special_requirements": gSpec},
		"recipient": map[string]any{"id": rid, "name_company": rName, "address": rAddr, "fio": rFio, "telephone": rTel},
	}
	return m, nil
}

func (r *Repo) ListApplications(status string) ([]map[string]any, error) {
	query := `SELECT id, pickup_point_id, dropoff_point_id, status, created_at FROM applications`
	args := []any{}
	if status != "" {
		query += ` WHERE status=$1`
		args = append(args, status)
	}
	query += ` ORDER BY id DESC`
	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, pp, dp int64
		var st string
		var created time.Time
		if err := rows.Scan(&id, &pp, &dp, &st, &created); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"id": id, "pickup_point_id": pp, "dropoff_point_id": dp, "status": st, "created_at": created})
	}
	return out, nil
}

func (r *Repo) ListAvailableForRoute(pp, dp int64, before time.Time) ([]map[string]any, error) {
	rows, err := r.DB.Query(`
		SELECT a.id, g.total_weight, g.total_volume, a.pickup_point_id, a.dropoff_point_id, a.created_at
		FROM applications a
		JOIN cargos g ON g.id=a.cargo_id
		WHERE a.status='NEW' AND a.pickup_point_id=$1 AND a.dropoff_point_id=$2 AND a.created_at <= $3
		ORDER BY a.created_at ASC`, pp, dp, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id int64
		var w, v float64
		var ppi, dpi int64
		var created time.Time
		if err := rows.Scan(&id, &w, &v, &ppi, &dpi, &created); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"id": id, "total_weight": w, "total_volume": v, "pickup_point_id": ppi, "dropoff_point_id": dpi, "created_at": created})
	}
	return out, nil
}
