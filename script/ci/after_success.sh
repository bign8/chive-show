#!/bin/sh

set -e

# Upload coverage
codecov

# Deploy to appengine
export GAE_DIR=../go_appengine
export APP_DIR=.

# TODO: make this version number suck less (only deploy on PR builds, etc...)
export APP_VERSION="$TRAVIS_BRANCH-$TRAVIS_PULL_REQUEST"

python $GAE_DIR/appcfg.py --oauth2_refresh_token=$GAE_OAUTH_REFRESH_TOKEN update $APP_DIR -V $APP_VERSION

echo "SUCCESS!!!!"
