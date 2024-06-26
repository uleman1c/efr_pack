package tables

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
)

var uppUrl = "http://192.168.100.53:8080/efr_upp/" // "https://ow.apx-service.ru/tm_po/"
var user = "exch"
var pwd = "123456"

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func redirectPolicyFunc(req *http.Request, via []*http.Request) error {
	req.Header.Add("Authorization", "Basic "+basicAuth(user, pwd))
	return nil
}

func GetChangesFromCentral() error {

	UpdateConstant("exchange_in_progress", "1")

	var err error = nil

	//whid, _ := GetConstantValue("warehouse_id")

	changesExist := true

	for changesExist {

		client := &http.Client{CheckRedirect: redirectPolicyFunc}
		req, err := http.NewRequest(
			"GET", uppUrl+"hs/exch/req?request=getPackChanges", nil,
		)
		req.Header.Add("Authorization", "Basic "+basicAuth(user, pwd))

		if err != nil {
			return err
		}

		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		body = bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))

		var p map[string]interface{}
		err = json.Unmarshal(body, &p)
		if err != nil {
			return err
		}

		Changes := p["responses"].([]interface{})[0].(map[string]interface{})["PackChanges"].(map[string]interface{})

		Result := Changes["result"].([]interface{})

		if len(Result) == 0 {

			changesExist = false

		} else {

			changesResponse := []interface{}{}

			for i := 0; i < len(Result); i++ {

				change := Result[i].(map[string]interface{})

				changeResponse, err := loadObject(change)

				if err == nil {

					changesResponse = append(changesResponse, changeResponse)

				}

			}

			resMap := map[string]interface{}{}
			resMap["request"] = "setPackChanges"
			resMap["parameters"] = map[string]interface{}{
				//"warehouse_id": *whid,
				"changes": changesResponse,
			}

			jsonResponse, _ := json.Marshal(resMap)

			req, _ = http.NewRequest(
				"POST", uppUrl+"hs/exch/req",
				bytes.NewBuffer(jsonResponse),
			)
			req.Header.Set("Content-Type", "application/json; charset=UTF-8")
			req.Header.Add("Authorization", "Basic "+basicAuth(user, pwd))

			_, err = client.Do(req)
			if err != nil {
				return err
			}
		}

	}

	UpdateConstant("exchange_in_progress", "0")

	return err

}

