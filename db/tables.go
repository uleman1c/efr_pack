package tables

var Users = map[string]interface{}{

	"name": "users",
	"fields": []interface{}{

		map[string]interface{}{"name": "id", "type": "char", "length": 36},
		map[string]interface{}{"name": "name", "type": "char", "length": 50},
		map[string]interface{}{"name": "pwd", "type": "char", "length": 20},
	},
}

var Products = map[string]interface{}{

	"name": "products",
	"fields": []interface{}{

		map[string]interface{}{"name": "id", "type": "char", "length": 36},
		map[string]interface{}{"name": "name", "type": "char", "length": 50},
		map[string]interface{}{"name": "full_name", "type": "char", "length": 150},
	},
}

var MenuPlans = map[string]interface{}{

	"name": "menu_plans",
	"fields": []interface{}{

		map[string]interface{}{"name": "id", "type": "char", "length": 36},
		map[string]interface{}{"name": "date", "type": "char", "length": 14},
	},
	"tables": map[string]interface{}{
		"products": []interface{}{

			map[string]interface{}{"name": "id", "type": "char", "length": 36},
			map[string]interface{}{"name": "doc_id", "type": "char", "length": 36},
			map[string]interface{}{"name": "line_number", "type": "int", "length": 5},
			map[string]interface{}{"name": "product_id", "type": "char", "length": 36},
			map[string]interface{}{"name": "quantity", "type": "int", "length": 10},
		},
	},
}

var Tables = []interface{}{

	Users, Products, MenuPlans,
}

func Copy(sourceTable map[string]interface{}) map[string]interface{} {

	tu := map[string]interface{}{}

	fields := sourceTable["fields"].([]interface{})

	tu["name"] = sourceTable["name"]
	tu["fields"] = make([]interface{}, len(fields))

	for i := 0; i < len(fields); i++ {

		field := fields[i].(map[string]interface{})

		tu["fields"].([]interface{})[i] = map[string]interface{}{"name": field["name"], "type": field["type"], "length": field["length"]}

	}

	return tu
}
