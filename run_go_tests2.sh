#!/bin/bash
cd /Users/san/Fun/sqlparser-rs/go
go test ./... 2>&1 | tail -50
