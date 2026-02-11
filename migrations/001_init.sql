PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS certificates (
    id TEXT PRIMARY KEY,
    common_name TEXT NOT NULL CHECK(length(common_name) <= 255),
    serial_number TEXT NOT NULL CHECK(length(serial_number) <= 128),
    issuer TEXT NOT NULL CHECK(length(issuer) <= 255),
    not_before DATETIME NOT NULL,
    not_after DATETIME NOT NULL,
    fingerprint_sha256 TEXT NOT NULL UNIQUE CHECK(length(fingerprint_sha256) = 64),
    created_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_cert_not_after 
ON certificates(not_after);
