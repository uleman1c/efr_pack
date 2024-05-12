package httphandlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

func addCorsHeader(res http.ResponseWriter) {
	headers := res.Header()
	headers.Add("Access-Control-Allow-Origin", "*")
	headers.Add("Vary", "Origin")
	headers.Add("Vary", "Access-Control-Request-Method")
	headers.Add("Vary", "Access-Control-Request-Headers")
	headers.Add("Access-Control-Allow-Headers", "Content-Type, Origin, Accept, token")
	headers.Add("Access-Control-Allow-Methods", "GET, POST,OPTIONS")
}

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

func HandleFunc(response http.ResponseWriter, request *http.Request, f func(in map[string]interface{}) (map[string]interface{}, error)) {

	addCorsHeader(response)

	var tm map[string]interface{}

	if request.Method == "GET" {

		ue, _ := url.QueryUnescape(request.URL.RawQuery)

		tm = StrToMap(ue, "&")

	} else if request.Method == "POST" {

		err := json.NewDecoder(request.Body).Decode(&tm)
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}

	} else if request.Method == "OPTIONS" {

		response.WriteHeader(http.StatusOK)
		return
	}

	res, err := f(tm)

	var js []byte = make([]byte, 0)

	js, err = json.Marshal(res)
	if err != nil {
		http.Error(response, err.Error(), http.StatusBadRequest)
		return
	}

	response.Header().Set("Content-Type", "application/json")
	response.Write(js)

}
