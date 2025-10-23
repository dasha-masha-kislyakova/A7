CREATE TABLE IF NOT EXISTS clients (
	id BIGSERIAL PRIMARY KEY,
	name TEXT NOT NULL,
	inn TEXT NOT NULL,
	fio TEXT NOT NULL,
	tel TEXT NOT NULL,
	email TEXT NOT NULL,
	UNIQUE (inn, name)
);

CREATE TABLE IF NOT EXISTS cargos (
	id BIGSERIAL PRIMARY KEY,
	name TEXT NOT NULL,
	boxes BIGINT NOT NULL,
	total_weight DOUBLE PRECISION NOT NULL,
	total_volume DOUBLE PRECISION NOT NULL,
	special_requirements TEXT
);

CREATE TABLE IF NOT EXISTS recipients (
	id BIGSERIAL PRIMARY KEY,
	name_company TEXT NOT NULL,
	address TEXT NOT NULL,
	fio TEXT NOT NULL,
	telephone TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS applications (
	id BIGSERIAL PRIMARY KEY,
	pickup_point_id BIGINT NOT NULL,
	dropoff_point_id BIGINT NOT NULL,
	client_id BIGINT NOT NULL REFERENCES clients(id),
	cargo_id BIGINT NOT NULL REFERENCES cargos(id),
	recipient_id BIGINT NOT NULL REFERENCES recipients(id),
	status TEXT NOT NULL DEFAULT 'NEW',
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	accepted_at TIMESTAMPTZ,
	dispatched_at TIMESTAMPTZ,
	delivered_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_applications_status ON applications(status);
CREATE INDEX IF NOT EXISTS idx_applications_points ON applications(pickup_point_id, dropoff_point_id);
