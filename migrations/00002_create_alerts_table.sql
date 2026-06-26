
-- +goose Up


CREATE TABLE alerts (
    id              SERIAL  PRIMARY KEY,
    vehicle_id      TEXT               ,
    alert_type      TEXT               ,
    value           FLOAT              ,
    message         TEXT               ,
    timestamp       TIMESTAMPTZ
);

-- +goose Down

DROP TABLE alerts;