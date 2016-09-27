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
	"fmt"
	"io/ioutil"

	"github.com/labstack/echo"
)

// FailWith fails with the specified message
func FailWith(status int, message string, c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return c.String(status, fmt.Sprintf(`{"success":false,"reason":"%s"}`, message))
}

// SucceedWith sends payload to member with status 200
func SucceedWith(payload map[string]interface{}, c echo.Context) error {
	payload["success"] = true
	return c.JSON(200, payload)
}

//LoadJSONPayload loads the JSON payload to the given struct validating all fields are not null
//func LoadJSONPayload(payloadStruct interface{}, c echo.Context, l zap.Logger) error {
//log.D(l, "Loading payload...")

//data, err := GetRequestBody(c)
//if err != nil {
//log.E(l, "Loading payload failed.", func(cm log.CM) {
//cm.Write(zap.Error(err))
//})
//return err
//}

//unmarshaler, ok := payloadStruct.(EasyJSONUnmarshaler)
//if !ok {
//err := fmt.Errorf("Can't unmarshal specified payload since it does not implement easyjson interface")
//log.E(l, "Loading payload failed.", func(cm log.CM) {
//cm.Write(zap.Error(err))
//})
//return err
//}

//lexer := jlexer.Lexer{Data: []byte(data)}
//unmarshaler.UnmarshalEasyJSON(&lexer)
//if err = lexer.Error(); err != nil {
//log.E(l, "Loading payload failed.", func(cm log.CM) {
//cm.Write(zap.Error(err))
//})
//return err
//}

//if validatable, ok := payloadStruct.(Validatable); ok {
//missingFieldErrors := validatable.Validate()

//if len(missingFieldErrors) != 0 {
//err := errors.New(strings.Join(missingFieldErrors[:], ", "))
//log.E(l, "Loading payload failed.", func(cm log.CM) {
//cm.Write(zap.Error(err))
//})
//return err
//}
//}

//log.D(l, "Payload loaded successfully.")
//return nil
//}

//GetRequestBody from echo context
func GetRequestBody(c echo.Context) ([]byte, error) {
	bodyCache := c.Get("requestBody")
	if bodyCache != nil {
		return bodyCache.([]byte), nil
	}
	body := c.Request().Body()
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	c.Set("requestBody", b)
	return b, nil
}

//GetRequestJSON as the specified interface from echo context
func GetRequestJSON(payloadStruct interface{}, c echo.Context) error {
	body, err := GetRequestBody(c)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, payloadStruct)
	if err != nil {
		return err
	}

	return nil
}
