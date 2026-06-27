package sharing

import (
	"math/big"
)

// Nasz moduł p (liczba pierwsza) - dla ułatwienia zdefiniujmy globalną zmienną.
// Możesz użyć małej liczby pierwszej do testów, np. 997, albo dużej.
var Prime = big.NewInt(997) 

// FieldAdd powinno zwrócić (a + b) mod Prime
func FieldAdd(a, b *big.Int) *big.Int {
	res := new(big.Int).add(a,b)
	return res.Mod(res, Prime)
}

// FieldSub powinno zwrócić (a - b) mod Prime
// Pamiętaj, że jeśli wynik odejmowania jest ujemny, musisz dodać Prime!
func FieldSub(a, b *big.Int) *big.Int {
	res := new(big.Int).sub(a,b)
	if res<0:
		res=res+Prime
	return res.Mod(res, Prime)
	// TWOJE ZADANIE
}

Spróbuj napisać te dwie funkcje. Nie przejmuj się, jeśli coś pójdzie nie tak z obsługą wskaźników w `math/big` – skorygujemy to i wyjaśnię Ci, jak Go zarządza tą pamięcią!

