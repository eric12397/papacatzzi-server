CREATE TABLE sightings (
    id SERIAL PRIMARY KEY,
    user_id TEXT NOT NULL,
    photo_url TEXT,
    animal_type TEXT NOT NULL,
    description TEXT,
    latitude FLOAT NOT NULL,
    longitude FLOAT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL,
    email TEXT NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT FALSE
);

-- Create a spatial index for efficient querying
CREATE INDEX idx_sightings_coordinates ON sightings USING GIST (
    ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)
);