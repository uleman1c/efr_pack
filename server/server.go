package server

import (
	tables "efr_pack/db"
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
			res[temp[0]] = temp[1]
		}
	}

	return res
}

func Start() error {

	var err error = nil

	http.HandleFunc("/", func(response http.ResponseWriter, request *http.Request) {

		http.ServeFile(response, request, "./dist/index.html")

	})

	http.HandleFunc("/assets/", func(response http.ResponseWriter, request *http.Request) {

		http.ServeFile(response, request, "./dist"+request.RequestURI)

	})

	http.HandleFunc("/srv/gettable", func(response http.ResponseWriter, request *http.Request) {

		if request.Method == "GET" {

			ue, _ := url.QueryUnescape(request.URL.RawQuery)

			tm := StrToMap(ue, "&")

			rows, err := tables.GetTableData(tm)

			var js []byte = make([]byte, 0)

			if err != nil {

				js, _ = json.Marshal(err)

			} else {

				js, _ = json.Marshal(rows)
			}

			response.Header().Set("Content-Type", "application/json")
			response.Write(js)

		}

	})

	/* 	http.HandleFunc("/user", func(response http.ResponseWriter, request *http.Request) {

	   		if request.Method == "GET" {

	   			js, _ := json.Marshal(getUser())

	   			response.Header().Set("Content-Type", "application/json")
	   			response.Write(js)

	   		}

	   	})

	   	http.HandleFunc("/constants", func(response http.ResponseWriter, request *http.Request) {

	   		if request.Method == "GET" {

	   			ue, _ := url.QueryUnescape(request.URL.RawQuery)

	   			tm := StrToMap(ue, "&")

	   			fmt.Println(tm)

	   			rows := database.GetTable(&tables.Constants)

	   			js, _ := json.Marshal(rows)

	   			response.Header().Set("Content-Type", "application/json")
	   			response.Write(js)

	   		}

	   	})

	   	http.HandleFunc("/alltables", func(response http.ResponseWriter, request *http.Request) {

	   		if request.Method == "GET" {

	   			ue, _ := url.QueryUnescape(request.URL.RawQuery)

	   			tm := StrToMap(ue, "&")

	   			fmt.Println(tm)

	   			rows, _ := database.GetQueryResult("SELECT name FROM sqlite_master WHERE type='table'", []string{"name"}, []interface{}{})

	   			js, _ := json.Marshal(rows)

	   			response.Header().Set("Content-Type", "application/json")
	   			response.Write(js)

	   		}

	   	})

	   	http.HandleFunc("/whdct", func(response http.ResponseWriter, request *http.Request) {

	   		if request.Method == "GET" {

	   			ue, _ := url.QueryUnescape(request.URL.RawQuery)

	   			tm := StrToMap(ue, "&")

	   			result := map[string]interface{}{}
	   			result["success"] = true
	   			result["responses"] = make([]map[string]interface{}, 1)

	   			result["responses"].([]map[string]interface{})[0] = responseToRequest(tm)

	   			js, _ := json.Marshal(result)

	   			response.Header().Set("Content-Type", "application/json")
	   			response.Write(js)

	   		}
	   	})
	*/
	http.ListenAndServe(":8001", nil)

	return err

}
