CREATE TABLE IF NOT EXISTS pvzs (
    id UUID PRIMARY KEY,
    registration_date TIMESTAMP NOT NULL,
    city VARCHAR(50) NOT NULL
);

CREATE TABLE IF NOT EXISTS receptions (
  id UUID PRIMARY KEY,
  pvz_id UUID NOT NULL REFERENCES pvzs(id),
  date_time TIMESTAMP NOT NULL,
  status VARCHAR(20) NOT NULL
);

CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY,
    date_time TIMESTAMP NOT NULL,
    type VARCHAR(50) NOT NULL,
    reception_id UUID NOT NULL REFERENCES receptions(id)
);

CREATE TABLE IF NOT EXISTS users (
     id UUID PRIMARY KEY,
     email VARCHAR(255) UNIQUE NOT NULL,
     password_hash VARCHAR(255) NOT NULL,
     role VARCHAR(50) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_reception_date_time ON receptions (date_time);
CREATE INDEX IF NOT EXISTS idx_product_reception_id ON products (reception_id);
CREATE INDEX IF NOT EXISTS idx_reception_pvz_id ON receptions (pvz_id);