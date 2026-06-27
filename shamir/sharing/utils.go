package sharing

import (
	"math/big"
	"errors"
)

// Nasz moduł p (liczba pierwsza) - dla ułatwienia zdefiniujmy globalną zmienną.
// Możesz użyć małej liczby pierwszej do testów, np. 997, albo dużej.
var Prime = big.NewInt(997) 

// FieldAdd powinno zwrócić (a + b) mod Prime
func FieldAdd(a, b *big.Int) *big.Int {
	res := new(big.Int).Add(a,b)
	return res.Mod(res, Prime)
}

// FieldSub powinno zwrócić (a - b) mod Prime
// Pamiętaj, że jeśli wynik odejmowania jest ujemny, musisz dodać Prime!
func FieldSub(a, b *big.Int) *big.Int {
	res := new(big.Int).Sub(a,b)
	if res.Sign()<0{
		res.Add(res, Prime)
	}
	return res.Mod(res, Prime)
	// TWOJE ZADANIE
}

Spróbuj napisać te dwie funkcje. Nie przejmuj się, jeśli coś pójdzie nie tak z obsługą wskaźników w `math/big` – skorygujemy to i wyjaśnię Ci, jak Go zarządza tą pamięcią!

// FieldMult powinno zwrócić (a * b) mod Prime.
func FieldMult(a, b *big.Int) *big.Int {
	res := new(big.Int).Mul(a,b)
	return res.Mod(res, Prime)
}

// FieldDiv powinno zwrócić (a / b) mod Prime, czyli (a * b^-1) mod Prime.
// Jeśli b jest równe 0, funkcja powinna zwrócić błąd: errors.New("division by zero")
func FieldDiv(a, b *big.Int) (*big.Int, error) {
	inv := new(big.Int).ModInverse(b, Prime)
	if inv == nil{
		errors.New("division by zero")
	}
	res := new(big.Int).Mul(a,b)
	return res.Mod(res, Prime), nil
}
