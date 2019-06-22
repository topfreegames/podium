// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package api

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjsonA8a797f8DecodeGithubComTopfreegamesPodiumApi(in *jlexer.Lexer, out *setScoresPayload) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "score":
			out.Score = int64(in.Int64())
		case "leaderboards":
			if in.IsNull() {
				in.Skip()
				out.Leaderboards = nil
			} else {
				in.Delim('[')
				if out.Leaderboards == nil {
					if !in.IsDelim(']') {
						out.Leaderboards = make([]string, 0, 4)
					} else {
						out.Leaderboards = []string{}
					}
				} else {
					out.Leaderboards = (out.Leaderboards)[:0]
				}
				for !in.IsDelim(']') {
					var v1 string
					v1 = string(in.String())
					out.Leaderboards = append(out.Leaderboards, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonA8a797f8EncodeGithubComTopfreegamesPodiumApi(out *jwriter.Writer, in setScoresPayload) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"score\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int64(int64(in.Score))
	}
	{
		const prefix string = ",\"leaderboards\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.Leaderboards == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v2, v3 := range in.Leaderboards {
				if v2 > 0 {
					out.RawByte(',')
				}
				out.String(string(v3))
			}
			out.RawByte(']')
		}
	}
	out.RawByte('}')
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v setScoresPayload) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonA8a797f8EncodeGithubComTopfreegamesPodiumApi(w, v)
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *setScoresPayload) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonA8a797f8DecodeGithubComTopfreegamesPodiumApi(l, v)
}
func easyjsonA8a797f8DecodeGithubComTopfreegamesPodiumApi1(in *jlexer.Lexer, out *setScorePayload) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "score":
			out.Score = int64(in.Int64())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonA8a797f8EncodeGithubComTopfreegamesPodiumApi1(out *jwriter.Writer, in setScorePayload) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"score\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int64(int64(in.Score))
	}
	out.RawByte('}')
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v setScorePayload) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonA8a797f8EncodeGithubComTopfreegamesPodiumApi1(w, v)
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *setScorePayload) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonA8a797f8DecodeGithubComTopfreegamesPodiumApi1(l, v)
}
func easyjsonA8a797f8DecodeGithubComTopfreegamesPodiumApi2(in *jlexer.Lexer, out *setMembersScorePayload) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "members":
			if in.IsNull() {
				in.Skip()
				out.MembersScore = nil
			} else {
				in.Delim('[')
				if out.MembersScore == nil {
					if !in.IsDelim(']') {
						out.MembersScore = make([]*memberScorePayload, 0, 8)
					} else {
						out.MembersScore = []*memberScorePayload{}
					}
				} else {
					out.MembersScore = (out.MembersScore)[:0]
				}
				for !in.IsDelim(']') {
					var v4 *memberScorePayload
					if in.IsNull() {
						in.Skip()
						v4 = nil
					} else {
						if v4 == nil {
							v4 = new(memberScorePayload)
						}
						(*v4).UnmarshalEasyJSON(in)
					}
					out.MembersScore = append(out.MembersScore, v4)
					in.WantComma()
				}
				in.Delim(']')
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonA8a797f8EncodeGithubComTopfreegamesPodiumApi2(out *jwriter.Writer, in setMembersScorePayload) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"members\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.MembersScore == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v5, v6 := range in.MembersScore {
				if v5 > 0 {
					out.RawByte(',')
				}
				if v6 == nil {
					out.RawString("null")
				} else {
					(*v6).MarshalEasyJSON(out)
				}
			}
			out.RawByte(']')
		}
	}
	out.RawByte('}')
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v setMembersScorePayload) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonA8a797f8EncodeGithubComTopfreegamesPodiumApi2(w, v)
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *setMembersScorePayload) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonA8a797f8DecodeGithubComTopfreegamesPodiumApi2(l, v)
}
func easyjsonA8a797f8DecodeGithubComTopfreegamesPodiumApi3(in *jlexer.Lexer, out *memberScorePayload) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "score":
			out.Score = int64(in.Int64())
		case "publicID":
			out.PublicID = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonA8a797f8EncodeGithubComTopfreegamesPodiumApi3(out *jwriter.Writer, in memberScorePayload) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"score\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int64(int64(in.Score))
	}
	{
		const prefix string = ",\"publicID\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.PublicID))
	}
	out.RawByte('}')
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v memberScorePayload) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonA8a797f8EncodeGithubComTopfreegamesPodiumApi3(w, v)
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *memberScorePayload) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonA8a797f8DecodeGithubComTopfreegamesPodiumApi3(l, v)
}
func easyjsonA8a797f8DecodeGithubComTopfreegamesPodiumApi4(in *jlexer.Lexer, out *incrementScorePayload) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "increment":
			out.Increment = int(in.Int())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonA8a797f8EncodeGithubComTopfreegamesPodiumApi4(out *jwriter.Writer, in incrementScorePayload) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"increment\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int(int(in.Increment))
	}
	out.RawByte('}')
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v incrementScorePayload) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonA8a797f8EncodeGithubComTopfreegamesPodiumApi4(w, v)
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *incrementScorePayload) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonA8a797f8DecodeGithubComTopfreegamesPodiumApi4(l, v)
}
func easyjsonA8a797f8DecodeGithubComTopfreegamesPodiumApi5(in *jlexer.Lexer, out *Validation) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonA8a797f8EncodeGithubComTopfreegamesPodiumApi5(out *jwriter.Writer, in Validation) {
	out.RawByte('{')
	first := true
	_ = first
	out.RawByte('}')
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Validation) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonA8a797f8EncodeGithubComTopfreegamesPodiumApi5(w, v)
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Validation) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonA8a797f8DecodeGithubComTopfreegamesPodiumApi5(l, v)
}
