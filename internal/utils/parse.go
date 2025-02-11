package utils

import (
	"fmt"
	"math/big"
	"strconv"
)

// ParseInt64 safely parses an interface{} value to int64
func ParseInt64(value interface{}) (int64, error) {
	switch v := value.(type) {
	case string:
		return strconv.ParseInt(v, 10, 64)
	case float64:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("unexpected type for int64: %T", value)
	}
}

// ParseBigFloat safely parses an interface{} value to *big.Float
func ParseBigFloat(value interface{}) (*big.Float, error) {
	bf := new(big.Float)

	switch v := value.(type) {
	case string:
		_, _, err := bf.Parse(v, 10)
		if err != nil {
			return nil, err
		}
		return bf, nil
	case float64:
		return bf.SetFloat64(v), nil
	default:
		return nil, fmt.Errorf("unexpected type for big.Float: %T", value)
	}
}

func MustParseInt64(v interface{}) int64 {
	result, _ := ParseInt64(v)
	return result
}

func MustParseBigFloat(v interface{}) *big.Float {
	result, _ := ParseBigFloat(v)
	return result
}
