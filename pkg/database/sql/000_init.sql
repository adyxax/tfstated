CREATE TABLE schema_version (
  version INTEGER NOT NULL
) STRICT;

CREATE TABLE states (
  id INTEGER PRIMARY KEY,
  path TEXT NOT NULL,
  lock TEXT
) STRICT;
CREATE UNIQUE INDEX states_path on states(path);

CREATE TABLE versions (
  id INTEGER PRIMARY KEY,
  state_id INTEGER,
  data BLOB,
  lock TEXT,
  created INTEGER DEFAULT (unixepoch()),
  FOREIGN KEY(state_id) REFERENCES states(id) ON DELETE CASCADE
) STRICT;
