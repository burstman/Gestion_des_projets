CREATE TABLE IF NOT EXISTS workers_registry (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text NOT NULL,
    nationality text NOT NULL,
    id_number serial UNIQUE,
    place_of_residence text NOT NULL,
    Workplace text NOT NULL,
    blood_type text NOT NULL,
    name_of_sponsor text NOT NULL,
    sponsor bool NOT NULL, 
    image_reference text not NULL,
    version integer NOT NULL DEFAULT 1
    
    );