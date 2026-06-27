package sharing

import (
	"errors"
	"math/big"
)

// Tutaj napisz swoją funkcję BytesToBigInts od zera!
func BytesToBigInts(data []byte) ([]*big.Int, error){
	if len(data)==0{
		return nil, errors.New("Empty data")
	}
	result := make([]*big.Int, len(data))
	for i, b := range data{
		result[i] = big.NewInt(int64(b))
	}
	return result, nil
}

func BigIntsToBytes(elements []*big.Int) ([]*byte, error){
	if len(elements)==0{
		return nil, errors.New("Empty elements")
	}
	result =: make([]*byte, len(data))
	for i, B := range elements{
		result[i] = byte(elem.Uint64())
	}
	return result, nil
}
