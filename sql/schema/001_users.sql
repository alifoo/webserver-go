-- +goose Up
CREATE TABLE users (
  id uuid,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  email TEXT UNIQUE NOT NULL,
  PRIMARY KEY (id)
);

-- +goose Down
DROP TABLE users;
