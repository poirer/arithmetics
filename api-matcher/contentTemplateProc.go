package main

import (
	"bytes"
	"errors"
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
	var jsonBuf bytes.Buffer
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
				var typeBuf bytes.Buffer
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
			}
		}
	}
	return jsonBuf.String(), nil
}

func extractNextToken(spec string, start int) (token string, end int) {
	var depth = 0
	var tokenBuf bytes.Buffer
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
	var result []responseField
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

func checkResponseAgainstExpectedFields(content string, expectedFields []responseField) (bool, error) {
	// TODO: Implement algorithm that chacks that JSON has fields of specified type (perhaps, in the psecified order)
	return true, nil
}
