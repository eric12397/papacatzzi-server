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

INSERT INTO sightings (user_id, animal_type, latitude, longitude) VALUES 
    ('Johnny Park', 'Dog', 39.945147922934126, -75.14459310653145),
    ('David Robinson', 'Cat', 39.945752660141075, -75.144704423162),
    ('Brandon Gomez', 'Cat', 39.94562419967605, -75.14516242278123),
    ('Doug Lee', 'Cat', 39.945727657063856, -75.14653630166342);

-- Create a spatial index for efficient querying
CREATE INDEX idx_sightings_coordinates ON sightings USING GIST (
    ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)
);