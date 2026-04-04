run: 
	go run .

build:
	go build -o bin/bikeparts .

db_clear: 
	rm -rf ./db/data.db
