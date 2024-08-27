#!/usr/bin/env bash

# get script dir
script_dir=$( cd `dirname ${BASH_SOURCE[0]}` >/dev/null 2>&1 ; pwd -P )

echo "Go ..."

goenv install 1.22.5 --skip-existing
   
# create .go-version
goenv local 1.22.5

goenv versions

# use Go from .go-version for local development
eval "$(goenv init -)"
