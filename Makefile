run:
	go run .


seed:
	sqlite3 ./db/data.db < ./db/seed.sql

build:
	go build -o bin/bikeparts .

e2e_test:
	go test -v ./e2e/...

