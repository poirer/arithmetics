package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var generator valueGenerator

func isNestedObjectFollows(spec string, start int) bool {
	for j := start; j < len(spec); j++ {
		if spec[j] == '{' {
			return true
		} else if spec[j] == ',' {
			return false
		}
	}
	return false
}

func composeRequestBody(spec string) (string, error) {
	var jsonBuf = buffers.getBuffer()
	defer buffers.returnBuffer(jsonBuf)
	for i := 0; i < len(spec); {
		c := spec[i]
		if c != ':' {
			jsonBuf.WriteByte(c)
			i++
		} else {
			if nestedObj := isNestedObjectFollows(spec, i); nestedObj {
				jsonBuf.WriteByte(c)
				i++
			} else {
				var typeBuf = buffers.getBuffer()
				jsonBuf.WriteByte(c)
				for {
					i++
					c = spec[i]
					if c == ',' || c == '}' || c == '\n' {
						v, err := generator.genValue(typeBuf.String())
						if err != nil {
							return "", err
						}
						jsonBuf.WriteString(v)
						break
					} else if c == ' ' || c == '\t' { // Just to keep format; may be not needed
						jsonBuf.WriteByte(c)
					} else {
						typeBuf.WriteByte(c)
					}
				}
				buffers.returnBuffer(typeBuf)
			}
		}
	}
	return jsonBuf.String(), nil
}

func extractNextToken(spec string, start int) (token string, end int) {
	var tokenBuf = buffers.getBuffer()
	defer buffers.returnBuffer(tokenBuf)
	var depth = 0
	if spec[start] == '{' {
		for i := start + 1; i < len(spec); i++ {
			c := spec[i]
			if c != ' ' && c != '\n' && c != '\t' {
				start = i
				break
			}
		}
	}
	for i := start; i < len(spec); i++ {
		var c = spec[i]
		if c == ',' || c == '}' {
			if depth == 0 {
				end = i + 1
				break
			} else {
				if c == '}' {
					depth--
				}
				tokenBuf.WriteByte(c)
			}
		} else if c == '{' {
			depth++
			tokenBuf.WriteByte(c)
		} else {
			tokenBuf.WriteByte(c)
		}
	}
	token = tokenBuf.String()
	return
}

func parseResponseFields(spec string) ([]responseField, error) {
	var result = buffers.getFieldSlice()
	defer buffers.returnFieldSlice(result)
	var n = 0
	for n < len(spec) {
		var token string
		token, n = extractNextToken(spec, n)
		token = strings.Trim(token, "\n\t ")
		if strings.Index(token, "{") < 0 {
			parts := strings.Split(token, ":")
			if len(parts) == 2 {
				var fieldName = strings.Trim(parts[0], " ")
				var fieldType = strings.Trim(parts[1], " ")
				result = append(result, responseField{fieldName, fieldType, nil})
			} else {
				return nil, errors.New("Parse error: simple field definition must look like \"field\":type")
			}
		} else {
			ind := strings.Index(token, ":")
			objInd := strings.Index(token, "{")
			if ind < 1 || ind > objInd {
				return nil, errors.New("Parse error: object field definition must look like \"file\":{...}")
			}
			var fieldName = strings.Trim(token[0:ind], "\n\t ")
			var nestedObj = strings.Trim(token[ind+1:], "\n\t ")
			rf, err := parseResponseFields(nestedObj)
			if err != nil {
				return nil, err
			}
			result = append(result, responseField{fieldName, "object", rf})
		}
	}
	return result, nil
}

func checkResponseAgainstExpectedFields(content []byte, expectedFields []responseField) (bool, error) {
	var data interface{}
	err := json.Unmarshal(content, &data)
	if err != nil {
		return false, err
	}
	var fieldMaps []map[string]interface{}
	switch v := data.(type) {
	case map[string]interface{}:
		fieldMaps = make([]map[string]interface{}, 1, 1)
		fieldMaps[0] = v
	case []map[string]interface{}:
		fieldMaps = v
	case []interface{}:
		fieldMaps = make([]map[string]interface{}, len(v), len(v))
		for i, in := range v {
			fieldMaps[i] = in.(map[string]interface{})
		}
	}
	for _, fm := range fieldMaps {
		for _, rf := range expectedFields {
			val, found := fm[rf.name]
			if !found && rf.name[0] == '"' { // if expected field is quoted, then try to remove quotes
				val, found = fm[rf.name[1:len(rf.name)-1]]
			}
			if !found {
				return false, fmt.Errorf("Expected field %s not found", rf.name)
			}
			switch val.(type) {
			case string:
				if rf.typ != typeString {
					return false, fmt.Errorf("Field %s has wrong type, expected %s", rf.name, typeString)
				}
			case int:
				if rf.typ != typeInt {
					return false, fmt.Errorf("Field %s has wrong type, expected %s", rf.name, typeInt)
				}
			case bool:
				if rf.typ != typeBool {
					return false, fmt.Errorf("Field %s has wrong type, expected %s", rf.name, typeBool)
				}
			case float32:
				if rf.typ != typeFloat {
					return false, fmt.Errorf("Field %s has wrong type, expected %s", rf.name, typeFloat)
				}
			case []string:
				if rf.typ != typeStringArray {
					return false, fmt.Errorf("Field %s has wrong type, expected %s", rf.name, typeStringArray)
				}
			case []int:
				if rf.typ != typeIntArray {
					return false, fmt.Errorf("Field %s has wrong type, expected %s", rf.name, typeIntArray)
				}
			case []bool:
				if rf.typ != typeBoolArray {
					return false, fmt.Errorf("Field %s has wrong type, expected %s", rf.name, typeBoolArray)
				}
			case []float32:
				if rf.typ != typeFloatArray {
					return false, fmt.Errorf("Field %s has wrong type, expected %s", rf.name, typeFloatArray)
				}
			}
		}
	}
	// TODO: Implement algorithm that chacks that JSON has fields of specified type (perhaps, in the psecified order)
	return true, nil
}
