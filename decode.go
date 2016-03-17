package railing

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

var errMissingData = func(typ reflect.Type) error {
	return fmt.Errorf(
		"%s. every slice element must contain the same amount of data",
		&UnmarshalTypeError{"object", typ})
}

// UnmarshalTypeError describes an url.Value's value that was not appropriate
// for a value of a specific Go type.
type UnmarshalTypeError struct {
	Value string
	Type  reflect.Type
}

func (e *UnmarshalTypeError) Error() string {
	return "railing: cannot unmarshal " + e.Value + " into Go value of type " +
		e.Type.String()
}

// An InvalidUnmarshalError describes an invalid argument passed to Unmarshal.
// (The argument to Unmarshal must be a non-nil pointer.)
type InvalidUnmarshalError struct {
	Type reflect.Type
}

func (e *InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "railing: Unmarshal(nil)"
	}

	if e.Type.Kind() != reflect.Ptr {
		return "railing: Unmarshal(non-pointer " + e.Type.String() + ")"
	}
	return "railing: Unmarshal(nil " + e.Type.String() + ")"
}

// Unmarshaler is the interface implemented by objects that can unmarshal
// an url.Values description of themselves. The input contains keys and values
// for the current object. If the object is a nested one and the original map
// contains eg. "foo[name]" then the input will contain stripped keys - in this
// case - "name".
type Unmarshaler interface {
	UnmarshalQuery(Values) error
}

// Unmarshal parses the url.Value data and stores the result
// in the value pointed to by v.
//
// Unmarshal supports the url.Values which were created from parsing the rails
// style query params.
//
// Unmarshal can be used to unmarshal the data into structs and maps.
//
// Unmarshal is allocating maps, slices, and pointers as necessary.
//
// To unmarshal into an interface value, Unmarshals creates
// map[string]interface{} where any key which can be an object (eg. "foo[name]")
// becomes another map[string]interface{}
//
// To unmarshal into a map, unmarshal creates a new map where the key must be
// of a string type and tries to fill the data according to the given type.
//
// To unmarshal into a struct, Unmarshal matches incoming keys to the struct's
// field names or tags. If a field is a slice and tag contains comma option,
// unmarshal will try to decode the value by splitting it by comma. Fields are
// being unmarshaled before the embedded structs. If an embedded struct contains
// a field with the same tag as the top level struct then only the top level
// field will be filled.
//
// BUG(jszwec) If the struct contains the array of structs, due to url.Values
// structure, every element (object) of the array, must contain the same amount
// of data; if not it is not possible to say where certain elements belong.
func Unmarshal(m Values, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &InvalidUnmarshalError{reflect.TypeOf(v)}
	}
	mcopy := make(url.Values)
	for k, v := range m.Values {
		mcopy[k] = v
	}
	return (&decoder{}).unmarshal(mcopy, rv)
}

type decoder struct{}

// indirect walks down v allocating pointers as needed, until it gets to a
// non-pointer. if it encounters an Unmarshaler, indirect stops and returns it.
func (d *decoder) indirect(v reflect.Value) (Unmarshaler, reflect.Value) {
	if v.Kind() != reflect.Ptr && v.Type().Name() != "" && v.CanAddr() {
		v = v.Addr()
	}
	for {
		if v.Kind() == reflect.Interface && !v.IsNil() {
			e := v.Elem()
			if e.Kind() == reflect.Ptr && !e.IsNil() {
				v = e
				continue
			}
		}
		if v.Kind() != reflect.Ptr {
			break
		}
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		if v.Type().NumMethod() > 0 {
			if u, ok := v.Interface().(Unmarshaler); ok {
				return u, reflect.Value{}
			}
		}
		v = v.Elem()
	}
	return nil, v
}

// objectInterface builds a map[string]interface{} from the given url.Values.
// Nested objects will become another map[string]interface{}
// Example
//
// foo[name] will be turned into
//
// map[string]interface{}{
//  "foo":map[string]interface{}{
//    "name": name,
//  },
// }
//
// Any array key eg. "array[]" will be stripped from "[]".
func (d *decoder) objectInterface(values url.Values) map[string]interface{} {
	m := make(map[string]interface{})
	for k, v := range values {
		match := reObject.FindStringSubmatch(k)
		if match != nil {
			m[match[1]] = d.objectInterface(subMap(values, match[1]))
			continue
		}
		m[strings.TrimSuffix(k, "[]")] = v
	}
	return m
}

