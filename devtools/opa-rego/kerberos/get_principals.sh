#!/bin/bash

opa eval -i input.json -d policy.rego "data.kbridge.get_principals[get_principals]"
