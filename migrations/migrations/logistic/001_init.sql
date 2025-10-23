CREATE TABLE IF NOT EXISTS logistic_points (
	id BIGSERIAL PRIMARY KEY,
	title TEXT NOT NULL,
	address TEXT NOT NULL DEFAULT '',
	lat DOUBLE PRECISION,
	lon DOUBLE PRECISION,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS shipments (
	id BIGSERIAL PRIMARY KEY,
	max_weight_kg DOUBLE PRECISION NOT NULL,
	max_volume_m3 DOUBLE PRECISION NOT NULL,
	departure_at TIMESTAMPTZ NOT NULL,
	status TEXT NOT NULL DEFAULT 'PLANNED',
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS route_points (
	id BIGSERIAL PRIMARY KEY,
	shipment_id BIGINT NOT NULL REFERENCES shipments(id) ON DELETE CASCADE,
	point_id BIGINT NOT NULL REFERENCES logistic_points(id),
	planned_arrive_at TIMESTAMPTZ NOT NULL,
	ordinal SMALLINT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_route_points_unique ON route_points(shipment_id, ordinal);

CREATE TABLE IF NOT EXISTS assignments (
	id BIGSERIAL PRIMARY KEY,
	shipment_id BIGINT NOT NULL REFERENCES shipments(id) ON DELETE CASCADE,
	application_external_id BIGINT NOT NULL,
	pickup_point_external_id BIGINT,
	dropoff_point_external_id BIGINT,
	weight_kg DOUBLE PRECISION NOT NULL,
	volume_m3 DOUBLE PRECISION NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO logistic_points(title,address,lat,lon) VALUES
('Склад Москва', 'Москва, ул. Склада, 1', 55.7558, 37.6176),
('Склад СПб', 'Санкт-Петербург, Невский пр., 1', 59.9311, 30.3609)
ON CONFLICT DO NOTHING;
