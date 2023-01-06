package null

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
)

// Struct is a nullable struct. It supports SQL and JSON serialization.
// It will marshal to null if null.
type Struct[T any] struct {
	Struct T
	Valid  bool
}

// NewString creates a new Struct
func NewStruct[T any](object T, valid bool) Struct[T] {
	return Struct[T]{
		Struct: object,
		Valid:  valid,
	}
}

// StructFrom creates a new Struct that will never be blank.
func StructFrom[T any](object T) Struct[T] {
	return NewStruct(object, true)
}

// StructFromPtr creates a new Struct that be null if s is nil.
func StructFromPtr[T any](object *T) Struct[T] {
	if object == nil {
		var noop T
		return NewStruct(noop, false)
	}
	return NewStruct(*object, true)
}

// ValueOrZero returns the inner value if valid, otherwise zero.
func (s Struct[T]) ValueOrZero() T {
	if !s.Valid {
		var noop T
		return noop
	}
	return s.Struct
}

// UnmarshalJSON implements json.Unmarshaler.
// It supports unmarshaling the inner Struct and null input.
func (s *Struct[T]) UnmarshalJSON(data []byte) error {
	var v interface{}
	var err error
	if err = json.Unmarshal(data, &v); err != nil {
		return err
	}

	switch v.(type) {
	case map[string]interface{}:
		err = json.Unmarshal(data, &s.Struct)
	case nil:
		s.Valid = false
		return nil
	default:
		err = fmt.Errorf("json: cannot unmarshal %v into Go value of type %T", reflect.TypeOf(v).Name(), reflect.TypeOf(*new(T)).Name())
	}
	s.Valid = err == nil
	return err
}

// MarshalJSON implements json.Marshaler.
// It will encode null if this Struct is null.
func (s Struct[T]) MarshalJSON() ([]byte, error) {
	if !s.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(s.Struct)
}

// SetValid changes this Struct's value and also sets it to be non-null.
func (s *Struct[T]) SetValid(object T) {
	s.Struct = object
	s.Valid = true
}

func (s *Struct[T]) Ptr() *T {
	if !s.Valid {
		return nil
	}
	return &s.Struct
}

// IsZero returns true for invalid Structs.
// A non-null default Struct will not be considered zero.
func (s *Struct[T]) IsZero() bool {
	return !s.Valid
}

// Value implements the driver Valuer interface.
func (s *Struct[T]) Value() (driver.Value, error) {
	if !s.Valid {
		return nil, nil
	}
	return json.Marshal(s.Struct)
}

// Scan implements the Scanner interface.
func (s *Struct[T]) Scan(value interface{}) error {
	var err error
	switch x := value.(type) {
	case nil:
		s.Valid = false
		return nil
	case []byte:
		err = json.Unmarshal(x, &s.Struct)
	default:
		err = fmt.Errorf("null: cannot scan type %T into null.Struct: %v", value, value)
	}

	s.Valid = err == nil
	return err
}
