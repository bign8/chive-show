#!/bin/sh

set -e

# Upload coverage
codecov

# Deploy to appengine
export GAE_DIR=../go_appengine
export APP_DIR=.
echo "PR# $TRAVIS_PULL_REQUEST"
# TODO: use latest commit in part of deploy-id
python $GAE_DIR/appcfg.py --oauth2_refresh_token=$GAE_OAUTH_REFRESH_TOKEN update $APP_DIR

echo "SUCCESS!!!!"
