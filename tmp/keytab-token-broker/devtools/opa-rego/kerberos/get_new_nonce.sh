#!/bin/bash

opa eval -i input.json -d policy.rego "grant_new_nonce = data.kbridge.grant_new_nonce"
