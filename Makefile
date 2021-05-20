serve:
	GOOGLE_APPLICATION_CREDENTIALS=.github/service-account.json go run main.go
.PHONY:=serve

format:
	goimports -w -l -local github.com/bign8/chive-show .
.PHONY:=format

deploy:
	gcloud app deploy --version=test --project=crucial-alpha-706 --no-promote
.PHONY:=deploy

deps:
	brew install entr
.PHONY:=deps

watch:
	# http://eradman.com/entrproject/
	ls *.go api/*.go cron/*.go models/*.go keycache/*.go | entr -r make format serve
.PHONY:=watch

cover:
	go test -cover -coverprofile=coverage.cov ./...
	go tool cover -html=coverage.cov -o static/coverage.html
.PHONY:=cover