CREATE TABLE schema_version (
  version INTEGER NOT NULL
) STRICT;

CREATE TABLE states (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL,
  data BLOB,
  lock TEXT
) STRICT;
CREATE UNIQUE INDEX states_name on states(name);
