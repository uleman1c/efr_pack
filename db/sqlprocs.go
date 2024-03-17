package tables

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var Db *sql.DB = nil

var SqlType = ""

func Init(sq string) error {

	SqlType = sq

	var err error = nil

	Db, err = sql.Open("sqlite3", "./efr_pack.db?cache=shared")
	if err != nil {
		return err
	}

	Db.SetMaxOpenConns(1)

	tx, err := Db.Begin()
	if err != nil {
		return err
	}

	defer tx.Commit()

	if false {

		statement, err := tx.Prepare("DROP TABLE outcome_orders_products_by_orders")

		if err != nil {
			return err
		}

		statement.Exec()

	}

	for _, table := range Tables {

		statement, err := tx.Prepare(GetCreateQuery(table))

		if err != nil {
			return err
		}

		_, err = statement.Exec()
		if err != nil {
			return err
		}

		subTables := table["tables"]

		if subTables != nil {

			for subTableName, subTableFields := range subTables.(map[string]interface{}) {

				fmt.Println(subTableName, subTableFields)

				statement, err := tx.Prepare(GetCreateQuery(map[string]interface{}{
					"name":   table["name"].(string) + "_" + subTableName,
					"fields": subTableFields,
				}))

				if err != nil {
					return err
				}

				_, err = statement.Exec()
				if err != nil {
					return err
				}

			}

		}

	}

	return err

}

func GetCreateQuery(table map[string]interface{}) string {

	result := ""

	switch SqlType {
	case "sqlite":

		fields := []string{}

		for _, field := range table["fields"].([]map[string]interface{}) {

			key := ""
			if field["name"] == "id" {
				key = "PRIMARY KEY"
			}

			fields = append(fields, field["name"].(string)+" "+field["type"].(string)+" "+key)

		}

		result = "CREATE Table IF NOT EXISTS " + table["name"].(string) + " (" + strings.Join(fields, ",") + ")"

	}

	return result

}

func GetSelectQuery(table map[string]interface{}, filters []map[string]interface{}) string {

	result := ""

	tableName := table["name"].(string)

	switch SqlType {
	case "sqlite":

		fields := []string{}

		for _, field := range table["fields"].([]map[string]interface{}) {

			fields = append(fields, tableName+"."+field["name"].(string))

		}

		filterStrings := []string{}

		for _, filter := range filters {

			filterStrings = append(filterStrings, filter["text"].(string))

		}

		result = "SELECT " + strings.Join(fields, ",") + " FROM " + tableName

		if len(filterStrings) > 0 {

			result += " WHERE " + strings.Join(filterStrings, " and ")

		}

	}

	return result

}

func GetInsertQuery(table map[string]interface{}) string {

	result := ""

	switch SqlType {
	case "sqlite":

		fields := []string{}
		values := []string{}

		for _, field := range table["fields"].([]map[string]interface{}) {

			fields = append(fields, field["name"].(string))
			values = append(values, "?")

		}

		result = "INSERT INTO " + table["name"].(string) + " (" + strings.Join(fields, ",") + ") VALUES (" + strings.Join(values, ",") + ")"

	}

	return result

}

func GetDeleteQuery(tableName string, filters []map[string]interface{}) string {

	result := ""

	switch SqlType {
	case "sqlite":

		result = "DELETE FROM " + tableName

		filterStrings := []string{}

		for _, filter := range filters {

			filterStrings = append(filterStrings, filter["text"].(string))

		}

		if len(filterStrings) > 0 {

			result += " WHERE " + strings.Join(filterStrings, " and ")

		}

	}

	return result

}

func GetInsertQuerySub(tableName string, params map[string]interface{}) (string, []interface{}) {

	result := ""

	sqlParams := []interface{}{}

	switch SqlType {
	case "sqlite":

		fields := []string{}
		values := []string{}

		for field, value := range params {

			fields = append(fields, field)
			values = append(values, "?")

			sqlParams = append(sqlParams, value)

		}

		result = "INSERT INTO " + tableName + " (" + strings.Join(fields, ",") + ") VALUES (" + strings.Join(values, ",") + ")"

	}

	return result, sqlParams

}