func loadObject(change map[string]interface{}) (map[string]interface{}, error) {

	result := map[string]interface{}{"type": "", "name": "", "id": ""}

	var err error = nil

	switch change["Тип"] {

	case "Справочник":
		{

			result["type"] = "Справочник"

			switch change["Вид"] {

			case "Номенклатура":
				{

					result["name"] = "Номенклатура"

					id := change["Ссылка"].(string)

					result["id"] = id

					reqvisits := change["Реквизиты"].(map[string]interface{})

					params := map[string]interface{}{}
					params["id"] = id
					params["name"] = change["Наименование"]
					params["full_name"] = reqvisits["НаименованиеПолное"]

					err = UpdateObject(Tables["Products"], params)

				}

			case "Контейнеры":
				{

					result["name"] = "Контейнеры"

					id := change["Ссылка"].(string)

					result["id"] = id

					params := map[string]interface{}{}
					params["id"] = id
					params["name"] = change["Наименование"]

					err = UpdateObject(Tables["Containers"], params)

				}

			}

		}

	case "Документ":
		{

			result["type"] = "Документ"

			switch change["Вид"] {

			case "ПланМеню":
				{

					result["name"] = "ПланМеню"

					id := change["Ссылка"].(string)

					result["id"] = id

					params := map[string]interface{}{}
					params["id"] = id
					params["deletion_mark"] = change["ПометкаУдаления"]
					params["date"] = change["Дата"]
					params["number"] = change["Номер"]

					params["tables"] = CopyTables(Tables["MenuPlans"])

					lines := []interface{}{}

					lineCount := 0
					for _, line := range change["ТабличныеЧасти"].(map[string]interface{})["Товары"].([]interface{}) {

						addLine := map[string]interface{}{

							"id":          uuid.New().String(),
							"doc_id":      id,
							"line_number": lineCount,
							"product_id":  line.(map[string]interface{})["Номенклатура"].(map[string]interface{})["Ссылка"],
							"quantity":    line.(map[string]interface{})["Количество"],
						}

						lines = append(lines, addLine)

						lineCount += 1

					}

					params["tables"].(map[string]interface{})["products"].(map[string]interface{})["lines"] = lines

					err = UpdateObject(Tables["MenuPlans"], params)

				}

			}

		}

	case "РегистрСведений":
		{
			result["type"] = "РегистрСведений"

			switch change["Вид"] {

			case "Штрихкоды":
				{
					result["name"] = "Штрихкоды"

					reqvisits := change["Отбор"].(map[string]interface{})

					result["id"] = reqvisits

					params := map[string]interface{}{}
					params["barcode"] = reqvisits["Штрихкод"]
					params["product_id"] = reqvisits["Владелец"].(map[string]interface{})["Ссылка"].(string)

					filters := []map[string]interface{}{
						{
							"text": "barcode = ?",
							"parameter": map[string]interface{}{
								"name":  "barcode",
								"value": params["barcode"],
							},
						},
						{
							"text": "product_id = ?",
							"parameter": map[string]interface{}{
								"name":  "product_id",
								"value": params["product_id"],
							},
						},
					}

					err = ExecQuery(GetDeleteQuery(Tables["Barcodes"]["name"].(string), filters), filters)

					if err != nil {
						return result, err
					}

					for _, rec := range change["Записи"].([]interface{}) {

						params["id"] = uuid.New().String()
						params["barcode"] = rec.(map[string]interface{})["Штрихкод"]
						params["product_id"] = rec.(map[string]interface{})["Владелец"].(map[string]interface{})["Ссылка"].(string)

						err = UpdateObject(Tables["Barcodes"], params)

					}

				}
			case "ОтсканированныеШтрихкоды":
				{

					/* 					var Scanned = map[string]interface{}{

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
					*/

					result["name"] = "ОтсканированныеШтрихкоды"

					reqvisits := change["Отбор"].(map[string]interface{})

					result["id"] = reqvisits

					params := map[string]interface{}{}
					params["id"] = uuid.New().String()
					params["date"] = reqvisits["Период"]
					params["milliseconds"] = reqvisits["Миллисекунд"]
					params["barcode"] = reqvisits["Штрихкод"]

					filters := []map[string]interface{}{
						{
							"text": "date = ?",
							"parameter": map[string]interface{}{
								"name":  "date",
								"value": params["date"],
							},
						},
						{
							"text": "milliseconds = ?",
							"parameter": map[string]interface{}{
								"name":  "milliseconds",
								"value": params["milliseconds"],
							},
						},
						{
							"text": "barcode = ?",
							"parameter": map[string]interface{}{
								"name":  "barcode",
								"value": params["barcode"],
							},
						},
					}

					err = ExecQuery(GetDeleteQuery(Tables["Scanned"]["name"].(string), filters), filters)

					if err != nil {
						return result, err
					}

					for _, rec := range change["Записи"].([]interface{}) {

						params["id"] = uuid.New().String()
						params["container_id"] = rec.(map[string]interface{})["Контейнер"].(map[string]interface{})["Ссылка"].(string)
						params["menu_plan_id"] = rec.(map[string]interface{})["ПланМеню"].(map[string]interface{})["Ссылка"].(string)

						err = UpdateObject(Tables["Scanned"], params)
					}
				}
			}
		}
	}
	return result, err
}

func existsById(table map[string]interface{}, id string) (bool, error) {

	constFilter := []map[string]interface{}{
		{
			"text": "id = ?",
			"parameter": map[string]interface{}{
				"value": id,
			},
		},
	}

	return existsByFilter(table, constFilter)

}

func existsByFilter(table map[string]interface{}, filter []map[string]interface{}) (bool, error) {

	tx, err := Db.Begin()
	if err != nil {

		return false, err

	}

	defer tx.Commit()

	statement, err := tx.Prepare(GetSelectQuery(table, filter))

	if err != nil {
		return false, err
	}

	cp := GetParamsValuesFromFilter(filter)

	rows, err := statement.Query(cp...)

	if err != nil {
		return false, err
	}

	exist := rows.Next()

	defer rows.Close()

	return exist, err

}

