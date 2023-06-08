package null

import (
	"encoding/json"
	"reflect"
	"testing"
)

type Dummy struct {
	Value int64 `json:"value"`
}

func TestStructFrom(t *testing.T) {
	s := StructFrom(Dummy{Value: 10})
	assertStruct(t, s, Dummy{Value: 10})
}

func TestNewStructValid(t *testing.T) {
	s := NewStruct(Dummy{Value: 10}, true)
	assertNullStruct(t, s, Struct[Dummy]{Struct: Dummy{Value: 10}, Valid: true})
}

func TestNewStructInvalid(t *testing.T) {
	s := NewStruct(Dummy{}, false)
	assertNullStruct(t, s, Struct[Dummy]{Struct: Dummy{}, Valid: false})
}

func TestStructFromPtr(t *testing.T) {
	s := StructFromPtr(&Dummy{Value: 10})
	assertNullStruct(t, s, Struct[Dummy]{Struct: Dummy{Value: 10}, Valid: true})
}

func TestStructFromPtrNull(t *testing.T) {
	s := StructFromPtr[Dummy](nil)
	assertNullStruct(t, s, Struct[Dummy]{Struct: Dummy{}, Valid: false})
}

func TestValueOrZero(t *testing.T) {
	s := StructFrom(Dummy{Value: 10})
	assertObject(t, s.ValueOrZero(), Dummy{Value: 10})
}

func TestValueOrZeroDefault(t *testing.T) {
	s := NewStruct(Dummy{Value: 10}, false)
	assertObject(t, s.ValueOrZero(), Dummy{})
}

func TestUnmarshalJSON(t *testing.T) {
	var s Struct[Dummy]
	err := json.Unmarshal([]byte(`{"value": 10}`), &s)
	maybePanic(err)
	assertStruct(t, s, Dummy{Value: 10})

	var sd Struct[Dummy]
	err = json.Unmarshal([]byte(`{}`), &sd)
	maybePanic(err)
	assertStruct(t, sd, Dummy{})

	var sn Struct[Dummy]
	err = json.Unmarshal([]byte(`null`), &sn)
	maybePanic(err)
	assertNullStruct(t, sn, StructFromPtr[Dummy](nil))

	var si Struct[Dummy]
	err = json.Unmarshal([]byte(`invalid`), &si)
	if err == nil {
		panic("err should not be nil")
	}
	assertNullStruct(t, si, StructFromPtr[Dummy](nil))

	var si2 Struct[Dummy]
	err = json.Unmarshal([]byte(`{"value": "string"}`), &si2)
	if err == nil {
		panic("err should not be nil")
	}
	assertNullStruct(t, si, StructFromPtr[Dummy](nil))

	type customString string
	var ss Struct[customString]
	err = json.Unmarshal([]byte(`"hello"`), &ss)
	maybePanic(err)
	if ss.Struct != "hello" {
		panic("should be hello")
	}
}

func TestMarshalStruct(t *testing.T) {
	s := StructFrom(Dummy{10})
	data, err := json.Marshal(s)
	maybePanic(err)
	assertJSONEquals(t, data, `{"value":10}`, "non-empty json marshal")

	// invalid values should be encoded as null
	s = NewStruct(Dummy{10}, false)
	data, err = json.Marshal(s)
	maybePanic(err)
	assertJSONEquals(t, data, `null`, "non-empty json marshal")

	// String marshal
	s2 := NewStruct("test", true)
	data, err = json.Marshal(s2)
	maybePanic(err)
	assertJSONEquals(t, data, `"test"`, "non-empty json marshal")
}

func TestStructPtr(t *testing.T) {
	s := StructFrom(Dummy{10})
	ptr := s.Ptr()
	if ptr.Value != 10 {
		t.Errorf("bad %s Struct: %v ≠ %v\n", "pointer", ptr, 10)
	}

	null := NewStruct(Dummy{10}, false)
	ptr = null.Ptr()
	if ptr != nil {
		t.Errorf("bad %s float: %#v ≠ %s\n", "nil pointer", ptr, "nil")
	}
}

func TestStructSetValid(t *testing.T) {
	s := NewStruct(Dummy{}, false)
	s.SetValid(Dummy{Value: 10})
	assertNullStruct(t, s, NewStruct(Dummy{10}, true))
}

func TestStructIsZero(t *testing.T) {
	s := NewStruct(Dummy{}, false)
	if !s.IsZero() {
		t.Errorf("Struct{%v, Valid:%t} is not zero", s.Struct, s.Valid)
	}
	s.SetValid(Dummy{Value: 10})
	if s.IsZero() {
		t.Errorf("Struct{%v, Valid:%t} is zero", s.Struct, s.Valid)
	}
}

func TestStructValue(t *testing.T) {
	s := NewStruct(Dummy{}, false)
	if v, err := s.Value(); v != nil || err != nil {
		t.Errorf("Struct{%v, Valid:%t} value error {%v, err:%v}", s.Struct, s.Valid, v, err)
	}
	s.SetValid(Dummy{Value: 10})
	if v, err := s.Value(); !reflect.DeepEqual(v, []byte(`{"value":10}`)) || err != nil {
		t.Errorf("Struct{%v, Valid:%t} value error {%v, err:%v}", s.Struct, s.Valid, v, err)
	}
}

func TestStructScan(t *testing.T) {
	s := Struct[Dummy]{}
	err := s.Scan([]byte(`{"value": 10}`))
	maybePanic(err)
	assertNullStruct(t, s, StructFrom(Dummy{Value: 10}))

	s = Struct[Dummy]{}
	err = s.Scan(nil)
	maybePanic(err)
	assertNullStruct(t, s, Struct[Dummy]{})
}

func assertStruct(t *testing.T, a Struct[Dummy], from Dummy) {
	t.Helper()
	if from.Value != a.Struct.Value {
		t.Errorf("Struct{%v, Valid:%t} and Dummy{%v} not equal", a.Struct, a.Valid, from)
	}
}

func assertObject(t *testing.T, a, b Dummy) {
	t.Helper()
	if a.Value != b.Value {
		t.Errorf("Dummy{%v} and Dummy{%v} not equal", a, b)
	}
}

func assertNullStruct(t *testing.T, a, b Struct[Dummy]) {
	t.Helper()
	if a.Valid != b.Valid || a.Struct.Value != b.Struct.Value {
		t.Errorf("Struct{%v, Valid:%t} and Stuct{%v, Valid: %t} not equal", a.Struct, a.Valid, b.Struct, b.Valid)
	}
}
