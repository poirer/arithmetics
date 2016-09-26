package main

import (
	"errors"
	"math"
	"math/rand"
	"strconv"
	"strings"
)

type (
	valueGenerator interface {
		genValue(typeStr string) (string, error)
	}

	randValueGenerator struct{}
)

func (randValueGenerator) genValue(typeStr string) (string, error) {
	typeStr = strings.Trim(typeStr, "\n\t\r ")
	if typeStr[0] == '[' && typeStr[len(typeStr)-1] == ']' {
		typeStr = typeStr[1 : len(typeStr)-1]
		return genArray(typeStr), nil
	}
	switch typeStr {
	case typeString:
		return genString(), nil
	case typeInt:
		return genInt(), nil
	case typeBool:
		return genBool(), nil
	case typeFloat:
		return genFloat(), nil
	default:
		return "", errors.New("Unsupported type <" + typeStr + ">")
	}
}

func genString() string {
	var strBuf = buffers.getBuffer()
	defer buffers.returnBuffer(strBuf)
	const charSet = "abcdefghijklmnopqrstuvwxyz "
	var l = rand.Intn(24)
	strBuf.WriteByte('"')
	for i := 0; i < l; i++ {
		strBuf.WriteByte(charSet[rand.Intn(len(charSet))])
	}
	strBuf.WriteByte('"')
	return strBuf.String()
}

func genBool() string {
	return strconv.FormatBool(rand.Intn(2) != 0)
}

func genInt() string {
	return strconv.FormatInt(int64(rand.Intn(math.MaxInt32)), 10)
}

func genFloat() string {
	var value = rand.Float32()
	return strconv.FormatFloat(float64(value), 'f', -1, 32)
}

func genArray(itemType string) string {
	var buf = buffers.getBuffer()
	defer buffers.returnBuffer(buf)
	buf.WriteByte('[')
	var arlen = rand.Intn(10)
	for i := 0; i < arlen; i++ {
		switch itemType {
		case typeString:
			buf.WriteString(genString())
			buf.WriteString(", ")
		case typeInt:
			buf.WriteString(genInt())
			buf.WriteString(", ")
		case typeBool:
			buf.WriteString(genBool())
			buf.WriteString(", ")
		case typeFloat:
			buf.WriteString(genFloat())
			buf.WriteString(", ")
		}
	}
	if arlen > 0 {
		buf.Truncate(buf.Len() - 2)
	}
	buf.WriteByte(']')
	return buf.String()
}
