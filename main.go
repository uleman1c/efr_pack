package main

import (
	tables "efr_pack/db"
	"fmt"
	"log"
)

func main() {

	if err := run(); err != nil {

		log.Fatal(err)

	}

}

func run() error {

	tu := tables.Copy(tables.Users)

	tu["fields"].([]interface{})[0].(map[string]interface{})["name"] = "sdrgjkhkjhdskj"

	fmt.Println(tu)
	fmt.Println(tables.Users)

	tables.Init("sqlite")

	return nil

}
