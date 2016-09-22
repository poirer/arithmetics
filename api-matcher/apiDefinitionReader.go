package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
)

func readAPICallDefinitions(path string) ([]apiCallDef, error) {
	defFile, err := os.Open(path)
	defer defFile.Close()
	if err != nil {
		return nil, err
	}
	return readDefinitionsFromReader(defFile)
}

func readDefinitionsFromReader(source io.Reader) ([]apiCallDef, error) {
	xmlContent, err := ioutil.ReadAll(source)
	if err != nil {
		return nil, err
	}
	var defList apiDefList
	err = xml.Unmarshal(xmlContent, &defList)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(defList.DefList); i++ {
		trimFields(&defList.DefList[i])
	}
	return defList.DefList, nil
}

func castToType(variable interface{}) interface{} {
	switch t := variable.(type) {
	case *apiCallDef:
		return t
	case *reqAddr:
		return t
	case *formParam:
		return t
	case reqAddr:
		return &t
	case formParam:
		return &t
	default:
		fmt.Printf("Unexpected type %T\n", variable)
		return t
	}
}

// The idea was to write function that can trim all strings in any struct
// But due to we need cast types, the original idea was not implemented
// Now only apiCallDef can be processed correctly
func trimFields(dataToTrim interface{}) {
	typedData := castToType(dataToTrim)
	reflectData := reflect.ValueOf(typedData).Elem()
	for i := 0; i < reflectData.NumField(); i++ {
		field := reflectData.Field(i)
		if field.Type().Kind() == reflect.Ptr {
			innerStruct := castToType(field.Interface())
			trimFields(innerStruct)
		} else if field.Type().Kind() == reflect.Slice {
			for j := 0; j < field.Len(); j++ {
				fieldAtIndex := field.Index(j)
				if fieldAtIndex.Type().Kind() == reflect.Struct {
					innerElement := castToType(fieldAtIndex.Interface())
					trimFields(innerElement)
				} else if fieldAtIndex.Type().Kind() == reflect.String {
					var trimmedString = strings.Trim(fieldAtIndex.String(), "\n\t ")
					fieldAtIndex.SetString(trimmedString)
				}
			}
		} else if field.Type().Kind() == reflect.String {
			if field.CanSet() {
				var trimmedString = strings.Trim(field.String(), "\n\t ")
				field.SetString(trimmedString)
			}
		}
	}
}
