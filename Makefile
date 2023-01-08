TAG:=jonamat/sqlite-rest

watch:
	gow run ./cmd

run:
	go run ./cmd

serve:
	./bin

build:
	go build -v -x -o ./bin/sqlite-rest ./cmd

build-static:
	CGO_ENABLED=0 && GOOS=linux && GOARCH=amd64 && go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o ./bin/sqlite-rest ./cmd/sqlite-rest

build-docker:
	docker build -t $(TAG) --no-cache .

push-docker:
	docker push $(TAG)