package railing

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

var marshalerType = reflect.TypeOf(new(Marshaler)).Elem()

// Marshaler is the interface implemented by objects that can marshal themselves
// into Values.
type Marshaler interface {
	MarshalQuery() (Values, error)
}

// An UnsupportedTypeError is returned by Marshal when attempting to encode an
// unsupported value type.
type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return "railing: unsupported type: " + e.Type.String()
}

// A MarshalerError is returned by Marshal when an attempt to encode a Marshaler
// object fails.
type MarshalerError struct {
	Type reflect.Type
	Err  error
}

func (e *MarshalerError) Error() string {
	return "railing: error calling MarshalQuery for type " + e.Type.String() +
		": " + e.Err.Error()
}

// Marshal returns structs or maps encoded into Values with keys compatible
// with rails query param style.
//
// To marshal a struct into values, Marshal walks every exported field and it
// will find itself in the result unless:
//    - the field's tag is "-", or
//    - the field is empty and its tag specifies the "omitempty" option.
// The empty values are false, 0, any nil pointer or interface value, and any
// array, slice, map, or string of length zero. The object's default key string
// is the struct field name but can be specified in the struct field's tag
// value. The "railing" key in the struct field's tag value is the key name,
// followed by an optional comma and options. Examples:
//
//   // Field is ignored by this package.
//   Field int `railing:"-"`
//
//   // Field appears in Values as key "myName".
//   Field int `railing:"myName"`
//
//   // Field appears in Values as key "myName" and
//   // the field is omitted from the object if its value is empty,
//   // as defined above.
//   Field int `railing:"myName,omitempty"`
//
//   // Field appears in Values as key "Field" (the default), but
//   // the field is skipped if empty.
//   // Note the leading comma.
//   Field int `railing:",omitempty"`
//
//   // Field appears in Values as key "slice" - elements are joined by ','
//   // character because of 'comma' option.
//   Field []int `railing:"slice,comma"`
//
// Anonymous struct fields are marshaled as if their inner exported fields were
// fields in the outer struct. An anonymous struct field with a name given in
// its railing tag is treated as having that name, rather than being anonymous.
//
// To marshal a map into values, a map with string keys is required. Marshal
// will walk through every key trying to encode values. If a map is of type
// map[string]interface{} then the values can be a nested struct or other map
// producing a valid rails style structure.
//
// Marshal can encode values of types string, int, float, bool.
//
// Arrays and slices of simple types are easily encoded into []string. However,
// in case of arrays or slices of structs then every struct's field will have
// its own key and it will contain values of other elements in the value,
// according to the rails style params.
//    type A struct {
//      Int int
//      Str string
//    }
//
//    type B struct {
//      As []A `railing:"as"`
//    }
//
//    // where the slice is []A{{1,"str_1"}, {2, "str_2"}}
//    // the result will be
//    url.Values {
//      "as[][Int]": []string{"1", "2"},
//      "as[][Str]": []string{"str_1", "str_2"},
//    }
//
// WARNING: if you are willing to encode []interface{} value then keep in mind
// that in order to be encoded correctly it should contain only primitive types.
// If a struct or a map will be used as an element then the error will be
// returned.
func Marshal(v interface{}) (Values, error) {
	m, err := (&encoder{}).marshal(reflect.ValueOf(v))
	if err != nil {
		return Values{}, err
	}
	return Values{m}, nil
}

type encoder struct{}

