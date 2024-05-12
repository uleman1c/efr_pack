package exchange

import tables "efr_pack/db"

func StartExchange(in map[string]interface{}) (map[string]interface{}, error) {

	res := map[string]interface{}{
		"success": true,
		"message": "",
		"result":  []map[string]interface{}{},
	}

	go doExchange()

	return res, nil

}

func doExchange() {

	tables.GetChangesFromCentral()

}
