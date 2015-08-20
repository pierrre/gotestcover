#/bin/bash

coverage:
	gotestcover -coverprofile=cover.out github.com/pierre/gotestcover
	go tool cover -html=cover.out -o=cover.html
	
clean:
	-rm cover.html
	-rm cover.out
	gofmt -w .