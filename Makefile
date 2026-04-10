run: 
	go run .

build:
	go build -o bin/bikeparts .

e2e_test:
	go test -v ./e2e/...

db_clear: 
	rm -rf ./db/data.db
