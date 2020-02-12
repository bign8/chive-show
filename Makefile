serve:
	dev_appserver.py app.yaml --host=192.168.0.110 --admin_host=192.168.0.110
.PHONY:=serve

format:
	goimports -w -l -local github.com/bign8/chive-show .
.PHONY:=format

deploy:
	gcloud app deploy --version=test --project=crucial-alpha-706
.PHONY:=deploy
