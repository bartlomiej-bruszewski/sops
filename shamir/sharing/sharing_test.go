package sharing

import (
	"bytes"
	"math/big"
	"testing"
)

// TestFieldArithmetic sprawdza poprawność operacji modularnych w ciele Z_997
func TestFieldArithmetic(t *testing.T) {
	Prime = big.NewInt(997)

	// 1. Test dodawania: (500 + 600) mod 997 = 103
	a := big.NewInt(500)
	b := big.NewInt(600)
	expectedAdd := big.NewInt(103)
	if res := FieldAdd(a, b); res.Cmp(expectedAdd) != 0 {
		t.Errorf("FieldAdd(%s, %s) = %s; oczekiwano %s", a, b, res, expectedAdd)
	}

	// 2. Test odejmowania (wynik dodatni): (500 - 200) mod 997 = 300
	c := big.NewInt(500)
	d := big.NewInt(200)
	expectedSubPos := big.NewInt(300)
	if res := FieldSub(c, d); res.Cmp(expectedSubPos) != 0 {
		t.Errorf("FieldSub(%s, %s) = %s; oczekiwano %s", c, d, res, expectedSubPos)
	}

	// 3. Test odejmowania (wynik ujemny): (100 - 200) mod 997 = 897
	e := big.NewInt(100)
	f := big.NewInt(200)
	expectedSubNeg := big.NewInt(897)
	if res := FieldSub(e, f); res.Cmp(expectedSubNeg) != 0 {
		t.Errorf("FieldSub(%s, %s) = %s; oczekiwano %s", e, f, res, expectedSubNeg)
	}

	// 4. Test mnożenia: (10 * 100) mod 997 = 3
	g := big.NewInt(10)
	h := big.NewInt(100)
	expectedMult := big.NewInt(3)
	if res := FieldMult(g, h); res.Cmp(expectedMult) != 0 {
		t.Errorf("FieldMult(%s, %s) = %s; oczekiwano %s", g, h, res, expectedMult)
	}

	// 5. Test dzielenia: (10 / 2) mod 997 = 5
	i := big.NewInt(10)
	j := big.NewInt(2)
	expectedDiv := big.NewInt(5)
	resDiv, err := FieldDiv(i, j)
	if err != nil {
		t.Fatalf("FieldDiv(%s, %s) zwróciło błąd: %v", i, j, err)
	}
	if resDiv.Cmp(expectedDiv) != 0 {
		t.Errorf("FieldDiv(%s, %s) = %s; oczekiwano %s", i, j, resDiv, expectedDiv)
	}

	// 6. Test dzielenia przez zero
	zero := big.NewInt(0)
	_, errZero := FieldDiv(i, zero)
	if errZero == nil {
		t.Error("FieldDiv przez 0 nie zwróciło błędu!")
	}
}

// TestChunker sprawdza poprawność zamiany bajtów na duże liczby i z powrotem
func TestChunker(t *testing.T) {
	originalSecret := []byte("Sopot")

	// 1. Test []byte -> []*big.Int
	bigInts, err := BytesToBigInts(originalSecret)
	if err != nil {
		t.Fatalf("BytesToBigInts zwróciło błąd: %v", err)
	}

	if len(bigInts) != len(originalSecret) {
		t.Errorf("Oczekiwano %d elementów, otrzymano %d", len(originalSecret), len(bigInts))
	}

	if bigInts[0].Cmp(big.NewInt(83)) != 0 {
		t.Errorf("Oczekiwano pierwszej wartości 83 (litera S), otrzymano %s", bigInts[0])
	}

	// 2. Test []*big.Int -> []byte
	reconstructedBytes, err := BigIntsToBytes(bigInts)
	if err != nil {
		t.Fatalf("BigIntsToBytes zwróciło błąd: %v", err)
	}

	if !bytes.Equal(originalSecret, reconstructedBytes) {
		t.Errorf("Oczekiwano %q, otrzymano %q", string(originalSecret), string(reconstructedBytes))
	}
}
