#!/bin/sh
./migrate -database "postgres://user:password@localhost:5432/arrivo?sslmode=disable" -path internal/database/migrations up