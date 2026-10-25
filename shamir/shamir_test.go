package shamir

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// mustPanicWith checks that f panics with exactly the given message.
func mustPanicWith(t *testing.T, want string, f func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected panic %q, got none", want)
		}
		if got := fmt.Sprint(r); got != want {
			t.Fatalf("expected panic %q, got %q", want, got)
		}
	}()
	f()
}

// naiveEval evaluates sum(coeffs[i] * x^i) in GF(2^8) without Horner's method.
func naiveEval(coeffs []uint8, x uint8) uint8 {
	var out uint8
	pow := uint8(1)
	for _, c := range coeffs {
		out = add(out, mult(c, pow))
		pow = mult(pow, x)
	}
	return out
}

// ---------------------------------------------------------------------------
// Existing tests
// ---------------------------------------------------------------------------

func TestSplit_invalid(t *testing.T) {
	secret := []byte("test")

	if _, err := Split(secret, 0, 0); err == nil {
		t.Fatalf("expect error")
	}

	if _, err := Split(secret, 2, 3); err == nil {
		t.Fatalf("expect error")
	}

	if _, err := Split(secret, 1000, 3); err == nil {
		t.Fatalf("expect error")
	}

	if _, err := Split(secret, 10, 1); err == nil {
		t.Fatalf("expect error")
	}

	if _, err := Split(nil, 3, 2); err == nil {
		t.Fatalf("expect error")
	}
}

func TestSplit(t *testing.T) {
	secret := []byte("test")

	out, err := Split(secret, 5, 3)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if len(out) != 5 {
		t.Fatalf("bad: %v", out)
	}

	for _, share := range out {
		if len(share) != len(secret)+1 {
			t.Fatalf("bad: %v", out)
		}
	}
}

func TestCombine_invalid(t *testing.T) {
	// Not enough parts
	if _, err := Combine(nil); err == nil {
		t.Fatalf("should err")
	}

	// Mismatch in length
	parts := [][]byte{
		[]byte("foo"),
		[]byte("ba"),
	}
	if _, err := Combine(parts); err == nil {
		t.Fatalf("should err")
	}

	//Too short
	parts = [][]byte{
		[]byte("f"),
		[]byte("b"),
	}
	if _, err := Combine(parts); err == nil {
		t.Fatalf("should err")
	}

	parts = [][]byte{
		[]byte("foo"),
		[]byte("foo"),
	}
	if _, err := Combine(parts); err == nil {
		t.Fatalf("should err")
	}
}

func TestCombine(t *testing.T) {
	secret := []byte("test")

	out, err := Split(secret, 5, 3)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// There is 5*4*3 possible choices,
	// we will just brute force try them all
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			if j == i {
				continue
			}
			for k := 0; k < 5; k++ {
				if k == i || k == j {
					continue
				}
				parts := [][]byte{out[i], out[j], out[k]}
				recomb, err := Combine(parts)
				if err != nil {
					t.Fatalf("err: %v", err)
				}

				if !bytes.Equal(recomb, secret) {
					t.Errorf("parts: (i:%d, j:%d, k:%d) %v", i, j, k, parts)
					t.Fatalf("bad: %v %v", recomb, secret)
				}
			}
		}
	}
}

func TestField_MulDivSmoke(t *testing.T) {
	for a := range 256 {
		for b := range 256 {
			if b == 0 {
				if out := mult(uint8(a), uint8(b)); out != 0 {
					t.Fatalf("Bad: %v * %v = %v 0", a, b, out)
				}
			} else {
				if out := div(mult(uint8(a), uint8(b)), uint8(b)); out != uint8(a) {
					t.Fatalf("Bad: (%v * %v) / %v = %v %v", a, b, b, out, a)
				}
			}
		}
	}
}

func TestField_Add(t *testing.T) {
	if out := add(16, 16); out != 0 {
		t.Fatalf("Bad: %v 16", out)
	}

	if out := add(3, 4); out != 7 {
		t.Fatalf("Bad: %v 7", out)
	}
}

