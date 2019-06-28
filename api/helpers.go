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
	"context"
	"fmt"

	"github.com/mailru/easyjson/jlexer"
	"github.com/mailru/easyjson/jwriter"
	newrelic "github.com/newrelic/go-agent"
)

//EasyJSONUnmarshaler describes a struct able to unmarshal json
type EasyJSONUnmarshaler interface {
	UnmarshalEasyJSON(l *jlexer.Lexer)
}

//EasyJSONMarshaler describes a struct able to marshal json
type EasyJSONMarshaler interface {
	MarshalEasyJSON(w *jwriter.Writer)
}

func newFailMsg(msg string) string {
	return fmt.Sprintf(`{"success":false,"reason":"%s"}`, msg)
}

func withSegment(name string, ctx context.Context, f func() error) error {
	if txn := ctx.Value(newRelicContextKey{"txn"}); txn != nil {
		if txn := txn.(newrelic.Transaction); txn != nil {
			segment := newrelic.StartSegment(txn, name)
			defer segment.End()
		}
	}
	return f()
}
