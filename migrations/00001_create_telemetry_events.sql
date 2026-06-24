-- +goose Up

CREATE TABLE telemetry_events (
    id          SERIAL      PRIMARY KEY,
    vehicle_id  TEXT                   ,
    battery_pct FLOAT                  ,
    speed_kph   FLOAT                  ,
    temp_c      FLOAT                  ,
    is_charging BOOLEAN                ,
    timestamp   TIMESTAMPTZ
);

-- +goose Down

DROP TABLE telemetry_events;