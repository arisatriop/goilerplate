CREATE TABLE bass (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    code       VARCHAR(255) NOT NULL UNIQUE,
    bas        TEXT         NOT NULL,
    is_active  BOOLEAN      NOT NULL DEFAULT true,
    created_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP    NULL DEFAULT NULL,
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255) NOT NULL,
    deleted_by VARCHAR(255) DEFAULT NULL
);

COMMENT ON TABLE bass IS 'Bass table for storing bas records';

CREATE INDEX idx_bass_code ON bass(code);
CREATE INDEX idx_bass_is_active ON bass(is_active);
CREATE INDEX idx_bass_deleted_at ON bass(deleted_at);
CREATE INDEX idx_bass_active ON bass(code, is_active, deleted_at);
