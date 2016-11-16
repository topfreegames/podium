// podium
// https://github.com/topfreegames/podium
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>
// Forked from
// https://github.com/dayvson/go-leaderboard
// Copyright © 2013 Maxwell Dayvson da Silva

//go:generate easyjson -all -no_std_marshalers $GOFILE

package api

import "fmt"

//Validatable indicates that a struct can be validated
type Validatable interface {
	Validate() []string
}

//ValidatePayload for any validatable payload
func ValidatePayload(payload Validatable) []string {
	return payload.Validate()
}

//NewValidation for validating structs
func NewValidation() *Validation {
	return &Validation{
		errors: []string{},
	}
}

//Validation struct
type Validation struct {
	errors []string
}

func (v *Validation) validateRequired(name string, value interface{}) {
	if value == nil {
		v.errors = append(v.errors, fmt.Sprintf("%s is required", name))
	}
}

func (v *Validation) validateRequiredInt(name string, value int) {
	if value == 0 {
		v.errors = append(v.errors, fmt.Sprintf("%s is required", name))
	}
}

func (v *Validation) validateCustom(name string, valFunc func() []string) {
	errors := valFunc()
	if len(errors) > 0 {
		v.errors = append(v.errors, errors...)
	}
}

//Errors in validation
func (v *Validation) Errors() []string {
	return v.errors
}

type incrementScorePayload struct {
	Increment int `json:"increment"`
}

func (s *incrementScorePayload) Validate() []string {
	v := NewValidation()
	v.validateRequiredInt("increment", s.Increment)
	return v.Errors()
}

type setScorePayload struct {
	Score int `json:"score"`
}

func (s *setScorePayload) Validate() []string {
	v := NewValidation()
	//v.validateRequiredInt("score", s.Score)
	return v.Errors()
}

type setScoresPayload struct {
	Score        int      `json:"score"`
	Leaderboards []string `json:"leaderboards"`
}

func (s *setScoresPayload) Validate() []string {
	v := NewValidation()
	//v.validateRequiredInt("score", s.Score)

	v.validateCustom("leaderboards", func() []string {
		if len(s.Leaderboards) == 0 {
			return []string{"leaderboards is required"}
		}
		return []string{}
	})

	return v.Errors()
}
