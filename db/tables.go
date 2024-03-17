package tables

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

var MenuPlans = map[string]interface{}{

	"name": "menu_plans",
	"fields": []map[string]interface{}{

		{"name": "id", "type": "char", "length": 36},
		{"name": "date", "type": "char", "length": 14},
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

var Tables = map[string]map[string]interface{}{

	"Users":     Users,
	"Products":  Products,
	"MenuPlans": MenuPlans,
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