func (e *encoder) marshal(v reflect.Value) (m url.Values, err error) {
	m = make(url.Values)
	v = e.indirect(v)
	if !v.IsValid() {
		return m, nil
	}
	if m := e.marshaler(v); m != nil {
		values, err := m.MarshalQuery()
		if err != nil {
			return nil, &MarshalerError{v.Type(), err}
		}
		return values.Values, nil
	}
	switch v.Kind() {
	case reflect.Map:
		err = e.maps(m, v)
	case reflect.Struct:
		err = e.object(m, v)
	default:
		err = &UnsupportedTypeError{v.Type()}
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

// object encodes the given struct. It walks every field ignoring unexported
// ones and these with the tag "-".
func (e *encoder) object(values url.Values, v reflect.Value) error {
	for i := 0; i < v.NumField(); i++ {
		typ := v.Type().Field(i)
		if typ.PkgPath != "" && !typ.Anonymous {
			continue
		}
		tag := parseTag(typ)
		if tag.ignore || tag.omitEmpty && isEmptyValue(v.Field(i)) {
			continue
		}
		if typ.Anonymous && tag.empty {
			if err := e.marshalEmbedded(values, v.Field(i)); err != nil {
				return err
			}
			continue
		}
		if err := e.marshalField(values, v.Field(i), tag); err != nil {
			return err
		}
	}
	return nil
}

func (e *encoder) marshalField(values url.Values, v reflect.Value,
	tag tag) error {
	v = e.indirect(v)
	if !v.IsValid() {
		return nil
	}
	if m := e.marshaler(v); m != nil {
		subm, err := m.MarshalQuery()
		if err != nil {
			return &MarshalerError{v.Type(), err}
		}
		e.mergeByKey(tag.name, subm.Values, values)
		return nil
	}
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		if err := e.slices(tag, values, v); err != nil {
			return err
		}
	case reflect.Struct:
		s, err := e.marshal(v)
		if err != nil {
			return err
		}
		e.mergeByKey(tag.name, s, values)
	default:
		str, err := e.conv(v)
		if err != nil {
			return err
		}
		values.Set(tag.name, str)
	}
	return nil
}

func (e *encoder) marshalEmbedded(values url.Values, v reflect.Value) error {
	s, err := e.marshal(v)
	if err != nil {
		return err
	}
	for k, v := range s {
		if _, ok := values[k]; !ok {
			values[k] = v
		}
	}
	return nil
}

// mergeByKey merges src Value with dst Value where key serves as an object's
// name.
//
// dst := Values {
//  foo: []string{"foo"},
// }
//
// src := Values {
//   "name": []string{"name"},
// }
//
// mergeByKey("bar", src, dst) will result in dst becoming the following:
//
// Values {
//  "foo":       []string{"foo"},
//  "bar[name]": []string{"name"},
// }
//
func (e *encoder) mergeByKey(key string, src, dst url.Values) {
	for k, v := range src {
		match := reTopKey.FindStringSubmatch(k)
		if match != nil {
			dst[fmt.Sprintf("%s[%s]%s", key, match[1], match[2])] = v
		}
		if k == "" {
			dst[key] = v
		}
	}
}

// maps encodes maps into Values. The given map type must have a string as a
// key.
func (e *encoder) maps(values url.Values, v reflect.Value) error {
	if v.Type().Key().Kind() != reflect.String {
		return &UnsupportedTypeError{v.Type()}
	}
	if !v.IsValid() {
		return nil
	}
	for _, vkey := range v.MapKeys() {
		vv := e.indirect(v.MapIndex(vkey))
		if !vv.IsValid() {
			continue
		}
		switch vv.Kind() {
		case reflect.Slice, reflect.Array:
			if err := e.slices(tag{name: vkey.String()}, values,
				vv); err != nil {
				return err
			}
		case reflect.Struct:
			return &UnsupportedTypeError{v.Type()}
		case reflect.Map:
			m := make(url.Values)
			if err := e.maps(m, vv); err != nil {
				return err
			}
			e.mergeByKey(vkey.String(), m, values)
		default:
			s, err := e.conv(vv)
			if err != nil {
				return err
			}
			values.Set(vkey.String(), s)
		}
	}
	return nil
}

// structSlices encodes slices of structs by marshaling each one of them.
// Every struct's field must be encoded to one string, otherwise the final query
// string will become corrupted. That is why every slice inside the struct will
// be joined by a comma.
func (e *encoder) structSlices(tag tag, values url.Values,
	v reflect.Value) error {
	m := make(url.Values)
	for i := 0; i < v.Len(); i++ {
		s, err := e.marshal(v.Index(i))
		if err != nil {
			return err
		}
		for k, v := range s {
			m.Add(k, strings.Join(v, ","))
		}
	}
	e.mergeByKey(tag.name+"[]", m, values)
	return nil
}

// slices encodes slices into Values based on the given tag.
func (e *encoder) slices(tag tag, values url.Values, v reflect.Value) error {
	el := v.Type().Elem()
	switch el.Kind() {
	case reflect.Struct:
		return e.structSlices(tag, values, v)
	case reflect.Ptr:
		if el.Elem().Kind() == reflect.Struct {
			return e.structSlices(tag, values, v)
		}
	default:
	}
	if v.Len() < 1 {
		return nil
	}
	var strs []string
	for i := 0; i < v.Len(); i++ {
		vv := e.indirect(v.Index(i))
		if !vv.IsValid() {
			continue
		}
		str, err := e.conv(vv)
		if err != nil {
			return err
		}
		strs = append(strs, str)
	}
	if tag.comma {
		values.Set(tag.name, strings.Join(strs, ","))
		return nil
	}
	values[tag.name+"[]"] = strs
	return nil
}

// conv encodes simple types into string.
func (e *encoder) conv(v reflect.Value) (string, error) {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64), nil
	case reflect.String:
		return v.String(), nil
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), nil
	default:
		return "", &UnsupportedTypeError{v.Type()}
	}
}

// indirect walks down v, until it gets to a non-pointer.
func (e *encoder) indirect(v reflect.Value) reflect.Value {
	for {
		if v.Kind() == reflect.Interface || v.Kind() == reflect.Ptr {
			v = v.Elem()
		} else {
			break
		}
	}
	return v
}

// marshaler checks if v implements marshaler. If not, it returns nil marshaler.
func (e *encoder) marshaler(v reflect.Value) Marshaler {
	if v.Type().Implements(marshalerType) {
		return v.Interface().(Marshaler)
	}
	if v.CanAddr() {
		v = v.Addr()
		if v.Type().Implements(marshalerType) {
			return v.Interface().(Marshaler)
		}
	}
	return nil
}
