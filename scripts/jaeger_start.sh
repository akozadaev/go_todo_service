#!/bin/bash

set -e
cd /home/akozadaev/programms/jaeger/jaeger-2.2.0-linux-amd64 && ./jaeger --config=./cmd/jaeger/internal/all-in-one.yaml > /tmp/jaeger.log 2>&1 &







