// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package dto

import (
	json "encoding/json"

	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"

	model "github.com/patraden/ya-practicum-go-mart/internal/app/domain/model"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjsonAfcbe245DecodeGithubComPatradenYaPracticumGoMartInternalAppDto(in *jlexer.Lexer, out *OrderStatusResponseBatch) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		in.Skip()
		*out = nil
	} else {
		in.Delim('[')
		if *out == nil {
			if !in.IsDelim(']') {
				*out = make(OrderStatusResponseBatch, 0, 1)
			} else {
				*out = OrderStatusResponseBatch{}
			}
		} else {
			*out = (*out)[:0]
		}
		for !in.IsDelim(']') {
			var v1 OrderStatusResponse
			(v1).UnmarshalEasyJSON(in)
			*out = append(*out, v1)
			in.WantComma()
		}
		in.Delim(']')
	}
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonAfcbe245EncodeGithubComPatradenYaPracticumGoMartInternalAppDto(out *jwriter.Writer, in OrderStatusResponseBatch) {
	if in == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
		out.RawString("null")
	} else {
		out.RawByte('[')
		for v2, v3 := range in {
			if v2 > 0 {
				out.RawByte(',')
			}
			(v3).MarshalEasyJSON(out)
		}
		out.RawByte(']')
	}
}

// MarshalJSON supports json.Marshaler interface
func (v OrderStatusResponseBatch) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonAfcbe245EncodeGithubComPatradenYaPracticumGoMartInternalAppDto(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v OrderStatusResponseBatch) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonAfcbe245EncodeGithubComPatradenYaPracticumGoMartInternalAppDto(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *OrderStatusResponseBatch) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonAfcbe245DecodeGithubComPatradenYaPracticumGoMartInternalAppDto(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *OrderStatusResponseBatch) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonAfcbe245DecodeGithubComPatradenYaPracticumGoMartInternalAppDto(l, v)
}
func easyjsonAfcbe245DecodeGithubComPatradenYaPracticumGoMartInternalAppDto1(in *jlexer.Lexer, out *OrderStatusResponse) {
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
		case "number":
			out.ID = string(in.String())
		case "status":
			out.Status = model.Status(in.String())
		case "accrual":
			out.Accrual = float64(in.Float64())
		case "uploaded_at":
			out.CreatedAt = string(in.String())
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
func easyjsonAfcbe245EncodeGithubComPatradenYaPracticumGoMartInternalAppDto1(out *jwriter.Writer, in OrderStatusResponse) {
	out.RawByte('{')
	first := true
	_ = first
	if in.ID != "" {
		const prefix string = ",\"number\":"
		first = false
		out.RawString(prefix[1:])
		out.String(string(in.ID))
	}
	if in.Status != "" {
		const prefix string = ",\"status\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Status))
	}
	if in.Accrual != 0 {
		const prefix string = ",\"accrual\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Float64(float64(in.Accrual))
	}
	if in.CreatedAt != "" {
		const prefix string = ",\"uploaded_at\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.CreatedAt))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v OrderStatusResponse) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonAfcbe245EncodeGithubComPatradenYaPracticumGoMartInternalAppDto1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v OrderStatusResponse) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonAfcbe245EncodeGithubComPatradenYaPracticumGoMartInternalAppDto1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *OrderStatusResponse) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonAfcbe245DecodeGithubComPatradenYaPracticumGoMartInternalAppDto1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *OrderStatusResponse) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonAfcbe245DecodeGithubComPatradenYaPracticumGoMartInternalAppDto1(l, v)
}
func easyjsonAfcbe245DecodeGithubComPatradenYaPracticumGoMartInternalAppDto2(in *jlexer.Lexer, out *OrderStatusAccrual) {
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
		case "order":
			out.ID = string(in.String())
		case "status":
			out.Status = model.Status(in.String())
		case "accrual":
			if data := in.Raw(); in.Ok() {
				in.AddError((out.Accrual).UnmarshalJSON(data))
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
func easyjsonAfcbe245EncodeGithubComPatradenYaPracticumGoMartInternalAppDto2(out *jwriter.Writer, in OrderStatusAccrual) {
	out.RawByte('{')
	first := true
	_ = first
	if in.ID != "" {
		const prefix string = ",\"order\":"
		first = false
		out.RawString(prefix[1:])
		out.String(string(in.ID))
	}
	if in.Status != "" {
		const prefix string = ",\"status\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Status))
	}
	if true {
		const prefix string = ",\"accrual\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Raw((in.Accrual).MarshalJSON())
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v OrderStatusAccrual) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonAfcbe245EncodeGithubComPatradenYaPracticumGoMartInternalAppDto2(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v OrderStatusAccrual) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonAfcbe245EncodeGithubComPatradenYaPracticumGoMartInternalAppDto2(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *OrderStatusAccrual) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonAfcbe245DecodeGithubComPatradenYaPracticumGoMartInternalAppDto2(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *OrderStatusAccrual) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonAfcbe245DecodeGithubComPatradenYaPracticumGoMartInternalAppDto2(l, v)
}