//  podium
// https://github.com/topfreegames/podium
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>
// Forked from
// https://github.com/dayvson/go-leaderboard
// Copyright © 2013 Maxwell Dayvson da Silva

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/kataras/iris"
)

// FailWith fails with the specified message
func FailWith(status int, message string, c *iris.Context) {
	result, _ := json.Marshal(map[string]interface{}{
		"success": false,
		"reason":  message,
	})
	c.SetStatusCode(status)
	c.Write(string(result))
}

// SucceedWith sends payload to user with status 200
func SucceedWith(payload map[string]interface{}, c *iris.Context) {
	payload["success"] = true
	result, _ := json.Marshal(payload)
	c.SetStatusCode(200)
	c.Write(string(result))
}

// LoadJSONPayload loads the JSON payload to the given struct validating all fields are not null
func LoadJSONPayload(payloadStruct interface{}, c *iris.Context) error {
	if err := c.ReadJSON(payloadStruct); err != nil {
		if err != nil {
			return err
		}
	}

	data := c.RequestCtx.Request.Body()
	var jsonPayload map[string]interface{}
	err := json.Unmarshal(data, &jsonPayload)
	if err != nil {
		return err
	}

	var missingFieldErrors []string
	v := reflect.ValueOf(payloadStruct).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		r, n := utf8.DecodeRuneInString(t.Field(i).Name)
		field := string(unicode.ToLower(r)) + t.Field(i).Name[n:]
		if jsonPayload[field] == nil {
			missingFieldErrors = append(missingFieldErrors, fmt.Sprintf("%s is required", field))
		}
	}

	if len(missingFieldErrors) != 0 {
		error := errors.New(strings.Join(missingFieldErrors[:], ", "))
		return error
	}

	return nil
}