func TestField_Mult(t *testing.T) {
	if out := mult(3, 7); out != 9 {
		t.Fatalf("Bad: %v 9", out)
	}

	if out := mult(3, 0); out != 0 {
		t.Fatalf("Bad: %v 0", out)
	}

	if out := mult(0, 3); out != 0 {
		t.Fatalf("Bad: %v 0", out)
	}
}

func TestField_Divide(t *testing.T) {
	if out := div(0, 7); out != 0 {
		t.Fatalf("Bad: %v 0", out)
	}

	if out := div(3, 3); out != 1 {
		t.Fatalf("Bad: %v 1", out)
	}

	if out := div(6, 3); out != 2 {
		t.Fatalf("Bad: %v 2", out)
	}
}

func TestPolynomial_Random(t *testing.T) {
	p, err := makePolynomial(42, 2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if p.coefficients[0] != 42 {
		t.Fatalf("bad: %v", p.coefficients)
	}
}

func TestPolynomial_Eval(t *testing.T) {
	p, err := makePolynomial(42, 1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if out := p.evaluate(0); out != 42 {
		t.Fatalf("bad: %v", out)
	}

	out := p.evaluate(1)
	exp := add(42, mult(1, p.coefficients[1]))
	if out != exp {
		t.Fatalf("bad: %v %v %v", out, exp, p.coefficients)
	}
}

func TestInterpolate_Rand(t *testing.T) {
	for i := 0; i < 256; i++ {
		p, err := makePolynomial(uint8(i), 2)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		xVals := []uint8{1, 2, 3}
		yVals := []uint8{p.evaluate(1), p.evaluate(2), p.evaluate(3)}
		out := interpolatePolynomial(xVals, yVals, 0)
		if out != uint8(i) {
			t.Fatalf("Bad: %v %d", out, i)
		}
	}
}

// ---------------------------------------------------------------------------
// Split validation
// ---------------------------------------------------------------------------

// The check "threshold cannot exceed 255" is unreachable when the checks on
// parts run first (parts<threshold, parts>255). Whatever the exact message,
// every combination with threshold > 255 must be rejected.
func TestSplit_thresholdExceeds255(t *testing.T) {
	secret := []byte("test")

	cases := []struct{ parts, threshold int }{
		{255, 256},
		{256, 256},
		{300, 256},
		{1000, 1000},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%d_%d", c.parts, c.threshold), func(t *testing.T) {
			out, err := Split(secret, c.parts, c.threshold)
			if err == nil {
				t.Fatalf("expected error for parts=%d threshold=%d", c.parts, c.threshold)
			}
			if out != nil {
				t.Fatalf("expected nil output on error, got %v", out)
			}
		})
	}
}

func TestSplit_partsEqualThreshold(t *testing.T) {
	secret := []byte("test")

	out, err := Split(secret, 3, 3)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(out) != 3 {
		t.Fatalf("expected 3 shares, got %d", len(out))
	}

	// All shares are required to reconstruct the secret.
	recomb, err := Combine(out)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !bytes.Equal(recomb, secret) {
		t.Fatalf("bad: %v %v", recomb, secret)
	}
}

