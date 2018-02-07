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
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo"
	"github.com/mailru/easyjson/jlexer"
	"github.com/mailru/easyjson/jwriter"
	newrelic "github.com/newrelic/go-agent"
	"github.com/topfreegames/podium/log"
	"github.com/topfreegames/podium/util"
	"go.uber.org/zap"
)

//EasyJSONUnmarshaler describes a struct able to unmarshal json
type EasyJSONUnmarshaler interface {
	UnmarshalEasyJSON(l *jlexer.Lexer)
}

//EasyJSONMarshaler describes a struct able to marshal json
type EasyJSONMarshaler interface {
	MarshalEasyJSON(w *jwriter.Writer)
}

// FailWith fails with the specified message
func FailWith(status int, message string, c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return c.String(status, fmt.Sprintf(`{"success":false,"reason":"%s"}`, message))
}

// FailWithError fails with the specified error
func FailWithError(e error, c echo.Context) error {
	var statusCode int
	switch e.(type) {
	default:
		statusCode = http.StatusInternalServerError
	case *util.LeaderboardExpiredError:
		statusCode = http.StatusBadRequest
	}
	return FailWith(statusCode, e.Error(), c)
}

// SucceedWith sends payload to member with status 200
func SucceedWith(payload map[string]interface{}, c echo.Context) error {
	if len(payload) == 0 {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		return c.String(200, `{"success":true}`)
	}
	payload["success"] = true
	return c.JSON(200, payload)
}

//LoadJSONPayload loads the JSON payload to the given struct validating all fields are not null
func LoadJSONPayload(payloadStruct interface{}, c echo.Context, l zap.Logger) error {
	log.D(l, "Loading payload...")

	data, err := GetRequestBody(c)
	if err != nil {
		log.E(l, "Loading payload failed.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return err
	}

	unmarshaler, ok := payloadStruct.(EasyJSONUnmarshaler)
	if !ok {
		err := fmt.Errorf("Can't unmarshal specified payload since it does not implement easyjson interface")
		log.E(l, "Loading payload failed.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return err
	}

	lexer := jlexer.Lexer{Data: []byte(data)}
	unmarshaler.UnmarshalEasyJSON(&lexer)
	if err = lexer.Error(); err != nil {
		log.E(l, "Loading payload failed.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return err
	}

	if validatable, ok := payloadStruct.(Validatable); ok {
		missingFieldErrors := validatable.Validate()

		if len(missingFieldErrors) != 0 {
			err := errors.New(strings.Join(missingFieldErrors[:], ", "))
			log.E(l, "Loading payload failed.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return err
		}
	}

	log.D(l, "Payload loaded successfully.")
	return nil
}

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

//GetIntRouteParam gets a query parameter as integer
func GetIntRouteParam(app *App, c echo.Context, paramName string, defaultValue int) (int, error) {
	var err error
	paramStr := c.Param(paramName)
	param := int64(defaultValue)

	if paramStr != "" {
		param, err = strconv.ParseInt(paramStr, 10, 32)
		if err != nil {
			app.AddError()
			return 0, fmt.Errorf("Failed to process route param %s: %s", paramName, paramStr)
		}
	}

	return int(param), nil
}

//GetIntParam gets a query parameter as integer
func GetIntParam(app *App, c echo.Context, paramName string, defaultValue int) (int, error) {
	var err error
	paramStr := c.QueryParam(paramName)
	param := int64(defaultValue)

	if paramStr != "" {
		param, err = strconv.ParseInt(paramStr, 10, 32)
		if err != nil {
			app.AddError()
			return 0, fmt.Errorf("Failed to process param %s: %s", paramName, paramStr)
		}
	}

	return int(param), nil
}

//GetPageSize gets the page size from the querystring
func GetPageSize(app *App, c echo.Context, defaultPageSize int) (int, error) {
	var err error
	pageSize, err := GetIntParam(app, c, "pageSize", defaultPageSize)
	if err != nil {
		return 0, err
	}

	if pageSize > app.Config.GetInt("api.maxReturnedMembers") {
		msg := fmt.Sprintf(
			"Max pageSize allowed: %d. pageSize requested: %d",
			app.Config.GetInt("api.maxReturnedMembers"),
			pageSize,
		)
		return 0, fmt.Errorf(msg)
	}

	return pageSize, nil
}

//GetTX returns new relic transaction
func GetTX(c echo.Context) newrelic.Transaction {
	tx := c.Get("txn")
	if tx == nil {
		return nil
	}

	return tx.(newrelic.Transaction)
}

//WithSegment adds a segment to new relic transaction
func WithSegment(name string, c echo.Context, f func() error) error {
	tx := GetTX(c)
	if tx == nil {
		return f()
	}
	segment := newrelic.StartSegment(tx, name)
	defer segment.End()
	return f()
}
