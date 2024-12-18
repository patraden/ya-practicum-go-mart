// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package model

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

func easyjsonB5e4bf8DecodeGithubComPatradenYaPracticumGoMartInternalAppDomainModel(in *jlexer.Lexer, out *UserBalance) {
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
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "current":
			if data := in.Raw(); in.Ok() {
				in.AddError((out.Balance).UnmarshalJSON(data))
			}
		case "withdrawn":
			if data := in.Raw(); in.Ok() {
				in.AddError((out.Withdrawn).UnmarshalJSON(data))
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
func easyjsonB5e4bf8EncodeGithubComPatradenYaPracticumGoMartInternalAppDomainModel(out *jwriter.Writer, in UserBalance) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"current\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Raw((in.Balance).MarshalJSON())
	}
	{
		const prefix string = ",\"withdrawn\":"
		out.RawString(prefix)
		out.Raw((in.Withdrawn).MarshalJSON())
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v UserBalance) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonB5e4bf8EncodeGithubComPatradenYaPracticumGoMartInternalAppDomainModel(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v UserBalance) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonB5e4bf8EncodeGithubComPatradenYaPracticumGoMartInternalAppDomainModel(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *UserBalance) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonB5e4bf8DecodeGithubComPatradenYaPracticumGoMartInternalAppDomainModel(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *UserBalance) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonB5e4bf8DecodeGithubComPatradenYaPracticumGoMartInternalAppDomainModel(l, v)
}