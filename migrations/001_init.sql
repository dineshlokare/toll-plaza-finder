-- Toll Plazas Table
CREATE TABLE IF NOT EXISTS toll_plazas (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    geo_state VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_toll_plaza UNIQUE (name, latitude, longitude)
);

-- Indexes for fast coordinate, bounding box and state searches
CREATE INDEX IF NOT EXISTS idx_toll_plazas_lat ON toll_plazas(latitude);
CREATE INDEX IF NOT EXISTS idx_toll_plazas_lon ON toll_plazas(longitude);
CREATE INDEX IF NOT EXISTS idx_toll_plazas_lat_lon ON toll_plazas(latitude, longitude);
CREATE INDEX IF NOT EXISTS idx_toll_plazas_state ON toll_plazas(geo_state);

-- Pincode Geocoding Cache Table
CREATE TABLE IF NOT EXISTS pincode_cache (
    pincode VARCHAR(10) PRIMARY KEY,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    display_name TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
