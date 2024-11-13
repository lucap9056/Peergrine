package configurator

import (
	"flag"
	"log"
	"os"
	"reflect"
)

// ReadConfigValues reads values from environment variables and command-line flags
// based on struct field tags. It returns a map where the keys are the config tags
// and the values are pointers to the corresponding string values.
func ReadConfigValues[T any](config *T) map[string]*string {
	configValues := reflect.ValueOf(config).Elem()
	configTypes := configValues.Type()

	insertValues := make(map[string]*string)

	for i := 0; i < configValues.NumField(); i++ {
		fieldType := configTypes.Field(i)

		tag := fieldType.Tag.Get("config")
		if tag == "" {
			continue
		}

		var value string
		env := os.Getenv(tag)
		flag.StringVar(&value, tag, env, fieldType.Name)

		insertValues[tag] = &value
	}

	flag.Parse()

	return insertValues
}

// ApplyConfigValues applies values from command-line flags and environment variables
// to the struct fields if they are non-empty and not "-".
func ApplyConfigValues[T any](config *T, insertValues map[string]*string) {
	configValues := reflect.ValueOf(config).Elem()

	if configValues.Kind() != reflect.Struct {
		return
	}

	configTypes := configValues.Type()

	for i := 0; i < configValues.NumField(); i++ {

		fieldType := configTypes.Field(i)

		tag := fieldType.Tag.Get("config")

		if tag == "" {
			continue
		}

		if insertValue, ok := insertValues[tag]; ok && insertValue != nil && *insertValue != "" {

			value := configValues.Field(i)

			if value.CanSet() {
				value.SetString(*insertValue)
			}

		}
	}
}

// ReadConfigValuesFromStruct reads the values of string fields from the provided struct
// and returns a map where the keys are the config tags and the values are the corresponding string values.
func ReadConfigValuesFromStruct[T any](config *T) map[string]*string {
	configValues := reflect.ValueOf(config).Elem()

	if configValues.Kind() != reflect.Struct {
		return nil
	}

	configTypes := configValues.Type()

	values := make(map[string]*string)

	for i := 0; i < configValues.NumField(); i++ {
		fieldType := configTypes.Field(i)
		tag := fieldType.Tag.Get("config")

		if tag == "" {
			continue
		}

		fieldValue := configValues.Field(i)

		if fieldValue.Kind() == reflect.String {
			value := fieldValue.String()
			if value != "" {
				values[tag] = &value
			}
		}

	}

	return values
}

// ReplaceDashWithEmpty replaces any "-" in the string fields of the struct with the correct empty value
// (empty string, 0, etc.) for different field types like string, int, and float64.
func ReplaceDashWithEmpty[T any](config *T) {
	v := reflect.ValueOf(config).Elem()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)

		switch field.Kind() {
		case reflect.String:
			if field.String() == "-" {
				field.SetString("")
			}
		case reflect.Int:
			if field.Int() == -1 {
				field.SetInt(0)
			}
		case reflect.Float64:
			if field.Float() == -1 {
				field.SetFloat(0.0)
			}
		}
	}
}

// Print prints the values of the struct fields to the log based on their config tags.
func Print[T any](config *T) {
	configValue := reflect.ValueOf(config).Elem()

	if configValue.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < configValue.NumField(); i++ {
		field := configValue.Type().Field(i)
		fieldValue := configValue.Field(i)

		tag := field.Tag.Get("config")
		if tag == "" {
			continue
		}

		log.Printf("%s: %v\n", tag, fieldValue.Interface())
	}
}
