package myhttp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Response interface {
	StatusHTTP() int
}

func WriteResponseTo(rw http.ResponseWriter, r Response) error {
	json, err := json.Marshal(r)
	if err != nil {
		return err
	}
	rw.WriteHeader(r.StatusHTTP())
	rw.Write(json)
	return nil
}

func UnmarshalRequestBody(req *http.Request, v interface{}) error {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return fmt.Errorf("myhttp: %v", err)
	}
	return json.Unmarshal(body, v)
}
