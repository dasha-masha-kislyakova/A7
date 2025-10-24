package office

import (
	"context"
	"database/sql"
)

type Repo interface {
	EnsureSchema(ctx context.Context) error
	Insert(ctx context.Context, r CreateApplicationRequest) (Application, error)
	GetByID(ctx context.Context, id int64) (Application, error)
	UpdateStatus(ctx context.Context, id int64, status ApplicationStatus) error
	List(ctx context.Context, status *string) ([]Application, error)
}

type pgRepo struct{ db *sql.DB }

func NewRepo(db *sql.DB) Repo { return &pgRepo{db: db} }

func (r *pgRepo) EnsureSchema(ctx context.Context) error {
	ddl := `
CREATE TYPE IF NOT EXISTS application_status AS ENUM ('NEW','IN_PROGRESS','SHIPPED','DELIVERED','CANCELLED');

CREATE TABLE IF NOT EXISTS applications (
  id BIGSERIAL PRIMARY KEY,
  status application_status NOT NULL DEFAULT 'NEW',
  logistics_point_id BIGINT NOT NULL,

  sender_org_name TEXT NOT NULL,
  sender_inn TEXT NOT NULL,
  sender_contact_fio TEXT NOT NULL,
  sender_contact_phone TEXT NOT NULL,
  sender_email TEXT,

  cargo_name TEXT NOT NULL,
  cargo_count INTEGER NOT NULL CHECK (cargo_count > 0),
  cargo_weight NUMERIC(10,2) NOT NULL CHECK (cargo_weight > 0),
  cargo_volume NUMERIC(10,2) NOT NULL CHECK (cargo_volume > 0),
  special_requirements TEXT,

  recipient_org_name TEXT NOT NULL,
  recipient_address TEXT NOT NULL,
  recipient_contact_fio TEXT NOT NULL,
  recipient_contact_phone TEXT NOT NULL,

  created_by_manager_id BIGINT NOT NULL DEFAULT 1,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_applications_status ON applications(status);
CREATE INDEX IF NOT EXISTS idx_applications_point ON applications(logistics_point_id);
`
	_, err := r.db.ExecContext(ctx, ddl)
	return err
}

func (r *pgRepo) Insert(ctx context.Context, req CreateApplicationRequest) (Application, error) {
	row := r.db.QueryRowContext(ctx, `
INSERT INTO applications (
  status, logistics_point_id,
  sender_org_name, sender_inn, sender_contact_fio, sender_contact_phone, sender_email,
  cargo_name, cargo_count, cargo_weight, cargo_volume, special_requirements,
  recipient_org_name, recipient_address, recipient_contact_fio, recipient_contact_phone
) VALUES ('NEW',$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
RETURNING id,status,logistics_point_id,sender_org_name,sender_inn,sender_contact_fio,sender_contact_phone,sender_email,
          cargo_name,cargo_count,cargo_weight,cargo_volume,special_requirements,
          recipient_org_name,recipient_address,recipient_contact_fio,recipient_contact_phone,
          created_by_manager_id,created_at,updated_at
`,
		req.LogisticsPointID,
		req.SenderOrgName, req.SenderINN, req.SenderContactFIO, req.SenderContactPhone, req.SenderEmail,
		req.CargoName, req.CargoCount, req.CargoWeight, req.CargoVolume, req.SpecialRequirements,
		req.RecipientOrgName, req.RecipientAddress, req.RecipientContactFIO, req.RecipientContactPhone,
	)

	var app Application
	err := row.Scan(
		&app.ID, &app.Status, &app.LogisticsPointID,
		&app.SenderOrgName, &app.SenderINN, &app.SenderContactFIO, &app.SenderContactPhone, &app.SenderEmail,
		&app.CargoName, &app.CargoCount, &app.CargoWeight, &app.CargoVolume, &app.SpecialRequirements,
		&app.RecipientOrgName, &app.RecipientAddress, &app.RecipientContactFIO, &app.RecipientContactPhone,
		&app.CreatedByManagerID, &app.CreatedAt, &app.UpdatedAt,
	)
	return app, err
}

func (r *pgRepo) GetByID(ctx context.Context, id int64) (Application, error) {
	row := r.db.QueryRowContext(ctx, `
SELECT id,status,logistics_point_id,
       sender_org_name,sender_inn,sender_contact_fio,sender_contact_phone,sender_email,
       cargo_name,cargo_count,cargo_weight,cargo_volume,special_requirements,
       recipient_org_name,recipient_address,recipient_contact_fio,recipient_contact_phone,
       created_by_manager_id,created_at,updated_at
FROM applications WHERE id=$1`, id)

	var app Application
	err := row.Scan(&app.ID, &app.Status, &app.LogisticsPointID,
		&app.SenderOrgName, &app.SenderINN, &app.SenderContactFIO, &app.SenderContactPhone, &app.SenderEmail,
		&app.CargoName, &app.CargoCount, &app.CargoWeight, &app.CargoVolume, &app.SpecialRequirements,
		&app.RecipientOrgName, &app.RecipientAddress, &app.RecipientContactFIO, &app.RecipientContactPhone,
		&app.CreatedByManagerID, &app.CreatedAt, &app.UpdatedAt)
	return app, err
}

func (r *pgRepo) UpdateStatus(ctx context.Context, id int64, status ApplicationStatus) error {
	_, err := r.db.ExecContext(ctx, `UPDATE applications SET status=$1, updated_at=NOW() WHERE id=$2`, status, id)
	return err
}

func (r *pgRepo) List(ctx context.Context, status *string) ([]Application, error) {
	q := `SELECT id,status,logistics_point_id,
       sender_org_name,sender_inn,sender_contact_fio,sender_contact_phone,sender_email,
       cargo_name,cargo_count,cargo_weight,cargo_volume,special_requirements,
       recipient_org_name,recipient_address,recipient_contact_fio,recipient_contact_phone,
       created_by_manager_id,created_at,updated_at
FROM applications`
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
	var list []Application
	for rows.Next() {
		var app Application
		if err := rows.Scan(&app.ID, &app.Status, &app.LogisticsPointID,
			&app.SenderOrgName, &app.SenderINN, &app.SenderContactFIO, &app.SenderContactPhone, &app.SenderEmail,
			&app.CargoName, &app.CargoCount, &app.CargoWeight, &app.CargoVolume, &app.SpecialRequirements,
			&app.RecipientOrgName, &app.RecipientAddress, &app.RecipientContactFIO, &app.RecipientContactPhone,
			&app.CreatedByManagerID, &app.CreatedAt, &app.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, app)
	}
	return list, nil
}
