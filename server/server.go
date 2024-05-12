package server

import (
	tables "efr_pack/db"
	"efr_pack/server/exchange"
	"efr_pack/server/httphandlers"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

func StrToMap(in, delimeter string) map[string]interface{} {
	res := make(map[string]interface{})

	if in != "" {

		array := strings.Split(in, delimeter)
		temp := make([]string, 2)

		for _, val := range array {
			temp = strings.Split(string(val), "=")

			if len(temp) > 1 {
				res[temp[0]] = temp[1]

			}

		}
	}

	return res
}

func Start() error {

	var err error = nil

	http.HandleFunc("/", func(response http.ResponseWriter, request *http.Request) {

		res := ""

		switch request.RequestURI {

		case "":
			res = "/index.html"
		default:
			res = request.RequestURI

		}

		http.ServeFile(response, request, "./dist"+res)

	})

	http.HandleFunc("/menuplans/", func(response http.ResponseWriter, request *http.Request) {

		res := ""

		switch strings.ReplaceAll(request.RequestURI, "/menuplans/", "") {

		case "":
			res = "/index.html"
		default:
			res = request.RequestURI

		}

		http.ServeFile(response, request, "./dist"+res)

	})

	http.HandleFunc("/assets/", func(response http.ResponseWriter, request *http.Request) {

		http.ServeFile(response, request, "./dist"+request.RequestURI)

	})

	http.HandleFunc("/srv/gettable", func(response http.ResponseWriter, request *http.Request) {

		response.Header().Set("Access-Control-Allow-Origin", "*")
		response.Header().Set("Access-Control-Allow-Credentials", "true")
		response.Header().Set("Access-Control-Allow-Methods", "GET,HEAD,OPTIONS,POST,PUT")
		response.Header().Set("Access-Control-Allow-Headers", "Access-Control-Allow-Headers, Origin,Accept, X-Requested-With, Content-Type, Access-Control-Request-Method, Access-Control-Request-Headers")

		var tm map[string]interface{}

		if request.Method == "GET" {

			ue, _ := url.QueryUnescape(request.URL.RawQuery)

			tm = StrToMap(ue, "&")

			filter := []interface{}{}

			filterStr := ""
			_, ok := tm["filter"]
			if ok {
				filterStr = tm["filter"].(string)
			}

			if filterStr != "" {

				kv := strings.Split(filterStr, " eq ")

				filter = append(filter, map[string]interface{}{
					"text": kv[0] + " = ?",
					"parameter": map[string]interface{}{
						"name":  kv[0],
						"value": kv[1],
					},
				})

			}

			tm["filter"] = filter

		} else if request.Method == "POST" {

			err := json.NewDecoder(request.Body).Decode(&tm)
			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}

		}

		var js []byte = make([]byte, 0)

		_, ok := tm["name"]
		if ok {

			rows, err := tables.GetTableData(tm)

			if err != nil {

				js, _ = json.Marshal(map[string]interface{}{
					"success": false,
					"message": err,
				})

			} else {

				js, _ = json.Marshal(map[string]interface{}{
					"success": true,
					"result":  rows,
				})
			}
		} else {

			js, err = json.Marshal(map[string]interface{}{
				"success": false,
				"message": "name required",
			})

		}
		response.Header().Set("Content-Type", "application/json")
		response.Write(js)

	})

	http.HandleFunc("/srv/startexch", func(response http.ResponseWriter, request *http.Request) {

		httphandlers.HandleFunc(response, request, exchange.StartExchange)

	})

	http.HandleFunc("/srv/insertrecord", func(response http.ResponseWriter, request *http.Request) {

		httphandlers.HandleFunc(response, request, tables.InsertRecord)

	})

	http.HandleFunc("/srv/insertrecords", func(response http.ResponseWriter, request *http.Request) {

		httphandlers.HandleFunc(response, request, tables.InsertRecords)

	})

	http.ListenAndServe(":8001", nil)

	return err

}
