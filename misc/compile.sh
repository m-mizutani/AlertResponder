#!/bin/bash

cat $1  | jq . -c | sed -e "s@\"__VAR__\"@\$\{$2\}@g"
