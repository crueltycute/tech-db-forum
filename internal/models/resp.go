package models

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func ErrResponse(res http.ResponseWriter, errCode int, errMsg string) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(errCode)
	//addErrBody(res, errMsg)

	marshalBody, err := json.Marshal(Error{Message: errMsg})
	if err != nil {
		fmt.Println(err)
		return
	}

	_, _ = res.Write(marshalBody)
}

func ResponseObject(res http.ResponseWriter, code int, body interface{ MarshalJSON() ([]byte, error) }) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(code)
	//addEasyJSONBody(res, body)

	binary, err := body.MarshalJSON()
	if err != nil {
		fmt.Println(err)
		return
	}

	_, _ = res.Write(binary)
}
