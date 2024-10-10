package pkg

import (
	"fmt"
	"io"
	"reflect"

	"github.com/joho/godotenv"
)

func DotEnvMarshal(s any, w io.Writer) error {
	pval := reflect.ValueOf(s)

	var ptype reflect.Type

	if !pval.IsValid() {
		return fmt.Errorf("cannot marshal invalid value")
	}

	if pval.Kind() == reflect.Ptr {
		ptype = pval.Type().Elem()
	} else {
		ptype = pval.Type()
	}

	envMap := map[string]string{}

	if ptype.Kind() != reflect.Struct {
		return fmt.Errorf("%s is not a struct", ptype)
	}

	for i := range ptype.NumField() {
		f := ptype.Field(i)

		if f.Type.String() != "string" {
			continue
		}

		fieldName := f.Name

		if s := f.Tag.Get("dot"); s != "" {
			fieldName = s
		}

		envMap[fieldName] = reflect.Indirect(pval).FieldByName(f.Name).String()
	}

	marshal, err := godotenv.Marshal(envMap)
	if err != nil {
		return fmt.Errorf("could not marshal dotenv: %w", err)
	}

	_, err = io.WriteString(w, marshal)
	if err != nil {
		return fmt.Errorf("could not write: %w", err)
	}

	return nil
}