/*
func GetConstantValue(name string) (*string, error) {

	result := ""

	tx, err := Db.Begin()
	if err != nil {
		fmt.Printf("begin. Exec error=%s", err)
		return &result, err
	}

	defer tx.Commit()

	constFilter := []tables.Filter{}
	constFilter = append(constFilter, tables.Filter{Text: "name = ?", TypE: "text", Parameter: tables.Parameter{Name: name, Value: name}})

	statement, err := tx.Prepare(tables.GetSelectQuery(&tables.Constants, constFilter))

	if err != nil {
		fmt.Println(err)
		return &result, err
	}

	df := tables.GetParamsFromFilter(constFilter)
	fmt.Println(df)

	rows, err := statement.Query(df[0])

	if err != nil {
		fmt.Println(err)
		return &result, err
	}

	defer rows.Close()

	if rows.Next() {

		id := ""

		name := ""

		rows.Scan(&id, &name, &result)

	}

	return &result, nil

}

func UpdateConstant(name string, value any) {

	tx, err := Db.Begin()
	if err != nil {
		fmt.Printf("begin. Exec error=%s", err)
		return
	}

	defer tx.Commit()

	constFilter := []tables.Filter{}
	constFilter = append(constFilter, tables.Filter{Text: "name = ?", TypE: "text", Parameter: tables.Parameter{Name: name, Value: value}})

	statement, err := tx.Prepare(tables.GetSelectQuery(&tables.Constants, constFilter))

	if err != nil {
		fmt.Println(err)
	}

	df := tables.GetParamsFromFilter(constFilter)
	//fmt.Println(df)

	rows, err := statement.Query(df[0])

	if err != nil {
		fmt.Println(err)
	}

	defer rows.Close()

	if !rows.Next() {

		statement, err = tx.Prepare(tables.GetInsertQuery(&tables.Constants))

		if err != nil {
			fmt.Println(err)
		}

		rows, err := statement.Exec(uuid.New().String(), constFilter[0].Parameter.Name, constFilter[0].Parameter.Value.(string))

		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(rows)
	}

}

func existsById(table *tables.Table, id string) bool {

	tx, err := Db.Begin()
	if err != nil {
		fmt.Printf("begin. Exec error=%s", err)

	}

	defer tx.Commit()

	constFilter := []tables.Filter{}
	constFilter = append(constFilter, tables.Filter{Text: "id = ?", TypE: "text", Parameter: tables.Parameter{Name: "id", Value: id}})

	statement, err := tx.Prepare(tables.GetSelectQuery(table, constFilter))

	if err != nil {
		fmt.Println(err)
	}

	df := tables.GetParamsFromFilter(constFilter)
	fmt.Println(df)

	rows, err := statement.Query(id)

	if err != nil {
		fmt.Println(err)
	}

	exist := rows.Next()

	defer rows.Close()

	return exist

}

func UpdateWarehouseUser(id string, deletionMark bool, name, pwd string) error {

	if !existsById(&tables.WarehouseUsers, id) {

		tx, err := Db.Begin()
		if err != nil {
			return err
		}

		defer tx.Commit()

		statement, err := tx.Prepare(tables.GetInsertQuery(&tables.WarehouseUsers))

		if err != nil {
			return err
		}

		dm := 0

		if deletionMark {
			dm = 1
		}

		_, err = statement.Exec(id, dm, name, pwd)

		if err != nil {
			return err
		}

	}

	return nil

}

func UpdateProduct(id string, deletionMark bool, artikul, name string, weight int) error {

	if !existsById(&tables.Products, id) {

		tx, err := Db.Begin()
		if err != nil {
			return err
		}

		defer tx.Commit()

		statement, err := tx.Prepare(tables.GetInsertQuery(&tables.Products))

		if err != nil {
			return err
		}

		dm := 0

		if deletionMark {
			dm = 1
		}

		_, err = statement.Exec(id, dm, artikul, name, weight)

		if err != nil {
			return err
		}

	}

	return nil

}

func UpdateUnit(id string, deletionMark bool, product_id, name string, coefficient int) error {

	table := &tables.Units

	if !existsById(table, id) {

		tx, err := Db.Begin()
		if err != nil {
			return err
		}

		defer tx.Commit()

		statement, err := tx.Prepare(tables.GetInsertQuery(table))

		if err != nil {
			return err
		}

		dm := 0

		if deletionMark {
			dm = 1
		}

		_, err = statement.Exec(id, dm, product_id, name, coefficient)

		if err != nil {
			return err
		}

	}

	return nil

}

func UpdateCharacteristic(id string, deletionMark bool, name string) error {

	table := &tables.Characteristics

	if !existsById(table, id) {

		tx, err := Db.Begin()
		if err != nil {
			return err
		}

		defer tx.Commit()

		statement, err := tx.Prepare(tables.GetInsertQuery(table))

		if err != nil {
			return err
		}

		dm := 0

		if deletionMark {
			dm = 1
		}

		_, err = statement.Exec(id, dm, name)

		if err != nil {
			return err
		}

	}

	return nil

}

func UpdateCell(id string, deletionMark bool, name, section, line, rack, level string) error {

	table := &tables.Cells

	if !existsById(table, id) {

		tx, err := Db.Begin()
		if err != nil {
			return err
		}

		defer tx.Commit()

		statement, err := tx.Prepare(tables.GetInsertQuery(table))

		if err != nil {
			return err
		}

		dm := 0

		if deletionMark {
			dm = 1
		}

		_, err = statement.Exec(id, dm, name, section, line, rack, level)

		if err != nil {
			return err
		}

	}

	return nil

}

func UpdateContainer(id string, deletionMark bool, name, production_date string) error {

	table := &tables.Containers

	if !existsById(table, id) {

		tx, err := Db.Begin()
		if err != nil {
			return err
		}

		defer tx.Commit()

		statement, err := tx.Prepare(tables.GetInsertQuery(table))

		if err != nil {
			return err
		}

		dm := 0

		if deletionMark {
			dm = 1
		}

		_, err = statement.Exec(id, dm, name, production_date)

		if err != nil {
			return err
		}

	}

	return nil

}

func UpdateObject(table *tables.Table, params map[string]interface{}) error {

	if !existsById(table, params["id"].(string)) {

		tx, err := Db.Begin()
		if err != nil {
			return err
		}

		statement, err := tx.Prepare(tables.GetInsertQuery(table))

		if err != nil {
			return err
		}

		sqlParams := []interface{}{}
		for _, tableField := range tables.GetFields(table) {

			sqlParams = append(sqlParams, params[tableField])

		}

		_, err = statement.Exec(sqlParams...)

		if err != nil {
			return err
		}

		tx.Commit()

	}

	if _, ok := params["tables"]; ok {

		if err := UpdateObjectTables(tables.GetName(table), params); err != nil {

			return err

		}

	}

	return nil

}

func UpdateObjectTables(mainTableName string, params map[string]interface{}) error {

	tx, err := Db.Begin()
	if err != nil {
		return err
	}

	defer tx.Commit()

	for tableName, tableData := range params["tables"].(map[string]interface{}) {

		filters := []interface{}{
			map[string]interface{}{
				"text": "doc_id = ?",
				"params": map[string]interface{}{
					"name":  "doc_id",
					"value": params["id"],
				},
			},
		}

		q := tables.GetDeleteQueryByStr(mainTableName+"_"+tableName, filters)

		statement, err := tx.Prepare(q)

		if err != nil {
			return err
		}

		_, err = statement.Exec(params["id"])

		if err != nil {
			return err
		}

		for _, line := range tableData.(map[string]interface{})["lines"].([]interface{}) {

			q, sqlParams := tables.GetInsertQuerySub(mainTableName+"_"+tableName, line.(map[string]interface{}))

			statement, err := tx.Prepare(q)

			if err != nil {
				return err
			}

			_, err = statement.Exec(sqlParams...)

			if err != nil {
				return err
			}

		}

	}
	return nil
}

type Constant struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Constants []*Constant

func GetQueryResult(queryStr string, fields []string, params []interface{}) ([]interface{}, error) {

	result := []interface{}{}

	tx, err := Db.Begin()
	if err != nil {
		return result, err
	}

	defer tx.Commit()

	statement, err := tx.Prepare(queryStr)
	if err != nil {
		return result, err
	}

	var rows *sql.Rows = nil

	rows, err = statement.Query(params...)

	lenRecord := len(fields)

	for rows.Next() {

		item := make([]interface{}, lenRecord)

		switch lenRecord {
		case 1:
			{

				rows.Scan(&item[0])

			}
		case 2:
			{

				rows.Scan(&item[0], &item[1])

			}
		case 3:
			{

				rows.Scan(&item[0], &item[1], &item[2])

			}
		case 4:
			{

				rows.Scan(&item[0], &item[1], &item[2], &item[3])

			}
		case 5:
			{

				rows.Scan(&item[0], &item[1], &item[2], &item[3], &item[4])

			}
		case 6:
			{

				rows.Scan(&item[0], &item[1], &item[2], &item[3], &item[4], &item[5])

			}
		case 7:
			{

				rows.Scan(&item[0], &item[1], &item[2], &item[3], &item[4], &item[5], &item[6])

			}
		case 8:
			{

				rows.Scan(&item[0], &item[1], &item[2], &item[3], &item[4], &item[5], &item[6], &item[7])

			}
		case 9:
			{

				rows.Scan(&item[0], &item[1], &item[2], &item[3], &item[4], &item[5], &item[6], &item[7], &item[8])

			}
		case 10:
			{

				rows.Scan(&item[0], &item[1], &item[2], &item[3], &item[4], &item[5], &item[6], &item[7], &item[8], &item[9])

			}
		}

		rec := map[string]interface{}{}

		for i, field := range fields {

			rec[field] = item[i]

		}

		result = append(result, rec)

	}

	return result, err

}

func GetTable(table *tables.Table) []Constant {

	tx, err := Db.Begin()
	if err != nil {
		fmt.Printf("begin. Exec error=%s", err)
		return []Constant{}
	}

	defer tx.Commit()

	result := make([]Constant, 0)

	constFilter := []tables.Filter{}

	statement, _ := tx.Prepare(tables.GetSelectQuery(table, constFilter))

	//df := tables.GetParamsFromFilter(constFilter)

	//if len(df) == 0 {

	rows, err := statement.Query()

	// } else {

	// 	rows, err := statement.Query(df[0])

	// }

	if err != nil {
		fmt.Println(err)
	}

	defer rows.Close()

	//item := []any{}
	for rows.Next() {

		item := Constant{}
		result = append(result, item)

		rows.Scan(&result[len(result)-1].Id, &result[len(result)-1].Name, &result[len(result)-1].Value)

	}

	return result

}

func GetTableWithParams(table *tables.Table, filter []tables.Filter, top, order, group string) []map[string]interface{} {

	result := make([]map[string]interface{}, 0)

	tx, err := Db.Begin()
	if err != nil {
		fmt.Printf("begin. Exec error=%s", err)
		return result
	}

	defer tx.Commit()

	statement, err := tx.Prepare(tables.GetSelectQueryTOG(table, filter, top, order, group))

	cp := tables.GetParamsValuesFromFilter(filter)

	var rows *sql.Rows = nil

	rows, err = statement.Query(cp...)

	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	lenRecord := tables.FieldsCount(table)

	for rows.Next() {

		item := make([]interface{}, lenRecord)

		switch lenRecord {
		case 1:
			{

				rows.Scan(&item[0])

			}
		case 2:
			{

				rows.Scan(&item[0], &item[1])

			}
		case 3:
			{

				rows.Scan(&item[0], &item[1], &item[2])

			}
		case 4:
			{

				rows.Scan(&item[0], &item[1], &item[2], &item[3])

			}
		case 5:
			{

				rows.Scan(&item[0], &item[1], &item[2], &item[3], &item[4])

			}
		case 6:
			{

				rows.Scan(&item[0], &item[1], &item[2], &item[3], &item[4], &item[5])

			}
		case 7:
			{

				rows.Scan(&item[0], &item[1], &item[2], &item[3], &item[4], &item[5], &item[6])

			}
		case 8:
			{

				rows.Scan(&item[0], &item[1], &item[2], &item[3], &item[4], &item[5], &item[6], &item[7])

			}
		case 9:
			{

				rows.Scan(&item[0], &item[1], &item[2], &item[3], &item[4], &item[5], &item[6], &item[7], &item[8])

			}
		case 10:
			{

				rows.Scan(&item[0], &item[1], &item[2], &item[3], &item[4], &item[5], &item[6], &item[7], &item[8], &item[9])

			}
		case 11:
			{

				rows.Scan(&item[0], &item[1], &item[2], &item[3], &item[4], &item[5], &item[6], &item[7], &item[8], &item[9], &item[10])

			}
		case 12:
			{

				rows.Scan(&item[0], &item[1], &item[2], &item[3], &item[4], &item[5], &item[6], &item[7], &item[8], &item[9], &item[10], &item[11])

			}
		case 13:
			{

				rows.Scan(&item[0], &item[1], &item[2], &item[3], &item[4], &item[5], &item[6], &item[7], &item[8], &item[9], &item[10], &item[11], &item[12])

			}
		case 14:
			{

				rows.Scan(&item[0], &item[1], &item[2], &item[3], &item[4], &item[5], &item[6], &item[7], &item[8], &item[9], &item[10], &item[11], &item[12], &item[13])

			}
		case 15:
			{

				rows.Scan(&item[0], &item[1], &item[2], &item[3], &item[4], &item[5], &item[6], &item[7], &item[8], &item[9],
					&item[10], &item[11], &item[12], &item[13], &item[14])

			}
		case 16:
			{

				rows.Scan(&item[0], &item[1], &item[2], &item[3], &item[4], &item[5], &item[6], &item[7], &item[8], &item[9],
					&item[10], &item[11], &item[12], &item[13], &item[14], &item[15])

			}
		case 17:
			{

				rows.Scan(&item[0], &item[1], &item[2], &item[3], &item[4], &item[5], &item[6], &item[7], &item[8], &item[9],
					&item[10], &item[11], &item[12], &item[13], &item[14], &item[15], &item[16])

			}
		case 18:
			{

				rows.Scan(&item[0], &item[1], &item[2], &item[3], &item[4], &item[5], &item[6], &item[7], &item[8], &item[9],
					&item[10], &item[11], &item[12], &item[13], &item[14], &item[15], &item[16], &item[17])

			}
		case 19:
			{

				rows.Scan(&item[0], &item[1], &item[2], &item[3], &item[4], &item[5], &item[6], &item[7], &item[8], &item[9],
					&item[10], &item[11], &item[12], &item[13], &item[14], &item[15], &item[16], &item[17], &item[18])

			}
		case 20:
			{

				rows.Scan(&item[0], &item[1], &item[2], &item[3], &item[4], &item[5], &item[6], &item[7], &item[8], &item[9],
					&item[10], &item[11], &item[12], &item[13], &item[14], &item[15], &item[16], &item[17], &item[18], &item[19])

			}
		}

		rec := map[string]interface{}{}

		for i, field := range tables.GetFields(table) {

			rec[field] = item[i]

		}

		result = append(result, rec)
		//fmt.Println(result)

	}

	return result

}

func GetTableData(table map[string]interface{}) []map[string]interface{} {

	result := []map[string]interface{}{}

	filter := []tables.Filter{}

	filterStr := ""
	_, ok := table["filter"]
	if ok {
		filterStr = table["filter"].(string)
	}

	if filterStr != "" {

		kv := strings.Split(filterStr, " eq ")

		filter = append(filter, tables.Filter{Text: kv[0] + " = ?", TypE: "TEXT", Parameter: tables.Parameter{Name: kv[0], Value: kv[1]}})

	}

	top := ""
	_, ok = table["top"]
	if ok {
		top = table["top"].(string)
	}

	order := ""
	_, ok = table["order"]
	if ok {
		order = table["order"].(string)
	}

	group := ""
	_, ok = table["group"]
	if ok {
		group = table["group"].(string)
	}

	var tTable *tables.Table = nil

	switch table["name"].(string) {
	case "Constants":
		tTable = &tables.Constants
	case "Users":
		tTable = &tables.Users
	case "WarehouseUsers":
		tTable = &tables.WarehouseUsers
	case "Products":
		tTable = &tables.Products
	case "Characteristics":
		tTable = &tables.Characteristics
	case "Units":
		tTable = &tables.Units
	case "Cells":
		tTable = &tables.Cells
	case "Containers":
		tTable = &tables.Containers
	case "Leftovers":
		tTable = &tables.Leftovers
	case "DctMenuSettings":
		tTable = &tables.DctMenuSettings
	case "AccessGroups":
		tTable = &tables.AccessGroups
	case "Navigation":
		tTable = &tables.Navigation
	case "Partners":
		tTable = &tables.Partners
	case "Contractors":
		tTable = &tables.Contractors
	case "CustomerOrders":
		tTable = &tables.CustomerOrders
	case "OutcomeOrders":
		tTable = &tables.OutcomeOrders
	}

	if tTable != nil {

		result = GetTableWithParams(tTable, filter, top, order, group)
	}

	return result

}

type ErpSkladWhdctChanges struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Result  []any  `json:"result"`
}

type ErpSkladWhdctChangesResponse struct {
	ErpSkladWhdctChanges ErpSkladWhdctChanges
}

type ErpResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Responses []any  `json:"responses"`
}

type ChangeResponse struct {
	Type, Name, Id string
}

func GetWarehouseUsersFromCentral() {

	whid, _ := GetConstantValue("warehouse_id")

	changesExist := true

	for changesExist {

		client := &http.Client{}
		req, _ := http.NewRequest(
			"GET", "https://ow.apx-service.ru/tm_po/hs/dta/obj?request=getErpSkladWhdctChanges&warehouse_id="+*whid, nil,
		)

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			return
		}

		defer resp.Body.Close()

		dec := json.NewDecoder(resp.Body)
		dec.DisallowUnknownFields()

		var p ErpResponse
		err = dec.Decode(&p)
		if err != nil {

			fmt.Println(err)
			return

		}

		ErpSkladWhdctChanges := p.Responses[0].(map[string]interface{})["ErpSkladWhdctChanges"].(map[string]interface{})

		Result := ErpSkladWhdctChanges["result"].([]interface{})

		if len(Result) == 0 {

			changesExist = false

		} else {

			changesResponse := []ChangeResponse{}

			for i := 0; i < len(Result); i++ {

				change := Result[i].(map[string]interface{})

				changeResponse, err := loadObject(change)

				if err == nil {

					changesResponse = append(changesResponse, changeResponse)

				}

			}

			resMap := map[string]interface{}{}
			resMap["request"] = "setErpSkladWhdctChanges"
			resMap["parameters"] = map[string]interface{}{
				"warehouse_id": *whid,
				"changes":      changesResponse,
			}

			jsonResponse, _ := json.Marshal(resMap)

			req, _ = http.NewRequest(
				"POST", "https://ow.apx-service.ru/tm_po/hs/dta/obj",
				bytes.NewBuffer(jsonResponse),
			)
			req.Header.Set("Content-Type", "application/json; charset=UTF-8")

			_, err = client.Do(req)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

func loadObject(change map[string]interface{}) (ChangeResponse, error) {

	var err error = nil

	switch change["Тип"] {

	case "Справочник":
		{

			switch change["Вид"] {

			case "СкладскиеСотрудники":
				{

					id := change["Ссылка"].(string)

					err = UpdateWarehouseUser(id,
						change["ПометкаУдаления"].(bool),
						change["Наименование"].(string),
						change["Реквизиты"].(map[string]interface{})["ПинКод"].(string))

					return ChangeResponse{"Справочник", "СкладскиеСотрудники", id}, err

				}

			case "Номенклатура":
				{

					id := change["Ссылка"].(string)

					reqvisits := change["Реквизиты"].(map[string]interface{})

					err = UpdateProduct(id,
						change["ПометкаУдаления"].(bool),
						reqvisits["Артикул"].(string),
						change["Наименование"].(string),
						int(math.Round(reqvisits["ВесЧислитель"].(float64)*1000)))

					return ChangeResponse{"Справочник", "Номенклатура", id}, err

				}

			case "УпаковкиЕдиницыИзмерения":
				{

					id := change["Ссылка"].(string)

					err = UpdateUnit(id,
						change["ПометкаУдаления"].(bool),
						change["Владелец"].(map[string]interface{})["Ссылка"].(string),
						change["Наименование"].(string),
						int(change["Реквизиты"].(map[string]interface{})["Числитель"].(float64)))

					return ChangeResponse{"Справочник", "УпаковкиЕдиницыИзмерения", id}, err

				}

			case "ХарактеристикиНоменклатуры":
				{

					id := change["Ссылка"].(string)

					err = UpdateCharacteristic(id,
						change["ПометкаУдаления"].(bool),
						change["Наименование"].(string))

					return ChangeResponse{"Справочник", "ХарактеристикиНоменклатуры", id}, err

				}

			case "СкладскиеЯчейки":
				{

					id := change["Ссылка"].(string)

					err = UpdateCell(id,
						change["ПометкаУдаления"].(bool),
						change["Наименование"].(string),
						change["Реквизиты"].(map[string]interface{})["Секция"].(string),
						change["Реквизиты"].(map[string]interface{})["Линия"].(string),
						change["Реквизиты"].(map[string]interface{})["Стеллаж"].(string),
						change["Реквизиты"].(map[string]interface{})["Ярус"].(string))

					return ChangeResponse{"Справочник", "СкладскиеЯчейки", id}, err

				}

			case "СкладскиеКонтейнеры":
				{

					id := change["Ссылка"].(string)

					err = UpdateContainer(id,
						change["ПометкаУдаления"].(bool),
						change["Наименование"].(string),
						change["Реквизиты"].(map[string]interface{})["ДатаПроизводства"].(string))

					return ChangeResponse{"Справочник", "СкладскиеКонтейнеры", id}, err

				}

			case "НастройкиМенюТСД":
				{

					id := change["Ссылка"].(string)

					params := map[string]interface{}{}
					params["id"] = id
					params["is_folder"] = change["ЭтоГруппа"]
					params["name"] = change["Наименование"]
					params["navigation"] = change["Реквизиты"].(map[string]interface{})["Навигация"]
					params["sort_order"] = change["Реквизиты"].(map[string]interface{})["Порядок"]

					err = UpdateObject(&tables.DctMenuSettings, params)

					return ChangeResponse{"Справочник", "НастройкиМенюТСД", id}, err

				}

			case "ГруппыДоступаКНавигацииТСД":
				{

					id := change["Ссылка"].(string)

					params := map[string]interface{}{}
					params["id"] = id
					params["name"] = change["Наименование"]
					params["tables"] = tables.CopyTables(&tables.AccessGroups)

					wuLines := []interface{}{}

					lineCount := 0
					for _, line := range change["ТабличныеЧасти"].(map[string]interface{})["СкладскиеСотрудники"].([]interface{}) {

						addLine := map[string]interface{}{

							"id":                uuid.New().String(),
							"doc_id":            id,
							"line_number":       lineCount,
							"warehouse_user_id": line.(map[string]interface{})["СкладскойСотрудник"].(map[string]interface{})["Ссылка"],
						}

						wuLines = append(wuLines, addLine)

						lineCount += 1

					}

					params["tables"].(map[string]interface{})["warehouse_users"].(map[string]interface{})["lines"] = wuLines

					navLines := []interface{}{}

					lineCount = 0
					for _, line := range change["ТабличныеЧасти"].(map[string]interface{})["НавигацияТСД"].([]interface{}) {

						addLine := map[string]interface{}{

							"id":            uuid.New().String(),
							"doc_id":        id,
							"line_number":   lineCount,
							"navigation_id": line.(map[string]interface{})["НавигацияТСД"].(map[string]interface{})["Ссылка"],
						}

						navLines = append(navLines, addLine)

						lineCount += 1

					}

					params["tables"].(map[string]interface{})["navigation"].(map[string]interface{})["lines"] = navLines

					err = UpdateObject(&tables.AccessGroups, params)

					return ChangeResponse{"Справочник", "ГруппыДоступаКНавигацииТСД", id}, err

				}

			case "НавигацияТСД":
				{

					id := change["Ссылка"].(string)

					params := map[string]interface{}{}
					params["id"] = id
					params["name"] = change["Наименование"]

					err = UpdateObject(&tables.Navigation, params)

					return ChangeResponse{"Справочник", "НавигацияТСД", id}, err

				}

			case "Партнеры":
				{

					id := change["Ссылка"].(string)

					params := map[string]interface{}{}
					params["id"] = id
					params["name"] = change["Наименование"]
					params["full_name"] = change["Реквизиты"].(map[string]interface{})["НаименованиеПолное"]

					err = UpdateObject(&tables.Partners, params)

					return ChangeResponse{"Справочник", "Партнеры", id}, err

				}

			case "Контрагенты":
				{

					id := change["Ссылка"].(string)

					params := map[string]interface{}{}
					params["id"] = id
					params["name"] = change["Наименование"]
					params["full_name"] = change["Реквизиты"].(map[string]interface{})["НаименованиеПолное"]
					params["partner_id"] = change["Реквизиты"].(map[string]interface{})["Партнер"].(map[string]interface{})["Ссылка"]

					err = UpdateObject(&tables.Contractors, params)

					return ChangeResponse{"Справочник", "Контрагенты", id}, err

				}

			}

		}

	case "Документ":
		{

			switch change["Вид"] {

			case "ЗаказКлиента":
				{

					id := change["Ссылка"].(string)

					requis := change["Реквизиты"].(map[string]interface{})

					params := map[string]interface{}{}
					params["id"] = id
					params["deletion_mark"] = change["ПометкаУдаления"]
					params["date"] = change["Дата"]
					params["number"] = change["Номер"]
					params["customer_number"] = requis["НомерПоДаннымКлиента"]
					params["partner_id"] = requis["Партнер"].(map[string]interface{})["Ссылка"]
					params["contractor_id"] = requis["Контрагент"].(map[string]interface{})["Ссылка"]
					params["load_date"] = requis["ДатаЗагрузки"]
					params["shipping_date"] = requis["ДатаЗагрузки"]
					params["manager_id"] = requis["Менеджер"].(map[string]interface{})["Ссылка"]
					params["weight"] = requis["Вес"]
					params["safe_keeping"] = requis["ОтгрузкаСОтветственногоХранения"]
					params["comment"] = requis["Комментарий"]

					err = UpdateObject(&tables.CustomerOrders, params)

					return ChangeResponse{"Документ", "ЗаказКлиента", id}, err

				}

			case "РасходныйОрдерНаТовары":
				{

					id := change["Ссылка"].(string)

					params := map[string]interface{}{}
					params["id"] = id
					params["deletion_mark"] = change["ПометкаУдаления"]
					params["date"] = change["Дата"]
					params["number"] = change["Номер"]
					params["status"] = change["Реквизиты"].(map[string]interface{})["Статус"].(map[string]interface{})["Ссылка"]

					params["tables"] = tables.CopyTables(&tables.OutcomeOrders)

					lines := []interface{}{}

					lineCount := 0
					for _, line := range change["ТабличныеЧасти"].(map[string]interface{})["ТоварыПоРаспоряжениям"].([]interface{}) {

						addLine := map[string]interface{}{

							"id":                uuid.New().String(),
							"doc_id":            id,
							"line_number":       lineCount,
							"product_id":        line.(map[string]interface{})["Номенклатура"].(map[string]interface{})["Ссылка"],
							"characteristic_id": line.(map[string]interface{})["Характеристика"].(map[string]interface{})["Ссылка"],
							"order_id":          line.(map[string]interface{})["Распоряжение"].(map[string]interface{})["Ссылка"],
							"quantity":          line.(map[string]interface{})["Количество"],
						}

						lines = append(lines, addLine)

						lineCount += 1

					}

					params["tables"].(map[string]interface{})["products_by_orders"].(map[string]interface{})["lines"] = lines

					err = UpdateObject(&tables.OutcomeOrders, params)

					return ChangeResponse{"Документ", "РасходныйОрдерНаТовары", id}, err

				}

			}

		}
	}
	return ChangeResponse{}, nil
}

func UpdateWarehouseLeftovers(cell_id, container_id, product_id, characteristic_id string, quantity int, unit_id string, unitQuantity int) error {

	table := &tables.Leftovers

	tx, err := Db.Begin()
	if err != nil {
		return err
	}

	defer tx.Commit()

	statement, err := tx.Prepare(tables.GetInsertQuery(table))

	if err != nil {
		return err
	}

	_, err = statement.Exec(uuid.New().String(), cell_id, container_id, product_id, characteristic_id, quantity, unit_id, unitQuantity)

	if err != nil {
		return err
	}

	return nil

}

func DeleteTable(table *tables.Table, filters []tables.Filter) error {

	tx, err := Db.Begin()
	if err != nil {
		return err
	}

	defer tx.Commit()

	statement, _ := tx.Prepare(tables.GetDeleteQuery(table, filters))

	_, err = statement.Exec()

	if err != nil {
		return err
	}

	return nil

}

func GetWarehouseLeftovers() {

	client := &http.Client{}
	req, _ := http.NewRequest(
		"GET", "https://ow.apx-service.ru/tm_po/hs/dta/obj?request=getErpSkladWhdctLeftovers&warehouse_id=2e1e0023-b9b1-11e0-9a78-00155d7b7606", nil,
	)
	// добавляем заголовки
	// req.Header.Add("Accept", "text/html")     // добавляем заголовок Accept
	// req.Header.Add("User-Agent", "MSIE/15.0") // добавляем заголовок User-Agent

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	dec.DisallowUnknownFields()

	var p ErpResponse
	err = dec.Decode(&p)
	if err != nil {

		fmt.Println(err)
		return

	}

	ErpSkladWhdctLeftovers := p.Responses[0].(map[string]interface{})["ErpSkladWhdctLeftovers"].(map[string]interface{})

	Result := ErpSkladWhdctLeftovers["result"].([]interface{})

	DeleteTable(&tables.Leftovers, []tables.Filter{})

	for i := 0; i < len(Result); i++ {

		change := Result[i].(map[string]interface{})

		if UpdateWarehouseLeftovers(
			change["cell_id"].(string),
			change["container_id"].(string),
			change["product_id"].(string),
			change["characteristic_id"].(string),
			int(change["quantity"].(float64)),
			change["unit_id"].(string),
			int(change["unitQuantity"].(float64))) == nil {

		} else {

			fmt.Printf("begin. Exec error=%s", err)

		}

	}

}
*/