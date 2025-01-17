// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package dto

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

func easyjson9043616eDecodeGithubComPatradenYaPracticumGoMartInternalAppDto(in *jlexer.Lexer, out *UserCredentials) {
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
		case "login":
			out.Username = string(in.String())
		case "password":
			out.Password = string(in.String())
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
func easyjson9043616eEncodeGithubComPatradenYaPracticumGoMartInternalAppDto(out *jwriter.Writer, in UserCredentials) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"login\":"
		out.RawString(prefix[1:])
		out.String(string(in.Username))
	}
	{
		const prefix string = ",\"password\":"
		out.RawString(prefix)
		out.String(string(in.Password))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v UserCredentials) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson9043616eEncodeGithubComPatradenYaPracticumGoMartInternalAppDto(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v UserCredentials) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson9043616eEncodeGithubComPatradenYaPracticumGoMartInternalAppDto(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *UserCredentials) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson9043616eDecodeGithubComPatradenYaPracticumGoMartInternalAppDto(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *UserCredentials) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson9043616eDecodeGithubComPatradenYaPracticumGoMartInternalAppDto(l, v)
}
