CREATE TABLE IF NOT EXISTS subjects (
	id INT NOT NULL GENERATED ALWAYS AS IDENTITY, -- equivalent to 'AUTO_INCREMENT' in MySQL
  name TEXT,
  description TEXT,
  PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS events (
	id INT NOT NULL GENERATED ALWAYS AS IDENTITY,
  subject_id INT NOT NULL,
  occurrences INT,
  insert_ts TIMESTAMPTZ, -- this backend will always store this timestamp in UTC
  last_update TIMESTAMPTZ, -- Always UTC
  PRIMARY KEY (id),
  CONSTRAINT fk_subjects
    FOREIGN KEY(subject_id)
      REFERENCES subjects(id)
);
