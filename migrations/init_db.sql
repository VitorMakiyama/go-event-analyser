CREATE TABLE IF NOT EXISTS subjects (
	id INT NOT NULL GENERATED ALWAYS AS IDENTITY, -- equivalent to 'AUTO_INCREMENT' in MySQL
  name TEXT,
  description TEXT,
  PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS events (
	id INT NOT NULL GENERATED ALWAYS AS IDENTITY,
  subject_id INT NOT NULL,
  ocurrences INT,
  insert_ts TIMESTAMP,
  last_update TIMESTAMP,
  PRIMARY KEY (id),
  CONSTRAINT fk_subjects
    FOREIGN KEY(subject_id)
      REFERENCES subjects(id)
);
