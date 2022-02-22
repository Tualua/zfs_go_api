package main

import (
	"encoding/json"
	"net/http"
)

type jsonResponseGeneric struct {
	Action       string                 `json:"action"`
	Status       string                 `json:"status"`
	ErrorMessage string                 `json:"errormessage,omitempty"`
	Data         map[string]interface{} `json:"data,omitempty"`
}
type jsonResponseListAll struct {
	jsonResponseGeneric
	ZfsEntities []ZfsEntity `json:"data"`
}

func (j *jsonResponseGeneric) SetAction(action string) {
	j.Action = action
}

func (j *jsonResponseGeneric) Success() {
	j.Status = "success"
}

func (j *jsonResponseGeneric) Error(message string) {
	j.Status = "error"
	j.ErrorMessage = message
}

func (j *jsonResponseGeneric) SetVal(key string, val interface{}) {
	if j.Data == nil {
		j.Data = make(map[string]interface{})
	}
	j.Data[key] = val
}

func (j *jsonResponseGeneric) Write(w *http.ResponseWriter) {
	enc := json.NewEncoder(*w)
	enc.SetIndent("", "    ")
	enc.Encode(j)
}

func (j *jsonResponseListAll) Write(w *http.ResponseWriter) {
	enc := json.NewEncoder(*w)
	enc.SetIndent("", "    ")
	enc.Encode(j)
}
