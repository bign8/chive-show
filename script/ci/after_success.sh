#!/bin/sh

set -e

# Upload coverage
codecov

# Deploy to appengine
if [ "$TRAVIS_PULL_REQUEST" != "false" ]; then
	echo "PR Build: Deploying to Appengine"

	# TODO: make this version number the tag if it's a tagged build

	export APP_DIR=.
	export APP_VERSION="pr-$TRAVIS_PULL_REQUEST"
	export GAE_DIR=../go_appengine

	python $GAE_DIR/appcfg.py --oauth2_refresh_token=$GAE_OAUTH_REFRESH_TOKEN update $APP_DIR -V $APP_VERSION
fi

echo "SUCCESS!!!!"
