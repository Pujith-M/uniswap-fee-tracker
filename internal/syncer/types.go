package syncer

import (
	"database/sql/driver"
	"fmt"
	"math/big"
)

// BigInt represents a custom type for handling big.Int in GORM
type BigInt struct {
	*big.Int
}

func NewBigInt(int *big.Int) *BigInt {
	return &BigInt{Int: int}
}

// Value converts BigInt to database-friendly format
func (b BigInt) Value() (driver.Value, error) {
	if b.Int == nil {
		return nil, nil
	}
	return b.Int.String(), nil
}

// Scan converts database value to BigInt
func (b *BigInt) Scan(value interface{}) error {
	if value == nil {
		b.Int = big.NewInt(0)
		return nil
	}
	switch v := value.(type) {
	case string:
		b.Int = new(big.Int)
		b.Int.SetString(v, 10)
	case []byte:
		b.Int = new(big.Int)
		b.Int.SetString(string(v), 10)
	default:
		return fmt.Errorf("unsupported type: %T", value)
	}
	return nil
}

// BigFloat represents a custom type for handling big.Float in GORM
type BigFloat struct {
	*big.Float
}

func NewBigFloat(float *big.Float) *BigFloat {
	return &BigFloat{Float: float}
}

// Value converts BigFloat to database-friendly format
func (b BigFloat) Value() (driver.Value, error) {
	if b.Float == nil {
		return nil, nil
	}
	return b.Float.Text('f', 18), nil
}

// Scan converts database value to BigFloat
func (b *BigFloat) Scan(value interface{}) error {
	if value == nil {
		b.Float = big.NewFloat(0)
		return nil
	}
	switch v := value.(type) {
	case string:
		b.Float = new(big.Float)
		b.Float.SetString(v)
	case []byte:
		b.Float = new(big.Float)
		b.Float.SetString(string(v))
	default:
		return fmt.Errorf("unsupported type: %T", value)
	}
	return nil
}