// maps builds a map of the given type filling it with the data from url.Values.
// Nested keys remain as they are unless the map type is map[string]interface{}.
// Any array key eg. "array[]" will be stripped from "[]".
func (d *decoder) maps(values url.Values, v reflect.Value) error {
	if v.Type() == reflect.TypeOf(map[string]interface{}{}) {
		v.Set(reflect.ValueOf(d.objectInterface(values)))
		return nil
	}
	typ := v.Type()
	if typ.Key().Kind() != reflect.String {
		return &UnmarshalTypeError{"object", typ}
	}
	m := reflect.MakeMap(typ)
	for k, values := range values {
		newVal := reflect.Indirect(reflect.New(typ.Elem()))
		if err := d.conv(values, newVal); err != nil {
			return err
		}
		m.SetMapIndex(reflect.ValueOf(strings.TrimSuffix(k, "[]")), newVal)
	}
	v.Set(m)
	return nil
}

// slice builds a slice of the given type and attempts to translate the data
// from value arg.
func (d *decoder) slice(value []string, v reflect.Value) error {
	slice := reflect.MakeSlice(v.Type(), len(value), len(value))
	for i := 0; i < len(value); i++ {
		if err := d.conv([]string{value[i]}, slice.Index(i)); err != nil {
			return err
		}
	}
	v.Set(slice)
	return nil
}

// array builds a slice of the given type and attempts to translate the data
// from value arg and then copies it to the given array.
func (d *decoder) array(value []string, v reflect.Value) error {
	slice := reflect.MakeSlice(
		reflect.SliceOf(v.Type().Elem()), len(value), len(value))
	for i := 0; i < len(value); i++ {
		if err := d.conv([]string{value[i]}, slice.Index(i)); err != nil {
			return err
		}
	}
	reflect.Copy(v, slice)
	return nil
}

// conv attempts to convert a single url.Value's value to the v's type.
func (d *decoder) conv(value []string, v reflect.Value) error {
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return d.conv(value, v.Elem())
	case reflect.Interface:
		if v.NumMethod() == 0 {
			v.Set(reflect.ValueOf(value))
		} else {
			return &UnmarshalTypeError{"object", v.Type()}
		}
	case reflect.Slice:
		if err := d.slice(value, v); err != nil {
			return err
		}
	case reflect.Array:
		if err := d.array(value, v); err != nil {
			return err
		}
	case reflect.String:
		if len(value) >= 1 {
			v.SetString(value[0])
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if len(value) >= 1 {
			n, err := strconv.ParseInt(value[0], 10, 64)
			if err != nil || v.OverflowInt(n) {
				return &UnmarshalTypeError{"number " + value[0], v.Type()}
			}
			v.SetInt(n)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64:
		if len(value) >= 1 {
			n, err := strconv.ParseUint(value[0], 10, 64)
			if err != nil || v.OverflowUint(n) {
				return &UnmarshalTypeError{"number " + value[0], v.Type()}
			}
			v.SetUint(n)
		}
	case reflect.Float32, reflect.Float64:
		if len(value) >= 1 {
			n, err := strconv.ParseFloat(value[0], v.Type().Bits())
			if err != nil || v.OverflowFloat(n) {
				return &UnmarshalTypeError{"number " + value[0], v.Type()}
			}
			v.SetFloat(n)
		}
	case reflect.Bool:
		if len(value) >= 1 {
			b, err := strconv.ParseBool(value[0])
			if err != nil {
				return &UnmarshalTypeError{"bool " + value[0], v.Type()}
			}
			v.SetBool(b)
		}
	}
	return nil
}

func (d *decoder) unmarshal(values url.Values, v reflect.Value) error {
	u, v := d.indirect(v)
	if u != nil {
		return u.UnmarshalQuery(Values{values})
	}
	if v.Kind() == reflect.Interface && v.NumMethod() == 0 {
		v.Set(reflect.ValueOf(d.objectInterface(values)))
		return nil
	}
	switch v.Kind() {
	case reflect.Map:
		return d.maps(values, v)
	case reflect.Struct:
		return d.object(values, v)
	default:
		return &UnmarshalTypeError{"object", v.Type()}
	}
}

// fields returns a slice of fields indexes which are ordered in such a way
// that embedded fields are at the end. It returns indexes only of the exported
// fields.
func (d *decoder) fields(typ reflect.Type) (indexes []int) {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" && !field.Anonymous { // unexported
			continue
		}
		if field.Anonymous {
			indexes = append(indexes, i)
		} else {
			indexes = append([]int{i}, indexes...)
		}
	}
	return
}

