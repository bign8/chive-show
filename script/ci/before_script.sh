#!/bin/sh
set -e
echo "running goapp get to fetch dependencies..."

../go_appengine/goapp get golang.org/x/tools/cmd/cover
../go_appengine/goapp get ./app/...

echo "dependencies fetched."
exit 0

# - export HERE=`pwd`
# - cd $GOROOT/src && GOOS=linux GOARCH=amd64 ./make.bash --no-clean
# - export GOROOT=$HERE/go_appengine/goroot
# - go get github.com/mjibson/appstats
