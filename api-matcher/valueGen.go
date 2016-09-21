package main

import (
	"bytes"
	"errors"
	"strings"
)

func genValue(typeStr string) (string, error) {
	typeStr = strings.Trim(typeStr, "\n\t\r ")
	if typeStr[0] == '[' && typeStr[len(typeStr)-1] == ']' {
		typeStr = typeStr[1 : len(typeStr)-1]
		return genArray(typeStr), nil
	}
	switch typeStr {
	case "string":
		return genString(), nil
	case "int":
		return genInt(), nil
	case "bool":
		return genBool(), nil
	default:
		return "", errors.New("Unsupported type <" + typeStr + ">")
	}
}

func genString() string {
	return `"A"`
}

func genBool() string {
	return "true"
}

func genInt() string {
	return "0"
}

func genArray(itemType string) string {
	var buf bytes.Buffer
	buf.WriteByte('[')
	for i := 0; i < 3; i++ {
		switch itemType {
		case "string":
			buf.WriteString(genString())
			buf.WriteString(", ")
		case "int":
			buf.WriteString(genInt())
			buf.WriteString(", ")
		case "bool":
			buf.WriteString(genBool())
			buf.WriteString(", ")
		}
	}
	buf.Truncate(buf.Len() - 2)
	buf.WriteByte(']')
	return buf.String()
}