func TestSplit_xCoordinateEncoding(t *testing.T) {
	secret := []byte("test")

	// The x coordinate is stored in the last byte of each share and is
	// 1-indexed: share i has x == i+1 (x=0 would reveal the secret).
	out, err := Split(secret, 10, 3)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	seen := make(map[uint8]bool)
	for i, share := range out {
		x := share[len(share)-1]
		if x != uint8(i+1) {
			t.Fatalf("share %d: x coordinate = %d, expected %d", i, x, i+1)
		}
		if seen[x] {
			t.Fatalf("duplicate x coordinate %d in share %d", x, i)
		}
		seen[x] = true
	}

	// With the maximum number of parts every x in 1..255 is used exactly once.
	out, err = Split([]byte{1}, 255, 2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	seen = make(map[uint8]bool)
	for i, share := range out {
		x := share[len(share)-1]
		if x != uint8(i+1) {
			t.Fatalf("share %d: x coordinate = %d, expected %d", i, x, i+1)
		}
		seen[x] = true
	}
	if len(seen) != 255 {
		t.Fatalf("expected 255 unique x coordinates, got %d", len(seen))
	}
	if seen[0] {
		t.Fatalf("x coordinate 0 must not be used")
	}
}

func TestSplit_shareOverheadConstant(t *testing.T) {
	if ShareOverhead != 1 {
		t.Fatalf("ShareOverhead = %d, expected 1", ShareOverhead)
	}

	for _, size := range []int{1, 4, 32, 100} {
		secret := make([]byte, size)
		if _, err := rand.Read(secret); err != nil {
			t.Fatalf("err: %v", err)
		}
		out, err := Split(secret, 5, 3)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		for i, share := range out {
			if len(share) != len(secret)+ShareOverhead {
				t.Fatalf("share %d: len=%d, expected %d", i, len(share), len(secret)+ShareOverhead)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Combine validation
// ---------------------------------------------------------------------------

func TestCombine_onePart(t *testing.T) {
	out, err := Split([]byte("test"), 5, 3)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	_, err = Combine([][]byte{out[0]})
	if err == nil {
		t.Fatalf("should err")
	}
	if !strings.Contains(err.Error(), "less than two parts") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestCombine_insufficientSharesWrongResult(t *testing.T) {
	// 32 random bytes: the chance that a wrong reconstruction accidentally
	// equals the secret is negligible.
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatalf("err: %v", err)
	}

	out, err := Split(secret, 5, 3)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Only 2 shares with threshold 3: no error, but a corrupted secret.
	recomb, err := Combine([][]byte{out[0], out[1]})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(recomb) != len(secret) {
		t.Fatalf("unexpected length: %d", len(recomb))
	}
	if bytes.Equal(recomb, secret) {
		t.Fatalf("fewer than threshold shares must not reveal the secret")
	}
}

// ---------------------------------------------------------------------------
// Round-trip (Split -> Combine)
// ---------------------------------------------------------------------------

func TestSplitCombine_singleByteSecret(t *testing.T) {
	for b := range 256 {
		secret := []byte{byte(b)}

		out, err := Split(secret, 3, 2)
		if err != nil {
			t.Fatalf("byte %d: err: %v", b, err)
		}

		recomb, err := Combine([][]byte{out[0], out[1]})
		if err != nil {
			t.Fatalf("byte %d: err: %v", b, err)
		}
		if !bytes.Equal(recomb, secret) {
			t.Fatalf("byte %d: bad: %v %v", b, recomb, secret)
		}
	}
}

func TestSplitCombine_binarySecretWithNullBytes(t *testing.T) {
	secret := []byte{0x00, 0x00, 0x01, 0x00, 0xff, 0x00, 0x00}

	out, err := Split(secret, 5, 3)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	recomb, err := Combine([][]byte{out[4], out[1], out[2]})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !bytes.Equal(recomb, secret) {
		t.Fatalf("bad: %v %v", recomb, secret)
	}

	// A secret consisting only of null bytes.
	zeros := make([]byte, 16)
	out, err = Split(zeros, 5, 3)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	recomb, err = Combine([][]byte{out[0], out[3], out[4]})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !bytes.Equal(recomb, zeros) {
		t.Fatalf("bad: %v %v", recomb, zeros)
	}
}

func TestSplitCombine_largeSecret(t *testing.T) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatalf("err: %v", err)
	}

	out, err := Split(secret, 5, 3)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Every 3-element subset of the 5 shares must reconstruct the secret.
	for i := 0; i < 5; i++ {
		for j := i + 1; j < 5; j++ {
			for k := j + 1; k < 5; k++ {
				recomb, err := Combine([][]byte{out[i], out[j], out[k]})
				if err != nil {
					t.Fatalf("(%d,%d,%d): err: %v", i, j, k, err)
				}
				if !bytes.Equal(recomb, secret) {
					t.Fatalf("(%d,%d,%d): bad: %v %v", i, j, k, recomb, secret)
				}
			}
		}
	}
}

func TestSplitCombine_maxThreshold(t *testing.T) {
	secret := []byte{0xA7}

	out, err := Split(secret, 255, 255)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(out) != 255 {
		t.Fatalf("expected 255 shares, got %d", len(out))
	}

	recomb, err := Combine(out)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !bytes.Equal(recomb, secret) {
		t.Fatalf("bad: %v %v", recomb, secret)
	}
}

func TestSplitCombine_minThreshold(t *testing.T) {
	secret := []byte("minimum threshold")

	out, err := Split(secret, 5, 2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Any 2 of the 5 shares must reconstruct the secret.
	for i := 0; i < 5; i++ {
		for j := i + 1; j < 5; j++ {
			recomb, err := Combine([][]byte{out[i], out[j]})
			if err != nil {
				t.Fatalf("(%d,%d): err: %v", i, j, err)
			}
			if !bytes.Equal(recomb, secret) {
				t.Fatalf("(%d,%d): bad: %v %v", i, j, recomb, secret)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// GF(2^8) field arithmetic
// ---------------------------------------------------------------------------

func TestField_Add_identity(t *testing.T) {
	for a := range 256 {
		if out := add(uint8(a), 0); out != uint8(a) {
			t.Fatalf("add(%d, 0) = %d", a, out)
		}
		if out := add(0, uint8(a)); out != uint8(a) {
			t.Fatalf("add(0, %d) = %d", a, out)
		}
	}
}

func TestField_Add_selfInverse(t *testing.T) {
	for a := range 256 {
		if out := add(uint8(a), uint8(a)); out != 0 {
			t.Fatalf("add(%d, %d) = %d, expected 0", a, a, out)
		}
	}
}

func TestField_Add_commutativity(t *testing.T) {
	for a := range 256 {
		for b := range 256 {
			if add(uint8(a), uint8(b)) != add(uint8(b), uint8(a)) {
				t.Fatalf("add not commutative for %d, %d", a, b)
			}
		}
	}
}

func TestField_Mult_commutativity(t *testing.T) {
	for a := range 256 {
		for b := range 256 {
			if mult(uint8(a), uint8(b)) != mult(uint8(b), uint8(a)) {
				t.Fatalf("mult not commutative for %d, %d", a, b)
			}
		}
	}
}

func TestField_Mult_identity(t *testing.T) {
	for a := range 256 {
		if out := mult(uint8(a), 1); out != uint8(a) {
			t.Fatalf("mult(%d, 1) = %d", a, out)
		}
		if out := mult(1, uint8(a)); out != uint8(a) {
			t.Fatalf("mult(1, %d) = %d", a, out)
		}
	}
}

func TestField_Mult_distributivity(t *testing.T) {
	// Deterministic sample of triples (a, b, c) covering the whole range.
	for a := 0; a < 256; a += 3 {
		for b := 0; b < 256; b += 5 {
			for c := 0; c < 256; c += 7 {
				x, y, z := uint8(a), uint8(b), uint8(c)
				left := mult(x, add(y, z))
				right := add(mult(x, y), mult(x, z))
				if left != right {
					t.Fatalf("distributivity failed for a=%d b=%d c=%d: %d != %d", a, b, c, left, right)
				}
			}
		}
	}
}

func TestField_Inverse_nonzero(t *testing.T) {
	for a := 1; a < 256; a++ {
		if out := mult(uint8(a), inverse(uint8(a))); out != 1 {
			t.Fatalf("mult(%d, inverse(%d)) = %d, expected 1", a, a, out)
		}
	}
}

func TestField_Inverse_zero(t *testing.T) {
	if out := inverse(0); out != 0 {
		t.Fatalf("inverse(0) = %d, expected 0", out)
	}
}

func TestField_Divide_byZeroPanics(t *testing.T) {
	t.Run("zero_by_zero", func(t *testing.T) {
		mustPanicWith(t, "divide by zero", func() { div(0, 0) })
	})
	t.Run("nonzero_by_zero", func(t *testing.T) {
		mustPanicWith(t, "divide by zero", func() { div(5, 0) })
		mustPanicWith(t, "divide by zero", func() { div(255, 0) })
	})
}

// ---------------------------------------------------------------------------
// Polynomial internals
// ---------------------------------------------------------------------------

func TestMakePolynomial_coefficientsLength(t *testing.T) {
	cases := []struct{ v, d uint8 }{
		{0, 0},
		{42, 1},
		{255, 2},
		{1, 10},
		{128, 100},
		{7, 254},
	}
	for _, c := range cases {
		p, err := makePolynomial(c.v, c.d)
		if err != nil {
			t.Fatalf("v=%d d=%d: err: %v", c.v, c.d, err)
		}
		if len(p.coefficients) != int(c.d)+1 {
			t.Fatalf("v=%d d=%d: len=%d, expected %d", c.v, c.d, len(p.coefficients), int(c.d)+1)
		}
		if p.coefficients[0] != c.v {
			t.Fatalf("v=%d d=%d: intercept=%d", c.v, c.d, p.coefficients[0])
		}
	}
}

func TestMakePolynomial_randomCoefficientsSet(t *testing.T) {
	nonZero := false
	for i := 0; i < 10 && !nonZero; i++ {
		p, err := makePolynomial(42, 8)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		for _, c := range p.coefficients[1:] {
			if c != 0 {
				nonZero = true
				break
			}
		}
	}
	if !nonZero {
		t.Fatalf("random coefficients were all zero in 10 runs")
	}
}

func TestPolynomial_Eval_atOriginAlwaysIntercept(t *testing.T) {
	for _, degree := range []uint8{1, 2, 3, 10, 50, 254} {
		for _, intercept := range []uint8{0, 1, 42, 128, 255} {
			p, err := makePolynomial(intercept, degree)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if out := p.evaluate(0); out != intercept {
				t.Fatalf("degree=%d: evaluate(0)=%d, expected %d", degree, out, intercept)
			}
		}
	}
}

func TestPolynomial_Eval_hornerConsistency(t *testing.T) {
	// Fixed polynomial with known coefficients:
	// 42 + 7x + 200x^2 + 13x^3 + 99x^4
	p := polynomial{coefficients: []uint8{42, 7, 200, 13, 99}}

	for x := range 256 {
		got := p.evaluate(uint8(x))
		want := naiveEval(p.coefficients, uint8(x))
		if got != want {
			t.Fatalf("x=%d: Horner=%d, naive=%d", x, got, want)
		}
	}

	// Hand-computed spot check: at x=1 the value is the XOR of all coefficients.
	if out := p.evaluate(1); out != 42^7^200^13^99 {
		t.Fatalf("evaluate(1)=%d", out)
	}
}

// ---------------------------------------------------------------------------
// Interpolation
// ---------------------------------------------------------------------------

func TestInterpolate_linearPolynomial(t *testing.T) {
	for i := 0; i < 256; i++ {
		p, err := makePolynomial(uint8(i), 1)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		xVals := []uint8{1, 2}
		yVals := []uint8{p.evaluate(1), p.evaluate(2)}
		if out := interpolatePolynomial(xVals, yVals, 0); out != uint8(i) {
			t.Fatalf("Bad: %v %d", out, i)
		}
	}
}

func TestInterpolate_atNonZeroX(t *testing.T) {
	p, err := makePolynomial(99, 2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	xVals := []uint8{1, 2, 3}
	yVals := []uint8{p.evaluate(1), p.evaluate(2), p.evaluate(3)}

	for x := 1; x < 256; x++ {
		got := interpolatePolynomial(xVals, yVals, uint8(x))
		want := p.evaluate(uint8(x))
		if got != want {
			t.Fatalf("x=%d: interpolated=%d, evaluate=%d", x, got, want)
		}
	}
}