// indexedObject attempts to unmarshal the data in m to the slice or array of
// structs under v.
func (d *decoder) indexedObject(m url.Values, v reflect.Value) error {
	switch v.Kind() {
	case reflect.Array:
		slice, err := d.sliceObject(m, reflect.SliceOf(v.Type().Elem()))
		if err != nil {
			return err
		}
		reflect.Copy(v, slice)
	case reflect.Slice:
		slice, err := d.sliceObject(m, v.Type())
		if err != nil {
			return err
		}
		v.Set(slice)
	}
	return nil
}

// sliceObject is used to unmarshal an array of structs. It returns a slice of
// structs filled with the data from m. First it iterates m to check if every
// value is of the same length. It is neccessary that the data is complete and
// every key contains the same data. We divide the data in the map by index.
//
// For example - m[key1][0] and m[key2][0] describe the same struct, m[key1][1]
// and m[key2][1] describe the second element of an array.
//
// url.Values{
// 	"key1": []string{"a", "b"},
// 	"key2": []string{"1", "2"},
// }
//
// Element 1: {"a", "1"}
// Element 2: {"b", "2"}
//
// If the caller wants an array as part of the element, then the best workaround
// would be to make sure that the array is being sent as a string separated
// with some character, and then implement a type with custom Unmarshaler
// to handle it or use comma tag. Look at examples.
func (d *decoder) sliceObject(m url.Values,
	typ reflect.Type) (reflect.Value, error) {
	l := 0
	keys := []string{}
	for k, vv := range m {
		keys = append(keys, k)
		if l == 0 {
			l = len(vv)
			continue
		}
		if len(vv) != l {
			return reflect.Value{}, errMissingData(typ)
		}
	}
	slice := reflect.MakeSlice(typ, l, l)
	for i := 0; i < l; i++ {
		mm := make(url.Values)
		for _, key := range keys {
			mm.Set(key, m[key][i])
		}
		if err := d.unmarshal(mm, slice.Index(i)); err != nil {
			return reflect.Value{}, err
		}
	}
	return slice, nil
}

// object is used to unmarshal into a struct. It iterates over ordered fields
// and unmarshals, first the normal fields and then embedded fields.
//
// If the field is embedded, then unmarshal starts over again with the same
// url.Values. However, the keys which were already used are no longer in the
// map - every key should be used only once.
//
// If the key in a map is a nested struct then the sub-map is being created in
// order to unmarshal it. For example - "foo[name]" key will turn into "name" in
// the new map.
//
// If the key matches the tag and its not a nested struct then a normal
// convertion can be used.
//
// If the type implements Unmarshaler interface then UnmarshalQuery will be
// used instead of conv function.
func (d *decoder) object(m url.Values, v reflect.Value) (err error) {
	typ := v.Type()
	for _, i := range d.fields(typ) {
		fieldType := typ.Field(i)
		tag := parseTag(fieldType)
		if tag.ignore {
			continue
		}
		v := v.Field(i)
		if fieldType.Anonymous {
			if err := d.unmarshal(m, v); err != nil {
				return err
			}
			continue
		}
		subm, values := findValues(m, tag.name)
		if subm != nil {
			switch v.Kind() {
			case reflect.Slice, reflect.Array:
				if err := d.indexedObject(subm, v); err != nil {
					return err
				}
			default:
				if err := d.unmarshal(subm, v); err != nil {
					return err
				}
			}
			continue
		}
		if values != nil {
			u, v := d.indirect(v)
			if u != nil {
				if err := u.UnmarshalQuery(Values{m}); err != nil {
					return err
				}
				continue
			} else {
				if tag.comma {
					values = d.splitValues(values, ",")
				}
				if err := d.conv(values, v); err != nil {
					return err
				}
			}
			delete(m, tag.name)
			continue
		}
	}
	return nil
}

func (d *decoder) splitValues(values []string, sep string) (res []string) {
	for _, v := range values {
		res = append(res, strings.Split(v, sep)...)
	}
	return
}
