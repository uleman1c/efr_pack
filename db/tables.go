package tables

var Constants = map[string]interface{}{

	"name": "constants",
	"fields": []map[string]interface{}{

		{"name": "id", "type": "char", "length": 36},
		{"name": "name", "type": "char", "length": 50},
		{"name": "value", "type": "char", "length": 200},
	},
}

var Users = map[string]interface{}{

	"name": "users",
	"fields": []map[string]interface{}{

		{"name": "id", "type": "char", "length": 36},
		{"name": "name", "type": "char", "length": 50},
		{"name": "pwd", "type": "char", "length": 20},
	},
}

var Products = map[string]interface{}{

	"name": "products",
	"fields": []map[string]interface{}{

		{"name": "id", "type": "char", "length": 36},
		{"name": "name", "type": "char", "length": 50},
		{"name": "full_name", "type": "char", "length": 150},
	},
}

var Containers = map[string]interface{}{

	"name": "containers",
	"fields": []map[string]interface{}{

		{"name": "id", "type": "char", "length": 36},
		{"name": "name", "type": "char", "length": 50},
	},
}

var Barcodes = map[string]interface{}{

	"name": "barcodes",
	"fields": []map[string]interface{}{

		{"name": "id", "type": "char", "length": 36},
		{"name": "barcode", "type": "char", "length": 150},
		{"name": "product_id", "type": "char", "length": 36},
	},
}

var Scanned = map[string]interface{}{

	"name": "scanned",
	"fields": []map[string]interface{}{

		{"name": "id", "type": "char", "length": 36},
		{"name": "date", "type": "char", "length": 14},
		{"name": "milliseconds", "type": "int", "length": 2},
		{"name": "barcode", "type": "char", "length": 150},
		{"name": "container_id", "type": "char", "length": 36},
		{"name": "menu_plan_id", "type": "char", "length": 36},
	},
}

var MenuPlans = map[string]interface{}{

	"name": "menu_plans",
	"fields": []map[string]interface{}{

		{"name": "id", "type": "char", "length": 36},
		{"name": "date", "type": "char", "length": 14},
		{"name": "number", "type": "char", "length": 14},
		{"name": "deletion_mark", "type": "int", "length": 1},
	},
	"tables": map[string]interface{}{
		"products": []map[string]interface{}{

			{"name": "id", "type": "char", "length": 36},
			{"name": "doc_id", "type": "char", "length": 36},
			{"name": "line_number", "type": "int", "length": 5},
			{"name": "product_id", "type": "char", "length": 36},
			{"name": "quantity", "type": "int", "length": 10},
		},
	},
}

var ObjectsToSend = map[string]interface{}{

	"name": "objectstosend",
	"fields": []map[string]interface{}{

		{"name": "id", "type": "char", "length": 36},
		{"name": "name", "type": "char", "length": 150},
		{"name": "object_id", "type": "char", "length": 36},
	},
}

var Tables = map[string]map[string]interface{}{

	"Constants":     Constants,
	"constants":     Constants,
	"Users":         Users,
	"Products":      Products,
	"Containers":    Containers,
	"containers":    Containers,
	"Barcodes":      Barcodes,
	"MenuPlans":     MenuPlans,
	"menu_plans":    MenuPlans,
	"Scanned":       Scanned,
	"scanned":       Scanned,
	"ObjectsToSend": ObjectsToSend,
	"objectstosend": ObjectsToSend,
}

func Copy(sourceTable map[string]interface{}) map[string]interface{} {

	tu := map[string]interface{}{}

	fields := sourceTable["fields"].([]map[string]interface{})

	tu["name"] = sourceTable["name"]
	tu["fields"] = make([]interface{}, len(fields))

	for i := 0; i < len(fields); i++ {

		field := fields[i]

		tu["fields"].([]interface{})[i] = map[string]interface{}{"name": field["name"], "type": field["type"], "length": field["length"]}

	}

	return tu
}

func CopyTables(table map[string]interface{}) map[string]interface{} {

	tu := map[string]interface{}{}

	for tableName, fields := range table["tables"].(map[string]interface{}) {

		ct := map[string]interface{}{

			"name":   tableName,
			"fields": fields,
		}

		tu[tableName] = Copy(ct)

	}

	return tu
}
