#!/bin/bash

protoc --go_out=src/grpc/ --go_opt=paths=source_relative \
    --go-grpc_out=src/grpc/ --go-grpc_opt=paths=source_relative \
    golden-gate.proto