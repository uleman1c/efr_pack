package main

import (
	tables "efr_pack/db"
	"efr_pack/server"
	"efr_pack/server/exchange"
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

	tables.UpdateConstant("exchange_in_progress", "0")

	exchange.StartExchange(nil)

	err = server.Start()
	if err != nil {
		return err
	}

	return err

}
