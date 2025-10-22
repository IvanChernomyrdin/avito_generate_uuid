-- +migrate Up
CREATE TABLE IF NOT EXISTS test (     
    test TEXT NOT NULL
);

-- +migrate Down  
DROP TABLE IF EXISTS test;