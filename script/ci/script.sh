#!/bin/sh
echo "running tests and building..."

# ../go_appengine/goapp test ./app/...
# ../go_appengine/goapp build ./app/...

set -e
echo "" > coverage.txt

for d in $(find ./app/* -maxdepth 10 -type d); do
    if ls $d/*.go &> /dev/null; then
        go test  -coverprofile=profile.out -covermode=atomic $d
        if [ -f profile.out ]; then
            cat profile.out >> coverage.txt
            echo '<<<<<< EOF' >> coverage.txt
            rm profile.out
        fi
    fi
done
