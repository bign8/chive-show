#! /bin/bash -ex

INT=dependabot/go_modules
MAIN=latest

git fetch --prune
git checkout $MAIN # move off of a conflicting branch?
git branch -D $INT || true
git checkout -t origin/$INT || git checkout -b $INT
git merge origin/$MAIN
git push origin $INT