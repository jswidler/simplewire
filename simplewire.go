// Package simplewire is a dependency injection module that works using struct tags.
package simplewire

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

// Initializable can be implemented to automatically run initialization logic before dependency injection.
type Initializable interface {
	Init() error
}

// Connect will create a set of dependencies which can be injected by using the returned Injector.
// Each field in the reference that is eligible to be injected will also have its own dependencies injected.
// The reference interface should be a struct or pointer to a struct.
func Connect(tag string, reference interface{}) (Injector, error) {
	injector := injector{tag, reflect.Indirect(reflect.ValueOf(reference))}
	return injector, injector.Inject(getFields(reference)...)
}

type Injector interface {
	// Inject will iterate through each dest to inject dependencies. If a dest implements simplewire.Initializable, the Init method will be called.
	Inject(dest ...interface{}) error
}

type injector struct {
	tag       string
	reference reflect.Value
}

// Inject will iterate through each dest to inject dependencies. If a dest implements simplewire.Initializable, the Init method will be called.
func (i injector) Inject(dest ...interface{}) error {
	for _, d := range dest {
		if d == nil {
			continue
		}
		if hasInit, ok := d.(Initializable); ok {
			err := hasInit.Init()
			if err != nil {
				return err
			}
		}
		err := i.injectSingle(d)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i injector) injectSingle(dest interface{}) (err error) {
	// in case of panic, preserve the names of the field that was being worked on
	destStructName := ""
	destFieldName := ""
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("simplewire inject failed at %s:%s", destStructName, destFieldName)
		}
	}()

	// get value of the struct that is being injected
	destValue := reflect.ValueOf(dest)
	destValue = dereference(destValue)

	destStructName = destValue.Type().Name()
	if destValue.Kind() == reflect.Struct {
		// for each field in the dest struct
		for x := 0; x < destValue.NumField(); x++ {
			destField := destValue.Type().Field(x)
			destFieldName = destField.Name
			// check if it has a tag with the inject key
			refFieldName := destField.Tag.Get(i.tag)
			if refFieldName == "" {
				continue
			}
			// if so, find the field in the reference
			refField, err := i.getRefFieldByName(refFieldName)
			if err != nil {
				if err == errFieldNotFound {
					return fmt.Errorf("simplewire inject failed at %s:%s - %s not found in reference struct", destStructName, destFieldName, refFieldName)
				} else if err == errFieldNotExported {
					return fmt.Errorf("simplewire inject failed at %s:%s - %s must be exported from reference struct", destStructName, destFieldName, refFieldName)
				}
				panic(err) // no other error type is expected, but the panic is caught
			}

			destFieldValue := destValue.FieldByIndex([]int{x})
			refFieldValue := reflect.ValueOf(refField)
			// Check we will be able to set the destination field
			if !unicode.IsUpper(rune(destFieldName[0])) {
				return fmt.Errorf("simplewire inject failed at %s:%s - %s cannot be private", destStructName, destFieldName, destFieldName)
			} else if destFieldValue.Kind() != reflect.Ptr && destFieldValue.Kind() != reflect.Interface {
				return fmt.Errorf("simplewire inject failed at %s:%s - %s must be a pointer or interface", destStructName, destFieldName, destFieldName)
			} else if !destFieldValue.CanSet() {
				return fmt.Errorf("simplewire inject failed at %s:%s - %s cannot be changed", destStructName, destFieldName, destFieldName)
			} else if !refFieldValue.Type().AssignableTo(destFieldValue.Type()) {
				return fmt.Errorf("simplewire inject failed at %s:%s - %s is not assignable to %s", destStructName, destFieldName, refFieldValue.Type(), destFieldValue.Type())
			}
			destFieldValue.Set(refFieldValue)
		}
	}
	return nil
}

var (
	errFieldNotFound    = errors.New("field not found")
	errFieldNotExported = errors.New("field not exported")
)

func (i injector) getRefFieldByName(name string) (interface{}, error) {
	lname := strings.ToLower(name)
	f := i.reference.FieldByNameFunc(func(n string) bool {
		return strings.ToLower(n) == lname
	})
	if !f.IsValid() {
		return nil, errFieldNotFound
	} else if !f.CanInterface() {
		return nil, errFieldNotExported
	}
	return f.Interface(), nil
}

// getFields will return a slice containing the values of all the exported fields of s
func getFields(s interface{}) []interface{} {
	v := reflect.ValueOf(s)
	v = dereference(v)
	fields := []interface{}{}
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.CanInterface() {
			fields = append(fields, field.Interface())
		}
	}
	return fields
}

func dereference(v reflect.Value) reflect.Value {
	for {
		kind := v.Type().Kind()
		if kind == reflect.Interface || kind == reflect.Ptr {
			v = v.Elem()
		} else {
			return v
		}
	}
}
