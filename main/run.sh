#!/bin/bash

cd "$(dirname "$0")" || exit 2
go build && ./main
