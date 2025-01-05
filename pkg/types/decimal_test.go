package types

import (
	"encoding/xml"
	"log/slog"
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test1(t *testing.T) {
	str := "1234.56"
	d, err := NewDecimalFromString(str)
	if err != nil {
		t.Errorf("Parse error. %v", err)
		return
	}
	if !d.Equal(NewDecimal(123456, -2)) {
		t.Errorf("Invalid parsing. Original: %v, result: %v", str, d)
	}
}

func Test2(t *testing.T) {
	str := "-1234.56"
	d, err := NewDecimalFromString(str)
	if err != nil {
		t.Errorf("Parse error. %v", err)
		return
	}
	if !d.Equal(NewDecimal(-123456, -2)) {
		t.Errorf("Invalid parsing. Original: %v, result: %v", str, d)
	}
}

func Test3(t *testing.T) {
	str := "+1234.56"
	d, err := NewDecimalFromString(str)
	if err != nil {
		t.Errorf("Parse error. %v", err)
		return
	}
	if !d.Equal(NewDecimal(123456, -2)) {
		t.Errorf("Invalid parsing. Original: %v, result: %v", str, d)
	}
}

func TestFromMoney(t *testing.T) {
	//-1.75 is represented as `units`=-1 and `nanos`=-750,000,000.
	var d Decimal

	d = NewDecimalFromMoneyParts(0, 0)
	if !d.Equal(Zero) {
		t.Errorf("Not zero")
	}

	d = NewDecimalFromMoneyParts(-1, -750_000_000)
	if !d.Equal(NewDecimal(-175, -2)) {
		t.Errorf("Not equal -1.75, Value: %v", d)
	}

	d = NewDecimalFromMoneyParts(0, -750_000_000)
	if !d.Equal(NewDecimal(-75, -2)) {
		t.Errorf("Not equal -0.75, Value: %v", d)
	}

	d = NewDecimalFromMoneyParts(0, 68_000_000)
	if !d.Equal(NewDecimal(68, -3)) {
		t.Errorf("Not equal 0.068, Value: %v", d)
	}

	d = NewDecimalFromMoneyParts(10, 68_000_000)
	if !d.Equal(NewDecimal(10068, -3)) {
		t.Errorf("Not equal 10.068, Value: %v", d)
	}

	d = NewDecimalFromMoneyParts(97, 0)
	if !d.Equal(NewDecimalFromInt(97)) {
		t.Errorf("Not equal 97, Value: %v", d)
	}

	d = NewDecimalFromMoneyParts(-13, 0)
	if !d.Equal(NewDecimalFromInt(-13)) {
		t.Errorf("Not equal -13, Value: %v", d)
	}
}

func TestNormalize(t *testing.T) {
	var d Decimal
	v := big.Int{}
	d = NewDecimal(1000, 0)
	d.Normalize()
	if d.exp != 3 || d.value.Cmp(big.NewInt(1)) != 0 {
		t.Errorf("Invalid nornalize, Value: %v", d)
	}

	d = NewDecimal(10002, 0)
	d.Normalize()
	if d.exp != 0 || d.value.Cmp(big.NewInt(10002)) != 0 {
		t.Errorf("Invalid nornalize, Value: %v", d)
	}

	d = NewDecimal(-1000, 0)
	d.Normalize()
	if d.exp != 3 || d.value.Cmp(big.NewInt(-1)) != 0 {
		t.Errorf("Invalid nornalize, Value: %v", d)
	}

	d, _ = NewDecimalFromString("229223372036854775807000")
	d.Normalize()
	v.SetString("229223372036854775807", 10)
	if d.exp != 3 || d.value.Cmp(&v) != 0 {
		t.Errorf("Invalid nornalize, Value: %v", d)
	}

	d, _ = NewDecimalFromString("-229223372036854775807000")
	d.Normalize()
	v.SetString("-229223372036854775807", 10)
	if d.exp != 3 || d.value.Cmp(&v) != 0 {
		t.Errorf("Invalid nornalize, Value: %v", d)
	}

	d, _ = NewDecimalFromString("-2292233720368547758070001")
	d.Normalize()
	v.SetString("-2292233720368547758070001", 10)
	if d.exp != 0 || d.value.Cmp(&v) != 0 {
		t.Errorf("Invalid nornalize, Value: %v", d)
	}
}

func TestMoneyParts(t *testing.T) {
	var d Decimal

	var err error

	d = NewDecimal(175, -2)
	u, n := d.ToMoneyParts()

	require.Equal(t, int64(1), u)
	require.Equal(t, int32(750_000_000), n)

	d = NewDecimal(175, 2)
	u, n = d.ToMoneyParts()

	require.Equal(t, int64(17500), u)
	require.Equal(t, int32(0), n)

	d = NewDecimal(15643376614, -11)

	u, n = d.ToMoneyParts()
	require.Equal(t, int64(0), u)
	require.Equal(t, int32(156433766), n)

	d = NewDecimal(9223372036854775807, -2)
	u, n = d.ToMoneyParts()
	require.Equal(t, int64(92233720368547758), u)
	require.Equal(t, int32(70000000), n)

	d, err = NewDecimalFromString("9223372036854775807.12")

	require.NoError(t, err)

	u, n = d.ToMoneyParts()
	require.Equal(t, int64(9223372036854775807), u)
	require.Equal(t, int32(120000000), n)

	d, err = NewDecimalFromString("-9223372036854775807.12")
	require.NoError(t, err)

	if err != nil {
		t.Errorf("Parse decimal error: %v", err)
	}

	u, n = d.ToMoneyParts()
	require.Equal(t, int64(-9223372036854775807), u)
	require.Equal(t, int32(-120000000), n)
}

func TestDecimal_Abs(t *testing.T) {
	d, _ := NewDecimalFromString("-12.3")

	d1 := NewDecimal(123, -1)

	if !d.Abs().Equal(d1) {
		t.Fatalf("Invalid abs\n")
	}
}

func TestDecimal_Truncate(t *testing.T) {
	d := NewDecimal(12378, -2)

	d = d.Truncate(1)

	assert.True(t, d.Equal(NewDecimal(1237, -1)))

	d = NewDecimal(12378, -2)
	d = d.Truncate(0)

	assert.True(t, d.Equal(NewDecimal(123, 0)))
}

func TestDecimal_DivRound(t *testing.T) {
	cases := []struct {
		d      string
		d2     string
		prec   int32
		result string
	}{
		{"2", "2", 0, "1"},
		{"1", "2", 0, "1"},
		{"-1", "2", 0, "-1"},
		{"-1", "-2", 0, "1"},
		{"1", "-2", 0, "-1"},
		{"1", "-20", 1, "-0.1"},
		{"1", "-20", 2, "-0.05"},
		{"1", "20.0000000000000000001", 1, "0"},
		{"1", "19.9999999999999999999", 1, "0.1"},
	}
	for _, s := range cases {
		d, _ := NewDecimalFromString(s.d)
		d2, _ := NewDecimalFromString(s.d2)
		result, _ := NewDecimalFromString(s.result)
		prec := s.prec
		q := d.DivRound(d2, prec)
		if sign(q)*sign(d)*sign(d2) < 0 {
			t.Errorf("sign of quotient wrong, got: %v/%v is about %v", d, d2, q)
		}
		x := q.Mul(d2).Abs().Sub(d.Abs()).Mul(NewDecimal(2, 0))
		if x.Cmp(d2.Abs().Mul(NewDecimal(1, -prec))) > 0 {
			t.Errorf("wrong rounding, got: %v/%v prec=%d is about %v", d, d2, prec, q)
		}
		if x.Cmp(d2.Abs().Mul(NewDecimal(-1, -prec))) <= 0 {
			t.Errorf("wrong rounding, got: %v/%v prec=%d is about %v", d, d2, prec, q)
		}
		if !q.Equal(result) {
			t.Errorf("rounded division wrong %s / %s scale %d = %s, got %v", s.d, s.d2, prec, s.result, q)
		}
	}
}

func sign(d Decimal) int {
	return d.value.Sign()
}

func TestDecimal_RoundUp(t *testing.T) {
	d := NewDecimal(1521, -3)

	assert.Equal(t, d.RoundUp(2), NewDecimal(153, -2))

	d = NewDecimal(-1521, -3)

	assert.Equal(t, d.RoundUp(2), NewDecimal(-153, -2))

	d = NewDecimal(-152, -2)

	assert.Equal(t, d.RoundUp(2), NewDecimal(-152, -2))

	d = NewDecimal(15, -1)

	assert.Equal(t, d.RoundUp(2).Cmp(NewDecimal(15, -1)), 0)

	d = Zero

	assert.Equal(t, d.RoundUp(2).Cmp(Zero), 0)

	d = NewDecimal(1526, -3)

	assert.Equal(t, d.RoundUp(2), NewDecimal(153, -2))
}

func TestDecimalMax(t *testing.T) {
	data := []struct {
		A        Decimal
		B        Decimal
		Expected Decimal
	}{
		{
			A:        NewDecimalFromFloat(0.1),
			B:        NewDecimalFromFloat(0.2),
			Expected: NewDecimalFromFloat(0.2),
		},
		{
			A:        NewDecimalFromFloat(0.2),
			B:        NewDecimalFromFloat(0.1),
			Expected: NewDecimalFromFloat(0.2),
		},
		{
			A:        NewDecimalFromFloat(0.2),
			Expected: NewDecimalFromFloat(0.2),
		},
		{
			B:        NewDecimalFromFloat(0.2),
			Expected: NewDecimalFromFloat(0.2),
		},
		{},
	}

	for _, d := range data {
		if m := DecimalMax(d.A, d.B); !m.Equal(d.Expected) {
			t.Fatalf("decimal max %v %v expected %v got %v", d.A, d.B, d.Expected, m)
		}
	}
}

func TestDecimalMin(t *testing.T) {
	data := []struct {
		A        Decimal
		B        Decimal
		Expected Decimal
	}{
		{
			A:        NewDecimalFromFloat(0.1),
			B:        NewDecimalFromFloat(0.2),
			Expected: NewDecimalFromFloat(0.1),
		},
		{
			A:        NewDecimalFromFloat(0.2),
			B:        NewDecimalFromFloat(0.1),
			Expected: NewDecimalFromFloat(0.1),
		},
		{
			A: NewDecimalFromFloat(0.2),
		},
		{
			B: NewDecimalFromFloat(0.2),
		},
		{},
	}

	for _, d := range data {
		if m := DecimalMin(d.A, d.B); !m.Equal(d.Expected) {
			t.Fatalf("decimal min %v %v expected %v got %v", d.A, d.B, d.Expected, m)
		}
	}
}

func TestMarshalXML(t *testing.T) {
	d := NewDecimal(1235, -2)

	arr, err := xml.Marshal(d)

	assert.NoError(t, err)
	assert.Equal(t, "<Decimal>12.35</Decimal>", string(arr))
}

func TestUnmarshalXML(t *testing.T) {
	var d Decimal

	err := xml.Unmarshal([]byte("<a>45.71</a>"), &d)

	assert.NoError(t, err)
	assert.True(t, NewDecimal(4571, -2).Cmp(d) == 0)
}

func TestDecimal_string(t *testing.T) {
	d := NewDecimal(14540, -4)
	d = d.RoundBank(4)
	assert.Equal(t, "1.4540", d.string(false))
}

func TestDecimal_RoundBank(t *testing.T) {
	type testData struct {
		input         string
		places        int32
		expected      string
		expectedFixed string
	}
	tests := []testData{
		{"1.454", 0, "1", ""},
		{"1.454", 1, "1.5", ""},
		{"1.454", 2, "1.45", ""},
		{"1.454", 3, "1.454", ""},
		{"1.454", 4, "1.454", "1.4540"},
		{"1.454", 5, "1.454", "1.45400"},
		{"1.554", 0, "2", ""},
		{"1.554", 1, "1.6", ""},
		{"1.554", 2, "1.55", ""},
		{"0.554", 0, "1", ""},
		{"0.454", 0, "0", ""},
		{"0.454", 5, "0.454", "0.45400"},
		{"0", 0, "0", ""},
		{"0", 1, "0", "0.0"},
		{"0", 2, "0", "0.00"},
		{"0", -1, "0", ""},
		{"5", 2, "5", "5.00"},
		{"5", 1, "5", "5.0"},
		{"5", 0, "5", ""},
		{"500", 2, "500", "500.00"},
		{"545", -2, "500", ""},
		{"545", -3, "1000", ""},
		{"545", -4, "0", ""},
		{"499", -3, "0", ""},
		{"499", -4, "0", ""},
		{"1.45", 1, "1.4", ""},
		{"1.55", 1, "1.6", ""},
		{"1.65", 1, "1.6", ""},
		{"545", -1, "540", ""},
		{"565", -1, "560", ""},
		{"555", -1, "560", ""},
	}

	// add negative number tests
	for _, test := range tests {
		expected := test.expected
		if expected != "0" {
			expected = "-" + expected
		}
		expectedStr := test.expectedFixed
		if strings.ContainsAny(expectedStr, "123456789") && expectedStr != "" {
			expectedStr = "-" + expectedStr
		}
		tests = append(tests,
			testData{"-" + test.input, test.places, expected, expectedStr})
	}

	for _, test := range tests {
		d, err := NewDecimalFromString(test.input)
		if err != nil {
			panic(err)
		}

		// test Round
		expected, err := NewDecimalFromString(test.expected)
		if err != nil {
			panic(err)
		}
		got := d.RoundBank(test.places)
		if !got.Equal(expected) {
			t.Errorf("Bank Rounding %s to %d places, got %s, expected %s",
				d, test.places, got, expected)
		}

		// test StringFixed
		if test.expectedFixed == "" {
			test.expectedFixed = test.expected
		}
		gotStr := d.StringFixedBank(test.places)
		if gotStr != test.expectedFixed {
			t.Errorf("(%s).StringFixed(%d): got %s, expected %s",
				d, test.places, gotStr, test.expectedFixed)
		}
	}
}

func TestDecimal_Floor(t *testing.T) {
	assertFloor := func(input, expected Decimal) {
		got := input.Floor()
		if !got.Equal(expected) {
			t.Errorf("Floor(%s): got %s, expected %s", input, got, expected)
		}
	}
	type testDataString struct {
		input    string
		expected string
	}
	testsWithStrings := []testDataString{
		{"1.999", "1"},
		{"1", "1"},
		{"1.01", "1"},
		{"0", "0"},
		{"0.9", "0"},
		{"0.1", "0"},
		{"-0.9", "-1"},
		{"-0.1", "-1"},
		{"-1.00", "-1"},
		{"-1.01", "-2"},
		{"-1.999", "-2"},
	}
	for _, test := range testsWithStrings {
		expected, _ := NewDecimalFromString(test.expected)
		input, _ := NewDecimalFromString(test.input)
		assertFloor(input, expected)
	}

	type testDataDecimal struct {
		input    Decimal
		expected string
	}
	testsWithDecimals := []testDataDecimal{
		{NewDecimal(100, -1), "10"},
		{NewDecimal(10, 0), "10"},
		{NewDecimal(1, 1), "10"},
		{NewDecimal(1999, -3), "1"},
		{NewDecimal(101, -2), "1"},
		{NewDecimal(1, 0), "1"},
		{NewDecimal(0, 0), "0"},
		{NewDecimal(9, -1), "0"},
		{NewDecimal(1, -1), "0"},
		{NewDecimal(-1, -1), "-1"},
		{NewDecimal(-9, -1), "-1"},
		{NewDecimal(-1, 0), "-1"},
		{NewDecimal(-101, -2), "-2"},
		{NewDecimal(-1999, -3), "-2"},
	}
	for _, test := range testsWithDecimals {
		expected, _ := NewDecimalFromString(test.expected)
		assertFloor(test.input, expected)
	}
}

func TestDecimal_rescale(t *testing.T) {
	type Inp struct {
		int     int64
		exp     int32
		rescale int32
	}
	tests := map[Inp]string{
		Inp{1234, -3, -5}: "1.234",
		Inp{1234, -3, 0}:  "1",
		Inp{1234, 3, 0}:   "1234000",
		Inp{1234, -4, -4}: "0.1234",
	}

	// add negatives
	for p, s := range tests {
		if p.int > 0 {
			tests[Inp{-p.int, p.exp, p.rescale}] = "-" + s
		}
	}

	for input, s := range tests {
		d := NewDecimal(input.int, input.exp).rescale(input.rescale)

		if d.String() != s {
			t.Errorf("expected %s, got %s (%s, %d)",
				s, d.String(),
				d.value.String(), d.exp)
		}

		// test StringScaled
		s2 := NewDecimal(input.int, input.exp).StringScaled(input.rescale)
		if s2 != s {
			t.Errorf("expected %s, got %s", s, s2)
		}
	}
}

func TestDecimal_Ceil(t *testing.T) {
	assertCeil := func(input, expected Decimal) {
		got := input.Ceil()
		if !got.Equal(expected) {
			t.Errorf("Ceil(%s): got %s, expected %s", input, got, expected)
		}
	}
	type testDataString struct {
		input    string
		expected string
	}
	testsWithStrings := []testDataString{
		{"1.999", "2"},
		{"1", "1"},
		{"1.01", "2"},
		{"0", "0"},
		{"0.9", "1"},
		{"0.1", "1"},
		{"-0.9", "0"},
		{"-0.1", "0"},
		{"-1.00", "-1"},
		{"-1.01", "-1"},
		{"-1.999", "-1"},
	}
	for _, test := range testsWithStrings {
		expected, _ := NewDecimalFromString(test.expected)
		input, _ := NewDecimalFromString(test.input)
		assertCeil(input, expected)
	}

	type testDataDecimal struct {
		input    Decimal
		expected string
	}
	testsWithDecimals := []testDataDecimal{
		{NewDecimal(100, -1), "10"},
		{NewDecimal(10, 0), "10"},
		{NewDecimal(1, 1), "10"},
		{NewDecimal(1999, -3), "2"},
		{NewDecimal(101, -2), "2"},
		{NewDecimal(1, 0), "1"},
		{NewDecimal(0, 0), "0"},
		{NewDecimal(9, -1), "1"},
		{NewDecimal(1, -1), "1"},
		{NewDecimal(-1, -1), "0"},
		{NewDecimal(-9, -1), "0"},
		{NewDecimal(-1, 0), "-1"},
		{NewDecimal(-101, -2), "-1"},
		{NewDecimal(-1999, -3), "-1"},
	}
	for _, test := range testsWithDecimals {
		expected, _ := NewDecimalFromString(test.expected)
		assertCeil(test.input, expected)
	}
}

func TestDecimal_RoundAndStringFixed(t *testing.T) {
	type testData struct {
		input         string
		places        int32
		expected      string
		expectedFixed string
	}
	tests := []testData{
		{"1.454", 0, "1", ""},
		{"1.454", 1, "1.5", ""},
		{"1.454", 2, "1.45", ""},
		{"1.454", 3, "1.454", ""},
		{"1.454", 4, "1.454", "1.4540"},
		{"1.454", 5, "1.454", "1.45400"},
		{"1.554", 0, "2", ""},
		{"1.554", 1, "1.6", ""},
		{"1.554", 2, "1.55", ""},
		{"0.554", 0, "1", ""},
		{"0.454", 0, "0", ""},
		{"0.454", 5, "0.454", "0.45400"},
		{"0", 0, "0", ""},
		{"0", 1, "0", "0.0"},
		{"0", 2, "0", "0.00"},
		{"0", -1, "0", ""},
		{"5", 2, "5", "5.00"},
		{"5", 1, "5", "5.0"},
		{"5", 0, "5", ""},
		{"500", 2, "500", "500.00"},
		{"545", -1, "550", ""},
		{"545", -2, "500", ""},
		{"545", -3, "1000", ""},
		{"545", -4, "0", ""},
		{"499", -3, "0", ""},
		{"499", -4, "0", ""},
	}

	// add negative number tests
	for _, test := range tests {
		expected := test.expected
		if expected != "0" {
			expected = "-" + expected
		}
		expectedStr := test.expectedFixed
		if strings.ContainsAny(expectedStr, "123456789") && expectedStr != "" {
			expectedStr = "-" + expectedStr
		}
		tests = append(tests,
			testData{"-" + test.input, test.places, expected, expectedStr})
	}

	for _, test := range tests {
		d, err := NewDecimalFromString(test.input)
		if err != nil {
			t.Fatal(err)
		}

		// test Round
		expected, err := NewDecimalFromString(test.expected)
		if err != nil {
			t.Fatal(err)
		}

		got := d.Round(test.places)
		if !got.Equal(expected) {
			t.Errorf("Rounding %s to %d places, got %s, expected %s",
				d, test.places, got, expected)
		}

		// test StringFixed
		if test.expectedFixed == "" {
			test.expectedFixed = test.expected
		}

		gotStr := d.StringFixed(test.places)
		if gotStr != test.expectedFixed {
			t.Errorf("(%s).StringFixed(%d): got %s, expected %s",
				d, test.places, gotStr, test.expectedFixed)
		}
	}
}

func TestDecimal_RoundCeilAndStringFixed(t *testing.T) {
	type testData struct {
		input         string
		places        int32
		expected      string
		expectedFixed string
	}
	tests := []testData{
		{"1.454", 0, "2", ""},
		{"1.454", 1, "1.5", ""},
		{"1.454", 2, "1.46", ""},
		{"1.454", 3, "1.454", ""},
		{"1.454", 4, "1.454", "1.4540"},
		{"1.454", 5, "1.454", "1.45400"},
		{"1.554", 0, "2", ""},
		{"1.554", 1, "1.6", ""},
		{"1.554", 2, "1.56", ""},
		{"0.554", 0, "1", ""},
		{"0.454", 0, "1", ""},
		{"0.454", 5, "0.454", "0.45400"},
		{"0", 0, "0", ""},
		{"0", 1, "0", "0.0"},
		{"0", 2, "0", "0.00"},
		{"0", -1, "0", ""},
		{"5", 2, "5", "5.00"},
		{"5", 1, "5", "5.0"},
		{"5", 0, "5", ""},
		{"500", 2, "500", "500.00"},
		{"500", -2, "500", ""},
		{"545", -1, "550", ""},
		{"545", -2, "600", ""},
		{"545", -3, "1000", ""},
		{"545", -4, "10000", ""},
		{"499", -3, "1000", ""},
		{"499", -4, "10000", ""},
		{"1.1001", 2, "1.11", ""},
		{"-1.1001", 2, "-1.10", ""},
		{"-1.454", 0, "-1", ""},
		{"-1.454", 1, "-1.4", ""},
		{"-1.454", 2, "-1.45", ""},
		{"-1.454", 3, "-1.454", ""},
		{"-1.454", 4, "-1.454", "-1.4540"},
		{"-1.454", 5, "-1.454", "-1.45400"},
		{"-1.554", 0, "-1", ""},
		{"-1.554", 1, "-1.5", ""},
		{"-1.554", 2, "-1.55", ""},
		{"-0.554", 0, "0", ""},
		{"-0.454", 0, "0", ""},
		{"-0.454", 5, "-0.454", "-0.45400"},
		{"-5", 2, "-5", "-5.00"},
		{"-5", 1, "-5", "-5.0"},
		{"-5", 0, "-5", ""},
		{"-500", 2, "-500", "-500.00"},
		{"-500", -2, "-500", ""},
		{"-545", -1, "-540", ""},
		{"-545", -2, "-500", ""},
		{"-545", -3, "0", ""},
		{"-545", -4, "0", ""},
		{"-499", -3, "0", ""},
		{"-499", -4, "0", ""},
	}

	for _, test := range tests {
		d, err := NewDecimalFromString(test.input)
		if err != nil {
			t.Fatal(err)
		}

		// test Round
		expected, err := NewDecimalFromString(test.expected)
		if err != nil {
			t.Fatal(err)
		}
		got := d.RoundCeil(test.places)
		if !got.Equal(expected) {
			t.Errorf("Rounding ceil %s to %d places, got %s, expected %s",
				d, test.places, got, expected)
		}

		// test StringFixed
		if test.expectedFixed == "" {
			test.expectedFixed = test.expected
		}
		gotStr := got.StringFixed(test.places)
		if gotStr != test.expectedFixed {
			t.Errorf("(%s).StringFixed(%d): got %s, expected %s",
				d, test.places, gotStr, test.expectedFixed)
		}
	}
}

func TestDecimal_RoundFloorAndStringFixed(t *testing.T) {
	type testData struct {
		input         string
		places        int32
		expected      string
		expectedFixed string
	}
	tests := []testData{
		{"1.454", 0, "1", ""},
		{"1.454", 1, "1.4", ""},
		{"1.454", 2, "1.45", ""},
		{"1.454", 3, "1.454", ""},
		{"1.454", 4, "1.454", "1.4540"},
		{"1.454", 5, "1.454", "1.45400"},
		{"1.554", 0, "1", ""},
		{"1.554", 1, "1.5", ""},
		{"1.554", 2, "1.55", ""},
		{"0.554", 0, "0", ""},
		{"0.454", 0, "0", ""},
		{"0.454", 5, "0.454", "0.45400"},
		{"0", 0, "0", ""},
		{"0", 1, "0", "0.0"},
		{"0", 2, "0", "0.00"},
		{"0", -1, "0", ""},
		{"5", 2, "5", "5.00"},
		{"5", 1, "5", "5.0"},
		{"5", 0, "5", ""},
		{"500", 2, "500", "500.00"},
		{"500", -2, "500", ""},
		{"545", -1, "540", ""},
		{"545", -2, "500", ""},
		{"545", -3, "0", ""},
		{"545", -4, "0", ""},
		{"499", -3, "0", ""},
		{"499", -4, "0", ""},
		{"1.1001", 2, "1.10", ""},
		{"-1.1001", 2, "-1.11", ""},
		{"-1.454", 0, "-2", ""},
		{"-1.454", 1, "-1.5", ""},
		{"-1.454", 2, "-1.46", ""},
		{"-1.454", 3, "-1.454", ""},
		{"-1.454", 4, "-1.454", "-1.4540"},
		{"-1.454", 5, "-1.454", "-1.45400"},
		{"-1.554", 0, "-2", ""},
		{"-1.554", 1, "-1.6", ""},
		{"-1.554", 2, "-1.56", ""},
		{"-0.554", 0, "-1", ""},
		{"-0.454", 0, "-1", ""},
		{"-0.454", 5, "-0.454", "-0.45400"},
		{"-5", 2, "-5", "-5.00"},
		{"-5", 1, "-5", "-5.0"},
		{"-5", 0, "-5", ""},
		{"-500", 2, "-500", "-500.00"},
		{"-500", -2, "-500", ""},
		{"-545", -1, "-550", ""},
		{"-545", -2, "-600", ""},
		{"-545", -3, "-1000", ""},
		{"-545", -4, "-10000", ""},
		{"-499", -3, "-1000", ""},
		{"-499", -4, "-10000", ""},
	}

	for _, test := range tests {
		d, err := NewDecimalFromString(test.input)
		if err != nil {
			t.Fatal(err)
		}

		// test Round
		expected, err := NewDecimalFromString(test.expected)
		if err != nil {
			t.Fatal(err)
		}
		got := d.RoundFloor(test.places)
		if !got.Equal(expected) {
			t.Errorf("Rounding floor %s to %d places, got %s, expected %s",
				d, test.places, got, expected)
		}

		// test StringFixed
		if test.expectedFixed == "" {
			test.expectedFixed = test.expected
		}
		gotStr := got.StringFixed(test.places)
		if gotStr != test.expectedFixed {
			t.Errorf("(%s).StringFixed(%d): got %s, expected %s",
				d, test.places, gotStr, test.expectedFixed)
		}
	}
}

func TestDecimal_RoundUpAndStringFixed(t *testing.T) {
	type testData struct {
		input         string
		places        int32
		expected      string
		expectedFixed string
	}
	tests := []testData{
		{"1.454", 0, "2", ""},
		{"1.454", 1, "1.5", ""},
		{"1.454", 2, "1.46", ""},
		{"1.454", 3, "1.454", ""},
		{"1.454", 4, "1.454", "1.4540"},
		{"1.454", 5, "1.454", "1.45400"},
		{"1.554", 0, "2", ""},
		{"1.554", 1, "1.6", ""},
		{"1.554", 2, "1.56", ""},
		{"0.554", 0, "1", ""},
		{"0.454", 0, "1", ""},
		{"0.454", 5, "0.454", "0.45400"},
		{"0", 0, "0", ""},
		{"0", 1, "0", "0.0"},
		{"0", 2, "0", "0.00"},
		{"0", -1, "0", ""},
		{"5", 2, "5", "5.00"},
		{"5", 1, "5", "5.0"},
		{"5", 0, "5", ""},
		{"500", 2, "500", "500.00"},
		{"500", -2, "500", ""},
		{"545", -1, "550", ""},
		{"545", -2, "600", ""},
		{"545", -3, "1000", ""},
		{"545", -4, "10000", ""},
		{"499", -3, "1000", ""},
		{"499", -4, "10000", ""},
		{"1.1001", 2, "1.11", ""},
		{"-1.1001", 2, "-1.11", ""},
		{"-1.454", 0, "-2", ""},
		{"-1.454", 1, "-1.5", ""},
		{"-1.454", 2, "-1.46", ""},
		{"-1.454", 3, "-1.454", ""},
		{"-1.454", 4, "-1.454", "-1.4540"},
		{"-1.454", 5, "-1.454", "-1.45400"},
		{"-1.554", 0, "-2", ""},
		{"-1.554", 1, "-1.6", ""},
		{"-1.554", 2, "-1.56", ""},
		{"-0.554", 0, "-1", ""},
		{"-0.454", 0, "-1", ""},
		{"-0.454", 5, "-0.454", "-0.45400"},
		{"-5", 2, "-5", "-5.00"},
		{"-5", 1, "-5", "-5.0"},
		{"-5", 0, "-5", ""},
		{"-500", 2, "-500", "-500.00"},
		{"-500", -2, "-500", ""},
		{"-545", -1, "-550", ""},
		{"-545", -2, "-600", ""},
		{"-545", -3, "-1000", ""},
		{"-545", -4, "-10000", ""},
		{"-499", -3, "-1000", ""},
		{"-499", -4, "-10000", ""},
	}

	for _, test := range tests {
		d, err := NewDecimalFromString(test.input)
		if err != nil {
			t.Fatal(err)
		}

		// test Round
		expected, err := NewDecimalFromString(test.expected)
		if err != nil {
			t.Fatal(err)
		}
		got := d.RoundUp(test.places)
		if !got.Equal(expected) {
			t.Errorf("Rounding up %s to %d places, got %s, expected %s",
				d, test.places, got, expected)
		}

		// test StringFixed
		if test.expectedFixed == "" {
			test.expectedFixed = test.expected
		}
		gotStr := got.StringFixed(test.places)
		if gotStr != test.expectedFixed {
			t.Errorf("(%s).StringFixed(%d): got %s, expected %s",
				d, test.places, gotStr, test.expectedFixed)
		}
	}
}

func TestDecimal_RoundDownAndStringFixed(t *testing.T) {
	type testData struct {
		input         string
		places        int32
		expected      string
		expectedFixed string
	}
	tests := []testData{
		{"1.454", 0, "1", ""},
		{"1.454", 1, "1.4", ""},
		{"1.454", 2, "1.45", ""},
		{"1.454", 3, "1.454", ""},
		{"1.454", 4, "1.454", "1.4540"},
		{"1.454", 5, "1.454", "1.45400"},
		{"1.554", 0, "1", ""},
		{"1.554", 1, "1.5", ""},
		{"1.554", 2, "1.55", ""},
		{"0.554", 0, "0", ""},
		{"0.454", 0, "0", ""},
		{"0.454", 5, "0.454", "0.45400"},
		{"0", 0, "0", ""},
		{"0", 1, "0", "0.0"},
		{"0", 2, "0", "0.00"},
		{"0", -1, "0", ""},
		{"5", 2, "5", "5.00"},
		{"5", 1, "5", "5.0"},
		{"5", 0, "5", ""},
		{"500", 2, "500", "500.00"},
		{"500", -2, "500", ""},
		{"545", -1, "540", ""},
		{"545", -2, "500", ""},
		{"545", -3, "0", ""},
		{"545", -4, "0", ""},
		{"499", -3, "0", ""},
		{"499", -4, "0", ""},
		{"1.1001", 2, "1.10", ""},
		{"-1.1001", 2, "-1.10", ""},
		{"-1.454", 0, "-1", ""},
		{"-1.454", 1, "-1.4", ""},
		{"-1.454", 2, "-1.45", ""},
		{"-1.454", 3, "-1.454", ""},
		{"-1.454", 4, "-1.454", "-1.4540"},
		{"-1.454", 5, "-1.454", "-1.45400"},
		{"-1.554", 0, "-1", ""},
		{"-1.554", 1, "-1.5", ""},
		{"-1.554", 2, "-1.55", ""},
		{"-0.554", 0, "0", ""},
		{"-0.454", 0, "0", ""},
		{"-0.454", 5, "-0.454", "-0.45400"},
		{"-5", 2, "-5", "-5.00"},
		{"-5", 1, "-5", "-5.0"},
		{"-5", 0, "-5", ""},
		{"-500", 2, "-500", "-500.00"},
		{"-500", -2, "-500", ""},
		{"-545", -1, "-540", ""},
		{"-545", -2, "-500", ""},
		{"-545", -3, "0", ""},
		{"-545", -4, "0", ""},
		{"-499", -3, "0", ""},
		{"-499", -4, "0", ""},
	}

	for _, test := range tests {
		d, err := NewDecimalFromString(test.input)
		if err != nil {
			t.Fatal(err)
		}

		// test Round
		expected, err := NewDecimalFromString(test.expected)
		if err != nil {
			t.Fatal(err)
		}
		got := d.RoundDown(test.places)
		if !got.Equal(expected) {
			t.Errorf("Rounding down %s to %d places, got %s, expected %s",
				d, test.places, got, expected)
		}

		// test StringFixed
		if test.expectedFixed == "" {
			test.expectedFixed = test.expected
		}
		gotStr := got.StringFixed(test.places)
		if gotStr != test.expectedFixed {
			t.Errorf("(%s).StringFixed(%d): got %s, expected %s",
				d, test.places, gotStr, test.expectedFixed)
		}
	}
}

func TestDecimal_QuoRem(t *testing.T) {
	type Inp4 struct {
		d   string
		d2  string
		exp int32
		q   string
		r   string
	}
	cases := []Inp4{
		{"10", "1", 0, "10", "0"},
		{"1", "10", 0, "0", "1"},
		{"1", "4", 2, "0.25", "0"},
		{"1", "8", 2, "0.12", "0.04"},
		{"10", "3", 1, "3.3", "0.1"},
		{"100", "3", 1, "33.3", "0.1"},
		{"1000", "10", -3, "0", "1000"},
		{"1e-3", "2e-5", 0, "50", "0"},
		{"1e-3", "2e-3", 1, "0.5", "0"},
		{"4e-3", "0.8", 4, "5e-3", "0"},
		{"4.1e-3", "0.8", 3, "5e-3", "1e-4"},
		{"-4", "-3", 0, "1", "-1"},
		{"-4", "3", 0, "-1", "-1"},
	}

	for _, inp4 := range cases {
		d, _ := NewDecimalFromString(inp4.d)
		d2, _ := NewDecimalFromString(inp4.d2)
		prec := inp4.exp
		q, r := d.QuoRem(d2, prec)
		expectedQ, _ := NewDecimalFromString(inp4.q)
		expectedR, _ := NewDecimalFromString(inp4.r)
		if !q.Equal(expectedQ) || !r.Equal(expectedR) {
			t.Errorf("bad QuoRem division %s , %s , %d got %v, %v expected %s , %s",
				inp4.d, inp4.d2, prec, q, r, inp4.q, inp4.r)
		}
		if !d.Equal(d2.Mul(q).Add(r)) {
			t.Errorf("not fitting: d=%v, d2= %v, prec=%d, q=%v, r=%v",
				d, d2, prec, q, r)
		}
		if !q.Equal(q.Truncate(prec)) {
			t.Errorf("quotient wrong precision: d=%v, d2= %v, prec=%d, q=%v, r=%v",
				d, d2, prec, q, r)
		}
		if r.Abs().Cmp(d2.Abs().Mul(NewDecimal(1, -prec))) >= 0 {
			t.Errorf("remainder too large: d=%v, d2= %v, prec=%d, q=%v, r=%v",
				d, d2, prec, q, r)
		}
		if r.value.Sign()*d.value.Sign() < 0 {
			t.Errorf("signum of divisor and rest do not match: d=%v, d2= %v, prec=%d, q=%v, r=%v",
				d, d2, prec, q, r)
		}
	}
}

func TestDecimal_Div(t *testing.T) {
	type Inp struct {
		a string
		b string
	}

	inputs := map[Inp]string{
		Inp{"6", "3"}:                            "2",
		Inp{"10", "2"}:                           "5",
		Inp{"2.2", "1.1"}:                        "2",
		Inp{"-2.2", "-1.1"}:                      "2",
		Inp{"12.88", "5.6"}:                      "2.3",
		Inp{"1023427554493", "43432632"}:         "23563.5628642767953828", // rounded
		Inp{"1", "434324545566634"}:              "0.0000000000000023",
		Inp{"1", "3"}:                            "0.3333333333333333",
		Inp{"2", "3"}:                            "0.6666666666666667", // rounded
		Inp{"10000", "3"}:                        "3333.3333333333333333",
		Inp{"10234274355545544493", "-3"}:        "-3411424785181848164.3333333333333333",
		Inp{"-4612301402398.4753343454", "23.5"}: "-196268144782.9138440146978723",
	}

	for inp, expectedStr := range inputs {
		num, err := NewDecimalFromString(inp.a)
		if err != nil {
			t.FailNow()
		}
		denom, err := NewDecimalFromString(inp.b)
		if err != nil {
			t.FailNow()
		}
		got := num.Div(denom)
		expected, _ := NewDecimalFromString(expectedStr)
		if !got.Equal(expected) {
			t.Errorf("expected %v when dividing %v by %v, got %v",
				expected, num, denom, got)
		}
		got2 := num.DivRound(denom, int32(DivisionPrecision))
		if !got2.Equal(expected) {
			t.Errorf("expected %v on DivRound (%v,%v), got %v", expected, num, denom, got2)
		}
	}

	type Inp2 struct {
		n    int64
		exp  int32
		n2   int64
		exp2 int32
	}

	// test code path where exp > 0
	inputs2 := map[Inp2]string{
		Inp2{124, 10, 3, 1}: "41333333333.3333333333333333",
		Inp2{124, 10, 3, 0}: "413333333333.3333333333333333",
		Inp2{124, 10, 6, 1}: "20666666666.6666666666666667",
		Inp2{124, 10, 6, 0}: "206666666666.6666666666666667",
		Inp2{10, 10, 10, 1}: "1000000000",
	}

	for inp, expectedAbs := range inputs2 {
		for i := -1; i <= 1; i += 2 {
			for j := -1; j <= 1; j += 2 {
				n := inp.n * int64(i)
				n2 := inp.n2 * int64(j)
				num := NewDecimal(n, inp.exp)
				denom := NewDecimal(n2, inp.exp2)
				expected := expectedAbs
				if i != j {
					expected = "-" + expectedAbs
				}
				got := num.Div(denom)
				if got.String() != expected {
					t.Errorf("expected %s when dividing %v by %v, got %v",
						expected, num, denom, got)
				}
			}
		}
	}
}

func TestDecimal_Add(t *testing.T) {
	type Inp struct {
		a string
		b string
	}

	inputs := map[Inp]string{
		Inp{"2", "3"}:                     "5",
		Inp{"2454495034", "3451204593"}:   "5905699627",
		Inp{"24544.95034", ".3451204593"}: "24545.2954604593",
		Inp{".1", ".1"}:                   "0.2",
		Inp{".1", "-.1"}:                  "0",
		Inp{"0", "1.001"}:                 "1.001",
	}

	for inp, res := range inputs {
		a, err := NewDecimalFromString(inp.a)
		if err != nil {
			t.FailNow()
		}
		b, err := NewDecimalFromString(inp.b)
		if err != nil {
			t.FailNow()
		}
		c := a.Add(b)
		if c.String() != res {
			t.Errorf("expected %s, got %s", res, c.String())
		}
	}
}

func TestDecimal_Sub(t *testing.T) {
	type Inp struct {
		a string
		b string
	}

	inputs := map[Inp]string{
		Inp{"2", "3"}:                     "-1",
		Inp{"12", "3"}:                    "9",
		Inp{"-2", "9"}:                    "-11",
		Inp{"2454495034", "3451204593"}:   "-996709559",
		Inp{"24544.95034", ".3451204593"}: "24544.6052195407",
		Inp{".1", "-.1"}:                  "0.2",
		Inp{".1", ".1"}:                   "0",
		Inp{"0", "1.001"}:                 "-1.001",
		Inp{"1.001", "0"}:                 "1.001",
		Inp{"2.3", ".3"}:                  "2",
	}

	for inp, res := range inputs {
		a, err := NewDecimalFromString(inp.a)
		if err != nil {
			t.FailNow()
		}
		b, err := NewDecimalFromString(inp.b)
		if err != nil {
			t.FailNow()
		}
		c := a.Sub(b)
		if c.String() != res {
			t.Errorf("expected %s, got %s", res, c.String())
		}
	}
}

func TestDecimal_Neg(t *testing.T) {
	inputs := map[string]string{
		"0":     "0",
		"10":    "-10",
		"5.56":  "-5.56",
		"-10":   "10",
		"-5.56": "5.56",
	}

	for inp, res := range inputs {
		a, err := NewDecimalFromString(inp)
		if err != nil {
			t.FailNow()
		}
		b := a.Neg()
		if b.String() != res {
			t.Errorf("expected %s, got %s", res, b.String())
		}
	}
}

func TestDecimal_Mul(t *testing.T) {
	type Inp struct {
		a string
		b string
	}

	inputs := map[Inp]string{
		Inp{"2", "3"}:                     "6",
		Inp{"2454495034", "3451204593"}:   "8470964534836491162",
		Inp{"24544.95034", ".3451204593"}: "8470.964534836491162",
		Inp{".1", ".1"}:                   "0.01",
		Inp{"0", "1.001"}:                 "0",
	}

	for inp, res := range inputs {
		a, err := NewDecimalFromString(inp.a)
		if err != nil {
			t.FailNow()
		}
		b, err := NewDecimalFromString(inp.b)
		if err != nil {
			t.FailNow()
		}
		c := a.Mul(b)
		if c.String() != res {
			t.Errorf("expected %s, got %s", res, c.String())
		}
	}

	// positive scale
	c := NewDecimal(1234, 5).Mul(NewDecimal(45, -1))
	if c.String() != "555300000" {
		t.Errorf("Expected %s, got %s", "555300000", c.String())
	}
}

func TestDecimal_Pow(t *testing.T) {
	a := NewDecimal(4, 0)
	b := NewDecimal(2, 0)
	x := a.Pow(b)
	if x.String() != "16" {
		t.Errorf("Error, saw %s", x.String())
	}
}

func TestDecimal_NegativePow(t *testing.T) {
	a := NewDecimal(4, 0)
	b := NewDecimal(-2, 0)
	x := a.Pow(b)
	if x.String() != "0.0625" {
		t.Errorf("Error, saw %s", x.String())
	}
}

func TestDecimal_IsInteger(t *testing.T) {
	for _, testCase := range []struct {
		Dec       string
		IsInteger bool
	}{
		{"0", true},
		{"0.0000", true},
		{"0.01", false},
		{"0.01010101010000", false},
		{"12.0", true},
		{"12.00000000000000", true},
		{"12.10000", false},
		{"9999.0000", true},
		{"99999999.000000000", true},
		{"-656323444.0000000000000", true},
		{"-32768.01234", false},
		{"-32768.0123423562623600000", false},
	} {
		d, err := NewDecimalFromString(testCase.Dec)
		if err != nil {
			t.Fatal(err)
		}
		if d.IsInteger() != testCase.IsInteger {
			t.Errorf("expect %t, got %t, for %s", testCase.IsInteger, d.IsInteger(), testCase.Dec)
		}
	}
}

func TestDecimalStringAFT(t *testing.T) {
	tests := []struct {
		d Decimal
		s string
	}{
		{NewDecimalFromFloat(-12000.001001), "-12000.001"},
		{NewDecimalFromFloat(10000.0000), "10000.00"},
		{NewDecimalFromFloat(12000.000000), "12000.00"},
		{NewDecimalFromFloat(12000.000001), "12000.00"},
		{NewDecimalFromFloat(-12000.000001), "-12000.00"},
		{NewDecimalFromFloat(12000.1), "12000.10"},
		{NewDecimalFromFloat(12000.1230), "12000.123"},
		{NewDecimalFromFloat(12000.1230), "12000.123"},
		{NewDecimalFromInt(1), "1.00"},
		{NewDecimalFromInt(0), "0.00"},
		{Zero, "0.00"},
	}

	for i, test := range tests {
		require.Equalf(t, test.s, test.d.StringAFT(), "test index: %d", i)
	}
}

func TestScanEmptyString(t *testing.T) {
	d := OptDecimal{}

	err := d.Scan("")

	require.Error(t, err)
}

func TestOptDecimal_LogValue(t *testing.T) {
	v := OptDecimal{}

	val := v.LogValue()

	assert.Equal(t, slog.KindAny, val.Kind())

	v.SetValue(NewDecimalFromInt(5))
	val = v.LogValue()

	assert.Equal(t, slog.KindString, val.Kind())
}
