#!/bin/sh

set -e

# export GAE_OAUTH=<your_oauth_token>
# export GAE_DIR=../go_appengine
# export APP_DIR=.
echo "PR# $TRAVIS_PULL_REQUEST"
# python $GAE_DIR/appcfg.py --oauth2_refresh_token=$GAE_OAUTH update $APP_DIR


codecov

echo "SUCCESS!!!!"

echo "start coverage"
cat coverage.txt
echo "end coverage"
