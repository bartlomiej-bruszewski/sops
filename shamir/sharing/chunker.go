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
	for i, b in range data{
		big.NewInt(int64(b))
		result[i]=big
	}
	return result
}
