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

	}
	return result, err
}

func existsById(table map[string]interface{}, id string) (bool, error) {

	tx, err := Db.Begin()
	if err != nil {

		return false, err

	}

	defer tx.Commit()

	constFilter := []map[string]interface{}{
		{
			"text":      "id = ?",
			"parameter": id,
		},
	}

	statement, err := tx.Prepare(GetSelectQuery(table, constFilter))

	if err != nil {
		return false, err
	}

	rows, err := statement.Query(id)

	if err != nil {
		return false, err
	}

	exist := rows.Next()

	defer rows.Close()

	return exist, err

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
				"params": map[string]interface{}{
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
