package main

import (
	"strconv"
	"database/sql"
	"fmt"
	"os"
	"bufio"
	"github.com/sohamkamani/go-dependency-injection-example/database"
	"github.com/sohamkamani/go-dependency-injection-example/service"
	
)

func main() {
	// Create a new DB connection
	connString := "dbname=<your main db name> sslmode=disable"
	db, _ := sql.Open("postgres", connString)

	// Create a store dependency with the db connection
	store := database.NewStore(db)
	// Create the service by injecting the store as a dependency
	service := &service.Service{Store: store}

	// The following code implements a simple command line app to read the ID as input
	// and output the validity of the result of the entry with that ID in the database
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan(){
		ID, _ := strconv.Atoi(scanner.Text())
		err := service.GetNumber(ID)
		if err != nil {
			fmt.Printf("result invalid: %v", err)
			continue
		}
		fmt.Println("result valid")
	}
}