func InsertInTable(table, params map[string]interface{}) error {

	tx, err := Db.Begin()
	if err != nil {
		return err
	}

	defer tx.Commit()

	statement, err := tx.Prepare(GetInsertQuery(table))

	if err != nil {
		return err
	}

	sqlParams := []interface{}{}

	switch table["fields"].(type) {

	case []interface{}:

		for _, field := range table["fields"].([]interface{}) {

			sqlParams = append(sqlParams, params[field.(map[string]interface{})["name"].(string)])

		}

	case []map[string]interface{}:

		for _, tableField := range table["fields"].([]map[string]interface{}) {

			sqlParams = append(sqlParams, params[tableField["name"].(string)])

		}
	}
	_, err = statement.Exec(sqlParams...)

	if err != nil {
		return err
	}

	return nil

}

func UpdateInTable(table map[string]interface{}) error {

	tx, err := Db.Begin()
	if err != nil {
		return err
	}

	defer tx.Commit()

	statement, err := tx.Prepare(GetUpdateQuery(table, table["filter"].([]map[string]interface{})))

	if err != nil {
		return err
	}

	sqlParams := []interface{}{}
	for _, tableField := range table["record"].(map[string]interface{}) {

		sqlParams = append(sqlParams, tableField)

	}

	for _, tableField := range table["filter"].([]map[string]interface{}) {

		sqlParams = append(sqlParams, tableField["parameter"].(map[string]interface{})["value"])

	}
	_, err = statement.Exec(sqlParams...)

	if err != nil {
		return err
	}

	return nil

}

func UpdateConstant(name, value string) error {

	inpTable := map[string]interface{}{
		"name": "Constants",
		"filter": map[string]interface{}{
			"name": name,
		},
	}

	rows, err := GetTableData(inpTable)

	if err != nil {

		return err

	}

	if len(rows) > 0 {

		curId := rows[0]["id"].(string)

		if curId != value {

			inpTable["filter"] = []map[string]interface{}{
				{
					"text": "id = ?",
					"parameter": map[string]interface{}{
						"value": curId,
					},
				},
			}

			inpTable["record"] = map[string]interface{}{
				"value": value,
			}

			return UpdateInTable(inpTable)

		} else {

			return nil

		}

	} else {

		params := map[string]interface{}{
			"id":    uuid.New().String(),
			"name":  name,
			"value": value,
		}

		return InsertInTable(Constants, params)

	}

}

func UpdateObject(table map[string]interface{}, params map[string]interface{}) error {

	exists, err := existsById(table, params["id"].(string))

	if err != nil {

		return err

	}

	if !exists {

		tx, err := Db.Begin()
		if err != nil {
			return err
		}

		statement, err := tx.Prepare(GetInsertQuery(table))

		if err != nil {
			return err
		}

		sqlParams := []interface{}{}
		for _, tableField := range table["fields"].([]map[string]interface{}) {

			sqlParams = append(sqlParams, params[tableField["name"].(string)])

		}

		_, err = statement.Exec(sqlParams...)

		if err != nil {
			return err
		}

		tx.Commit()

	}

	if _, ok := params["tables"]; ok {

		if err := UpdateObjectTables(table["name"].(string), params); err != nil {

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

		filters := []map[string]interface{}{
			{
				"text": "doc_id = ?",
				"parameter": map[string]interface{}{
					"name":  "doc_id",
					"value": params["id"],
				},
			},
		}

		q := GetDeleteQuery(mainTableName+"_"+tableName, filters)

		statement, err := tx.Prepare(q)

		if err != nil {
			return err
		}

		_, err = statement.Exec(params["id"])

		if err != nil {
			return err
		}

		for _, line := range tableData.(map[string]interface{})["lines"].([]interface{}) {

			q, sqlParams := GetInsertQuerySub(mainTableName+"_"+tableName, line.(map[string]interface{}))

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
