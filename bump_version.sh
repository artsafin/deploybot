#!/bin/sh

REF=`cat .git/refs/heads/master`

sed -i "s#\".*\?\"#\"${REF}\"#" gitref.go

git commit -m "Bump version ${REF}" gitref.go
