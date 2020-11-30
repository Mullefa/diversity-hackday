CREATE TABLE IF NOT EXISTS journalists (
    journalist_name TEXT NOT NULL,
    profile_url TEXT,
    image_filename TEXT,
    gender TEXT,
    gender_confidence DOUBLE PRECISION,
    age_range_low SMALLINT,
    age_range_high SMALLINT,
    PRIMARY KEY (journalist_name)
);

CREATE TABLE IF NOT EXISTS articles (
    first_publication_date DATE NOT NULL,
    id TEXT NOT NULL,
    web_url TEXT NOT NULL,
    journalist_name TEXT NOT NULL,
    PRIMARY KEY (journalist_name, id),
    CONSTRAINT fk_journalist_name
        FOREIGN KEY(journalist_name)
            REFERENCES journalists(journalist_name)
);

CREATE INDEX id_index ON articles (id);
CREATE INDEX first_publication_date_index ON articles (first_publication_date);
CREATE INDEX journalist_name_index ON articles (journalist_name);
