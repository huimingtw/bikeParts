run:
	go run .

seed:
	sqlite3 ./db/data.db < ./db/seed.sql

build:
	go build -o bin/bikeparts .

build_windows:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc \
	go build -ldflags "-H windowsgui" -o bin/bikeparts.exe .

e2e_test:
	go test -v ./e2e/...

release:
	@read -p "Tag (e.g. v1.0.0): " tag; \
	git tag $$tag && git push origin $$tag
