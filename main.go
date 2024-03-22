package main

import (
	tables "efr_pack/db"
	"efr_pack/server"
	"log"
)

func main() {

	if err := run(); err != nil {

		log.Fatal(err)

	}

}

func run() error {

	var err error = nil

	err = tables.Init("sqlite")

	if err != nil {
		return err
	}

	err = tables.GetChangesFromCentral()
	if err != nil {
		return err
	}

	err = server.Start()
	if err != nil {
		return err
	}

	/* 	tu := tables.Copy(tables.Users)

	   	tu["fields"].([]interface{})[0].(map[string]interface{})["name"] = "sdrgjkhkjhdskj"

	   	fmt.Println(tu)
	   	fmt.Println(tables.Users)
	*/
	return err

}
