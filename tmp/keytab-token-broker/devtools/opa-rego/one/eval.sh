#!/bin/bash

opa eval -i input.json -d policy.rego "x = data.authz.allow" | jq '.result[].expressions[].value'
