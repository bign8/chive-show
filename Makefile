serve:
	GOOGLE_APPLICATION_CREDENTIALS=.github/service-account.json \
	GOOGLE_CLOUD_PROJECT=crucial-alpha-706 \
	go run main.go
.PHONY:=serve

format:
	goimports -w -l -local github.com/bign8/chive-show .
.PHONY:=format

deploy: .github/index.xml
	gcloud app deploy --version=test --project=crucial-alpha-706 --no-promote
.PHONY:=deploy

.github/index.xml: index.yaml
	gcloud datastore indexes create index.yaml
	touch .github/index.xml

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

int:
	@echo "This requires 'make serve' to be running"
	go test . -v -target http://localhost:8080
.PHONY:=int

event:
	curl \
		-i \
		-X POST \
		-H "Accept: application/vnd.github.v3+json" \
		https://api.github.com/repos/bign8/chive-show/dispatches \
		-d '{"event_type":"dependabot"}'