#!/bin/bash
# Script to create a superuser for Puxbay

cd backend || exit
go run ./cmd/createsuperuser/main.go
