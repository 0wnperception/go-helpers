// //nolint: lll,errorlint,gochecknoglobals,goerr113,mnd,funlen,varcheck,nestif,gocognit,gofumpt,unused,deadcode,gocyclo,gocritic,errcheck,revive,varnamelen
// modified version of https://github.com/shopspring/decimal
package types

import (
	"database/sql/driver"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/genproto/googleapis/type/money"

	"github.com/0wnperception/go-helpers/pkg/easyjson/jlexer"
	"github.com/0wnperception/go-helpers/pkg/easyjson/jwriter"
)

const nbase = 10000

// DivisionPrecision is the number of decimal places in the result when it
// doesn't divide exactly.
//
// Example:
//
//	d1 := decimal.NewDecimalFromFloat(2).Div(decimal.NewDecimalFromFloat(3)
//	d1.String() // output: "0.6666666666666667"
//	d2 := decimal.NewDecimalFromFloat(2).Div(decimal.NewDecimalFromFloat(30000)
//	d2.String() // output: "0.0000666666666667"
//	d3 := decimal.NewDecimalFromFloat(20000).Div(decimal.NewDecimalFromFloat(3)
//	d3.String() // output: "6666.6666666666666667"
//	decimal.DivisionPrecision = 3
//	d4 := decimal.NewDecimalFromFloat(2).Div(decimal.NewDecimalFromFloat(3)
//	d4.String() // output: "0.667"
var DivisionPrecision = 16

// PowPrecisionNegativeExponent specifies the maximum precision of the result (digits after decimal point)
// when calculating decimal power. Only used for cases where the exponent is a negative number.
// This constant applies to Pow, PowInt32 and PowBigInt methods, PowWithPrecision method is not constrained by it.
//
// Example:
//
//	d1, err := decimal.NewFromFloat(15.2).PowInt32(-2)
//	d1.String() // output: "0.0043282548476454"
//
//	decimal.PowPrecisionNegativeExponent = 24
//	d2, err := decimal.NewFromFloat(15.2).PowInt32(-2)
//	d2.String() // output: "0.004328254847645429362881"
var PowPrecisionNegativeExponent = 16

// MarshalJSONWithoutQuotes should be set to true if you want the decimal to
// be JSON marshaled as a number, instead of as a string.
// WARNING: this is dangerous for decimals with many digits, since many JSON
// unmarshallers (ex: Javascript's) will unmarshal JSON numbers to IEEE 754
// double-precision floating point numbers, which means you can potentially
// silently lose precision.
var MarshalJSONWithoutQuotes = true

// Zero constant, to make computations faster.
var Zero = NewDecimal(0, 0)

var (
	One     = NewDecimal(1, 0)
	Ten     = NewDecimal(1, 1)
	Hundred = NewDecimal(1, 2)
)

var (
	zeroInt    = big.NewInt(0)
	oneInt     = big.NewInt(1)
	twoInt     = big.NewInt(2)
	fiveInt    = big.NewInt(5)
	tenInt     = big.NewInt(10)
	factorials = []Decimal{NewDecimal(1, 0)}
)

var (
	big0       = big.NewInt(0)
	big1       = big.NewInt(1)
	big10      = big.NewInt(10)
	big100     = big.NewInt(100)
	big1000    = big.NewInt(1000)
	bigNBase   = big.NewInt(nbase)
	big10000   = bigNBase
	big100000  = big.NewInt(100000)
	big1000000 = big.NewInt(1000000)
	bigNBaseX2 = big.NewInt(nbase * nbase)
	bigNBaseX3 = big.NewInt(nbase * nbase * nbase)
	bigNBaseX4 = big.NewInt(nbase * nbase * nbase * nbase)
)

var exps = []*big.Int{big1, big10, big100, big1000, big10000, big100000, big1000000}

const (
	maxUint = ^uint(0)
	maxInt  = int(maxUint >> 1)
	minInt  = -maxInt - 1
)

var (
	bigMaxUint8  = big.NewInt(math.MaxUint8)
	bigMaxUint16 = big.NewInt(math.MaxUint16)
	bigMaxUint32 = big.NewInt(math.MaxUint32)
	bigMaxUint64 = (&big.Int{}).SetUint64(uint64(math.MaxUint64))
	bigMaxUint   = (&big.Int{}).SetUint64(uint64(maxUint))
	bigMaxInt8   = big.NewInt(math.MaxInt8)
	bigMinInt8   = big.NewInt(math.MinInt8)
	bigMaxInt16  = big.NewInt(math.MaxInt16)
	bigMinInt16  = big.NewInt(math.MinInt16)
	bigMaxInt32  = big.NewInt(math.MaxInt32)
	bigMinInt32  = big.NewInt(math.MinInt32)
	bigMaxInt64  = big.NewInt(math.MaxInt64)
	bigMinInt64  = big.NewInt(math.MinInt64)
	bigMaxInt    = big.NewInt(int64(maxInt))
	bigMinInt    = big.NewInt(int64(minInt))
)

// Decimal represents a fixed-point decimal. It is immutable number = value * 10 ^ exp.
type Decimal struct {
	value *big.Int
	exp   int32
}

// NewDecimal returns a new fixed-point decimal, value * 10 ^ exp.
func NewDecimal(value int64, exp int32) Decimal {
	ret := Decimal{
		value: big.NewInt(value),
		exp:   exp,
	}

	ret.Normalize()

	return ret
}

func NewDecimalFromMoneyParts(units int64, nanos int32) Decimal {
	if units == 0 && nanos == 0 {
		return NewDecimalFromInt(0)
	}

	neg := false

	if units == 0 {
		if nanos < 0 {
			neg = true
			nanos = -nanos
		}

		var exp int32 = -9

		for nanos%10 == 0 && exp <= 0 {
			nanos /= 10
			exp++
		}

		v := Decimal{value: big.NewInt(int64(nanos)), exp: exp}

		if neg {
			v = v.Neg()
		}

		return v
	}

	if units < 0 {
		neg = true

		units = abs(units)

		if nanos < 0 {
			nanos = -nanos
		}
	} else if nanos < 0 {
		nanos = -nanos
	}

	if nanos == 0 {
		v := NewDecimal(units, 0)

		if neg {
			v = v.Neg()
		}

		return v
	}

	var exp int32 = -9

	for nanos%10 == 0 && exp <= 0 {
		nanos /= 10
		exp++
	}

	v := NewDecimal(units, -exp).Add(NewDecimalFromInt(int64(nanos)))

	v.exp += exp

	if neg {
		v = v.Neg()
	}

	return v
}

// NewDecimalFromInt converts a int64 to Decimal.
//
// Example:
//
//	NewDecimalFromInt(123).String() // output: "123"
//	NewDecimalFromInt(-10).String() // output: "-10"
func NewDecimalFromInt(value int64) Decimal {
	d := Decimal{
		value: big.NewInt(value),
		exp:   0,
	}

	d.Normalize()

	return d
}

func NewDecimalFromString(value string) (Decimal, error) {
	originalInput := value

	var intString string

	var exp int64

	// Check if number is using scientific notation
	eIndex := strings.IndexAny(value, "Ee")
	if eIndex != -1 {
		expInt, err := strconv.ParseInt(value[eIndex+1:], 10, 32)
		if err != nil {
			if e, ok := err.(*strconv.NumError); ok && e.Err == strconv.ErrRange {
				return Decimal{}, fmt.Errorf("can't convert %s to decimal: fractional part too long", value)
			}

			return Decimal{}, fmt.Errorf("can't convert %s to decimal: exponent is not numeric", value)
		}

		value = value[:eIndex]
		exp = expInt
	}

	pIndex := -1
	vLen := len(value)

	for i := range vLen {
		if value[i] == '.' {
			if pIndex > -1 {
				return Decimal{}, fmt.Errorf("can't convert %s to decimal: too many .s", value)
			}

			pIndex = i
		}
	}

	if pIndex == -1 {
		// There is no decimal point, we can just parse the original string as
		// an int
		intString = value
	} else {
		if pIndex+1 < vLen {
			intString = value[:pIndex] + value[pIndex+1:]
		} else {
			intString = value[:pIndex]
		}

		expInt := -len(value[pIndex+1:])
		exp += int64(expInt)
	}

	var dValue *big.Int
	// strconv.ParseInt is faster than new(big.Int).SetString so this is just a shortcut for strings we know won't overflow
	if len(intString) <= 18 {
		parsed64, err := strconv.ParseInt(intString, 10, 64)
		if err != nil {
			return Decimal{}, fmt.Errorf("can't convert %s to decimal", value)
		}

		dValue = big.NewInt(parsed64)
	} else {
		dValue = new(big.Int)

		_, ok := dValue.SetString(intString, 10)
		if !ok {
			return Decimal{}, fmt.Errorf("can't convert %s to decimal", value)
		}
	}

	if exp < math.MinInt32 || exp > math.MaxInt32 {
		// NOTE(vadim): I doubt a string could realistically be this long
		return Decimal{}, fmt.Errorf("can't convert %s to decimal: fractional part too long", originalInput)
	}

	d := Decimal{
		value: dValue,
		exp:   int32(exp), //nolint:gosec
	}

	d.Normalize()

	return d, nil
}

func NewDecimalFromUint(value uint64) Decimal {
	return Decimal{
		value: new(big.Int).SetUint64(value),
		exp:   0,
	}
}

// NewDecimalFromFloat converts a float64 to Decimal.
//
// Example:
//
//	NewDecimalFromFloat(123.45678901234567).String() // output: "123.4567890123456"
//	NewDecimalFromFloat(.00000000000000001).String() // output: "0.00000000000000001"
//
// NOTE: this will panic on NaN, +/-inf.
func NewDecimalFromFloat(value float64) Decimal {
	if value == 0 {
		return NewDecimal(0, 0)
	}

	return newFromFloat(value, math.Float64bits(value), &float64info)
}

func NewDecimalFromFloat32(value float32) Decimal {
	if value == 0 {
		return NewDecimal(0, 0)
	}
	// XOR is workaround for https://github.com/golang/go/issues/26285
	a := math.Float32bits(value) ^ 0x80808080

	return newFromFloat(float64(value), uint64(a)^0x80808080, &float32info)
}

func newFromFloat(val float64, bits uint64, flt *floatInfo) Decimal {
	if math.IsNaN(val) || math.IsInf(val, 0) {
		panic(fmt.Sprintf("Cannot create a Decimal from %v", val))
	}

	exp := int(bits>>flt.mantbits) & (1<<flt.expbits - 1) //nolint:gosec
	mant := bits & (uint64(1)<<flt.mantbits - 1)

	switch exp {
	case 0:
		// denormalized
		exp++

	default:
		// add implicit top bit
		mant |= uint64(1) << flt.mantbits
	}

	exp += flt.bias

	var d decimal

	d.Assign(mant)
	d.Shift(exp - int(flt.mantbits)) //nolint:gosec
	d.neg = bits>>(flt.expbits+flt.mantbits) != 0

	roundShortest(&d, mant, exp, flt)
	// If less than 19 digits, we can do calculation in an int64.
	if d.nd < 19 {
		tmp := int64(0)
		m := int64(1)

		for i := d.nd - 1; i >= 0; i-- {
			tmp += m * int64(d.d[i]-'0')
			m *= 10
		}

		if d.neg {
			tmp *= -1
		}

		return Decimal{value: big.NewInt(tmp), exp: int32(d.dp) - int32(d.nd)} //nolint:gosec
	}

	dValue, ok := new(big.Int).SetString(string(d.d[:d.nd]), 10)
	if ok {
		return Decimal{value: dValue, exp: int32(d.dp) - int32(d.nd)} //nolint:gosec
	}

	return NewDecimalFromFloatWithExponent(val, int32(d.dp)-int32(d.nd)) //nolint:gosec
}

func NewFromMoney(m *money.Money) Decimal {
	return NewDecimalFromMoneyParts(m.GetUnits(), m.GetNanos())
}

func NewDecimalFromFloatWithExponent(value float64, exp int32) Decimal {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		panic(fmt.Sprintf("Cannot create a Decimal from %v", value))
	}

	bits := math.Float64bits(value)
	mant := bits & (1<<52 - 1)
	exp2 := int32((bits >> 52) & (1<<11 - 1)) //nolint:gosec

	if exp2 == 0 {
		// specials
		if mant == 0 {
			return Decimal{}
		}
		// subnormal
		exp2++
	} else {
		// normal
		mant |= 1 << 52
	}

	exp2 -= 1023 + 52

	// normalizing base-2 values
	for mant&1 == 0 {
		mant = mant >> 1
		exp2++
	}

	// maximum number of fractional base-10 digits to represent 2^N exactly cannot be more than -N if N<0
	if exp < 0 && exp < exp2 {
		if exp2 < 0 {
			exp = exp2
		} else {
			exp = 0
		}
	}

	// representing 10^M * 2^N as 5^M * 2^(M+N)
	exp2 -= exp

	temp := big.NewInt(1)
	dMant := big.NewInt(int64(mant)) //nolint:gosec

	// applying 5^M
	if exp > 0 {
		temp = temp.SetInt64(int64(exp))
		temp = temp.Exp(fiveInt, temp, nil)
	} else if exp < 0 {
		temp = temp.SetInt64(-int64(exp))
		temp = temp.Exp(fiveInt, temp, nil)
		dMant = dMant.Mul(dMant, temp)
		temp = temp.SetUint64(1)
	}

	// applying 2^(M+N)
	if exp2 > 0 {
		dMant = dMant.Lsh(dMant, uint(exp2))
	} else if exp2 < 0 {
		temp = temp.Lsh(temp, uint(-exp2))
	}

	// rounding and downscaling
	if exp > 0 || exp2 < 0 {
		halfDown := new(big.Int).Rsh(temp, 1)
		dMant = dMant.Add(dMant, halfDown)
		dMant = dMant.Quo(dMant, temp)
	}

	if sign := bits >> 63; sign == 1 {
		dMant = dMant.Neg(dMant)
	}

	return Decimal{
		value: dMant,
		exp:   exp,
	}
}

func abs(n int64) int64 {
	y := n >> 63

	return (n ^ y) - y
}

// rescale returns a rescaled version of the decimal. Returned
// decimal may be less precise if the given exponent is bigger
// than the initial exponent of the Decimal.
// NOTE: this will truncate, NOT round
//
// Example:
//
//	d := New(12345, -4)
//
// d2 := d.rescale(-1)
// d3 := d2.rescale(-4)
// println(d1)
// println(d2)
// println(d3)
//
// Output:
//
// 1.2345
// 1.2
// 1.2000.
func (d Decimal) rescale(exp int32) Decimal {
	d.ensureInitialized()

	if d.exp == exp {
		return Decimal{
			new(big.Int).Set(d.value),
			d.exp,
		}
	}

	diff := abs(int64(exp) - int64(d.exp))
	value := new(big.Int).Set(d.value)

	expScale := new(big.Int).Exp(tenInt, big.NewInt(diff), nil)

	if exp > d.exp {
		value = value.Quo(value, expScale)
	} else if exp < d.exp {
		value = value.Mul(value, expScale)
	}

	return Decimal{
		value: value,
		exp:   exp,
	}
}

// Abs returns the absolute value of the decimal.
func (d Decimal) Abs() Decimal {
	if !d.IsNegative() {
		return d
	}

	d.ensureInitialized()

	d2Value := new(big.Int).Abs(d.value)

	return Decimal{
		value: d2Value,
		exp:   d.exp,
	}
}

func (d Decimal) IsPositive() bool {
	return d.Sign() == 1
}

func (d Decimal) IsNegative() bool {
	return d.Sign() == -1
}

func (d Decimal) IsZero() bool {
	return d.Sign() == 0
}

func rescalePair(d1, d2 Decimal) (Decimal, Decimal) {
	d1.ensureInitialized()
	d2.ensureInitialized()

	if d1.exp == d2.exp {
		return d1, d2
	}

	baseScale := min(d1.exp, d2.exp)
	if baseScale != d1.exp {
		return d1.rescale(baseScale), d2
	}

	return d1, d2.rescale(baseScale)
}

// Add returns d + d2.
func (d Decimal) Add(d2 Decimal) Decimal {
	rd, rd2 := rescalePair(d, d2)

	d3Value := new(big.Int).Add(rd.value, rd2.value)

	return Decimal{
		value: d3Value,
		exp:   rd.exp,
	}
}

// Sub returns d - d2.
func (d Decimal) Sub(d2 Decimal) Decimal {
	rd, rd2 := rescalePair(d, d2)

	d3Value := new(big.Int).Sub(rd.value, rd2.value)

	ret := Decimal{
		value: d3Value,
		exp:   rd.exp,
	}

	ret.Normalize()

	return ret
}

// Neg returns -d.
func (d Decimal) Neg() Decimal {
	val := new(big.Int).Neg(d.value)

	return Decimal{
		value: val,
		exp:   d.exp,
	}
}

// Mul returns d * d2.
func (d Decimal) Mul(d2 Decimal) Decimal {
	d.ensureInitialized()
	d2.ensureInitialized()

	expInt64 := int64(d.exp) + int64(d2.exp)

	if expInt64 > math.MaxInt32 || expInt64 < math.MinInt32 {
		panic(fmt.Sprintf("exponent %v overflows an int32!", expInt64))
	}

	d3Value := new(big.Int).Mul(d.value, d2.value)

	ret := Decimal{
		value: d3Value,
		exp:   int32(expInt64), //nolint:gosec
	}

	ret.Normalize()

	return ret
}

func (d Decimal) exp10(n int) Decimal {
	d.ensureInitialized()
	expInt64 := int64(d.exp) + int64(n)

	if expInt64 > math.MaxInt32 || expInt64 < math.MinInt32 {
		panic(fmt.Sprintf("exponent %v overflows an int32!", expInt64))
	}

	return Decimal{
		value: new(big.Int).Set(d.value),
		exp:   int32(expInt64), //nolint:gosec
	}
}

// //nolint: gocognit
func (d *Decimal) Normalize() {
	if d.value != nil {
		sign := d.value.Sign()
		if sign != 0 {
			v := big.Int{}
			v.Abs(d.value)
			exp := d.exp

			if v.IsInt64() {
				i := v.Int64()

				if i%10 != 0 {
					return
				}

				for i%10 == 0 {
					i /= 10
					exp++
				}

				if sign < 0 {
					i = -i
				}

				d.exp = exp
				d.value = big.NewInt(i)
			} else {
				z := big.Int{}

				z.Set(&v)

				r := big.Int{}

				for {
					z.QuoRem(&z, big10, &r)

					if r.Sign() != 0 {
						break
					}

					exp++

					v.Set(&z)
				}

				if sign < 0 {
					v.Neg(&v)
				}

				d.value = &v
				d.exp = exp
			}
		}
	}
}

// IsProtoMoney reports whether x can be represented as a google.type.Money.
func (d Decimal) IsProtoMoney() bool {
	scaledD := d.rescale(0)

	return scaledD.value.IsInt64()
}

func (d Decimal) ToMoneyParts() (int64, int32) {
	s := d.Sign()

	if s == 0 {
		return 0, 0
	}

	v := d.Abs().Round(9)

	if v.exp >= 0 {
		uu := v.value

		for i := v.exp; i > 0; i-- {
			uu = uu.Mul(uu, big10)
		}

		if s < 0 {
			return -uu.Int64(), 0
		}

		return uu.Int64(), 0
	}

	u := v.rescale(0)

	units := u.value.Int64()

	fraction := v.Sub(u).Mul(NewDecimalFromInt(1_000_000_000)).rescale(0)

	nanos := int32(fraction.value.Int64()) //nolint:gosec

	if s < 0 {
		return -units, -nanos
	}

	return units, nanos
}

func (d Decimal) Mul10() Decimal {
	return d.exp10(1)
}

func (d Decimal) Mul100() Decimal {
	return d.exp10(2)
}

func (d Decimal) Mul1000() Decimal {
	return d.exp10(3)
}

func (d Decimal) Div10() Decimal {
	return d.exp10(-1)
}

func (d Decimal) Div100() Decimal {
	return d.exp10(-2)
}

func (d Decimal) Div1000() Decimal {
	return d.exp10(-3)
}

// Div returns d / d2. If it doesn't divide exactly, the result will have
// DivisionPrecision digits after the decimal point.
func (d Decimal) Div(d2 Decimal) Decimal {
	return d.DivRound(d2, int32(DivisionPrecision)) //nolint:gosec
}

// QuoRem does divsion with remainder
// d.QuoRem(d2,precision) returns quotient q and remainder r such that
//
//	d = d2 * q + r, q an integer multiple of 10^(-precision)
//	0 <= r < abs(d2) * 10 ^(-precision) if d>=0
//	0 >= r > -abs(d2) * 10 ^(-precision) if d<0
//
// Note that precision<0 is allowed as input.
func (d Decimal) QuoRem(d2 Decimal, precision int32) (Decimal, Decimal) {
	d.ensureInitialized()
	d2.ensureInitialized()

	if d2.value.Sign() == 0 {
		panic("decimal division by 0")
	}

	scale := -precision
	e := int64(d.exp) - int64(d2.exp) - int64(scale)

	if e > math.MaxInt32 || e < math.MinInt32 {
		panic("overflow in decimal QuoRem")
	}

	var aa, bb, expo big.Int

	var scalerest int32
	// d = a 10^ea
	// d2 = b 10^eb
	if e < 0 {
		aa = *d.value

		expo.SetInt64(-e)
		bb.Exp(tenInt, &expo, nil)
		bb.Mul(d2.value, &bb)

		scalerest = d.exp
	} else {
		expo.SetInt64(e)
		aa.Exp(tenInt, &expo, nil)
		aa.Mul(d.value, &aa)

		bb = *d2.value
		scalerest = scale + d2.exp
	}

	var q, r big.Int

	q.QuoRem(&aa, &bb, &r)

	dq := Decimal{value: &q, exp: scale}
	dr := Decimal{value: &r, exp: scalerest}

	return dq, dr
}

// RoundStep rounds the decimal to the nearest step value.
// Doesn't matter if step parameter is positive or negative,
// the result is the same.
// If lower is set to true, value will be rounded to the lower
// step value.
//
// Example:
//
//	    NewDecimalFromFloat(0.43).RoundStep(NewDecimalFromFloat(0.5), false).String() // output: "0.5"
//	    NewDecimalFromFloat(0.43).RoundStep(NewDecimalFromFloat(0.5), true).String() // output: "0"
//		   NewDecimalFromFloat(0.000355666).RoundStep(NewDecimalFromFloat(0.0005), false).String() // output: "0.0005"
//		   NewDecimalFromFloat(0.000355666).Round(NewDecimalFromFloat(0.0005), true).String() // output: "0"
//		   NewDecimalFromFloat(0.000355666).Round(NewDecimalFromFloat(0.0001), false).String() // output: "0.0004"
//		   NewDecimalFromFloat(0.000355666).Round(NewDecimalFromFloat(0.0001), true).String() // output: "0.0003"
func (d Decimal) RoundStep(step Decimal, lower bool) Decimal {
	step = step.Abs()

	if lower {
		return d.Div(step).Floor().Mul(step)
	}

	return d.Div(step).Round(0).Mul(step)
}

// DivRound divides and rounds to a given precision
// i.e. to an integer multiple of 10^(-precision)
//
//	for a positive quotient digit 5 is rounded up, away from 0
//	if the quotient is negative then digit 5 is rounded down, away from 0
//
// Note that precision<0 is allowed as input.
func (d Decimal) DivRound(d2 Decimal, precision int32) Decimal {
	// QuoRem already checks initialization
	q, r := d.QuoRem(d2, precision)
	// the actual rounding decision is based on comparing r*10^precision and d2/2
	// instead compare 2 r 10 ^precision and d2
	var rv2 big.Int

	rv2.Abs(r.value)
	rv2.Lsh(&rv2, 1)
	// now rv2 = abs(r.value) * 2
	r2 := Decimal{value: &rv2, exp: r.exp + precision}
	// r2 is now 2 * r * 10 ^ precision

	if c := r2.Cmp(d2.Abs()); c < 0 {
		return q
	}

	if d.value.Sign()*d2.value.Sign() < 0 {
		return q.Sub(NewDecimal(1, -precision))
	}

	return q.Add(NewDecimal(1, -precision))
}

// Mod returns d % d2.
func (d Decimal) Mod(d2 Decimal) Decimal {
	_, r := d.QuoRem(d2, 0)

	return r
}

// Compare compares the numbers represented by d and d2 and returns:
//
//	-1 if d <  d2
//	 0 if d == d2
//	+1 if d >  d2
func (d Decimal) Compare(d2 Decimal) int {
	return d.Cmp(d2)
}

// Pow returns d to the power d2.
func (d Decimal) Pow(d2 Decimal) Decimal {
	baseSign := d.Sign()
	expSign := d2.Sign()

	if baseSign == 0 {
		if expSign == 0 {
			return Decimal{}
		}

		if expSign == 1 {
			return Decimal{zeroInt, 0}
		}

		if expSign == -1 {
			return Decimal{}
		}
	}

	if expSign == 0 {
		return Decimal{oneInt, 0}
	}

	one := Decimal{oneInt, 0}
	expIntPart, expFracPart := d2.QuoRem(one, 0)

	if baseSign == -1 && !expFracPart.IsZero() {
		return Decimal{}
	}

	intPartPow, _ := d.PowBigInt(expIntPart.value)

	// if exponent is an integer we don't need to calculate d1**frac(d2)
	if expFracPart.value.Sign() == 0 {
		return intPartPow
	}

	digitsBase := d.NumDigits()
	digitsExponent := d2.NumDigits()

	precision := digitsBase

	if digitsExponent > precision {
		precision += digitsExponent
	}

	precision += 6

	// Calculate x ** frac(y), where
	// x ** frac(y) = exp(ln(x ** frac(y)) = exp(ln(x) * frac(y))
	fracPartPow, err := d.Abs().Ln(-d.exp + int32(precision)) //nolint:gosec
	if err != nil {
		return Decimal{}
	}

	fracPartPow = fracPartPow.Mul(expFracPart)

	fracPartPow, err = fracPartPow.ExpTaylor(-d.exp + int32(precision)) //nolint:gosec
	if err != nil {
		return Decimal{}
	}

	// Join integer and fractional part,
	// base ** (expBase + expFrac) = base ** expBase * base ** expFrac
	res := intPartPow.Mul(fracPartPow)

	return res
}

func (d Decimal) PowWithPrecision(d2 Decimal, precision int32) (Decimal, error) {
	baseSign := d.Sign()
	expSign := d2.Sign()

	if baseSign == 0 {
		if expSign == 0 {
			return Decimal{}, errors.New("cannot represent undefined value of 0**0")
		}

		if expSign == 1 {
			return Decimal{zeroInt, 0}, nil
		}

		if expSign == -1 {
			return Decimal{}, errors.New("cannot represent infinity value of 0 ** y, where y < 0")
		}
	}

	if expSign == 0 {
		return Decimal{oneInt, 0}, nil
	}

	one := Decimal{oneInt, 0}
	expIntPart, expFracPart := d2.QuoRem(one, 0)

	if baseSign == -1 && !expFracPart.IsZero() {
		return Decimal{}, errors.New("cannot represent imaginary value of x ** y, where x < 0 and y is non-integer decimal")
	}

	intPartPow, _ := d.powBigIntWithPrecision(expIntPart.value, precision)

	// if exponent is an integer we don't need to calculate d1**frac(d2)
	if expFracPart.value.Sign() == 0 {
		return intPartPow, nil
	}

	digitsBase := d.NumDigits()
	digitsExponent := d2.NumDigits()

	if int32(digitsBase) > precision { //nolint:gosec
		precision = int32(digitsBase) //nolint:gosec
	}

	if int32(digitsExponent) > precision { //nolint:gosec
		precision += int32(digitsExponent) //nolint:gosec
	}
	// increase precision by 10 to compensate for errors in further calculations
	precision += 10

	// Calculate x ** frac(y), where
	// x ** frac(y) = exp(ln(x ** frac(y)) = exp(ln(x) * frac(y))
	fracPartPow, err := d.Abs().Ln(precision)
	if err != nil {
		return Decimal{}, err
	}

	fracPartPow = fracPartPow.Mul(expFracPart)

	fracPartPow, err = fracPartPow.ExpTaylor(precision)
	if err != nil {
		return Decimal{}, err
	}

	// Join integer and fractional part,
	// base ** (expBase + expFrac) = base ** expBase * base ** expFrac
	res := intPartPow.Mul(fracPartPow)

	return res, nil
}

func (d Decimal) PowInt32(exp int32) (Decimal, error) {
	if d.IsZero() && exp == 0 {
		return Decimal{}, errors.New("cannot represent undefined value of 0**0")
	}

	isExpNeg := false
	if exp < 0 {
		isExpNeg = true
		exp = -exp
	}

	n, result := d, NewDecimal(1, 0)

	for exp > 0 {
		if exp%2 == 1 {
			result = result.Mul(n)
		}

		exp /= 2

		if exp > 0 {
			n = n.Mul(n)
		}
	}

	if isExpNeg {
		return NewDecimal(1, 0).DivRound(result, int32(PowPrecisionNegativeExponent)), nil //nolint:gosec
	}

	return result, nil
}

func (d Decimal) PowBigInt(exp *big.Int) (Decimal, error) {
	return d.powBigIntWithPrecision(exp, int32(PowPrecisionNegativeExponent)) //nolint:gosec
}

func (d Decimal) powBigIntWithPrecision(exp *big.Int, precision int32) (Decimal, error) {
	if d.IsZero() && exp.Sign() == 0 {
		return Decimal{}, errors.New("cannot represent undefined value of 0**0")
	}

	tmpExp := new(big.Int).Set(exp)
	isExpNeg := exp.Sign() < 0

	if isExpNeg {
		tmpExp.Abs(tmpExp)
	}

	n, result := d, NewDecimal(1, 0)

	for tmpExp.Sign() > 0 {
		if tmpExp.Bit(0) == 1 {
			result = result.Mul(n)
		}

		tmpExp.Rsh(tmpExp, 1)

		if tmpExp.Sign() > 0 {
			n = n.Mul(n)
		}
	}

	if isExpNeg {
		return NewDecimal(1, 0).DivRound(result, precision), nil
	}

	return result, nil
}

// NumDigits returns the number of digits of the decimal coefficient (d.Value)
func (d Decimal) NumDigits() int {
	if d.value == nil {
		return 1
	}

	if d.value.IsInt64() {
		i64 := d.value.Int64()
		// restrict fast path to integers with exact conversion to float64
		if i64 <= (1<<53) && i64 >= -(1<<53) {
			if i64 == 0 {
				return 1
			}

			return int(math.Log10(math.Abs(float64(i64)))) + 1
		}
	}

	estimatedNumDigits := int(float64(d.value.BitLen()) / math.Log2(10))

	// estimatedNumDigits (lg10) may be off by 1, need to verify
	digitsBigInt := big.NewInt(int64(estimatedNumDigits))
	errorCorrectionUnit := digitsBigInt.Exp(tenInt, digitsBigInt, nil)

	if d.value.CmpAbs(errorCorrectionUnit) >= 0 {
		return estimatedNumDigits + 1
	}

	return estimatedNumDigits
}

// Ln calculates natural logarithm of d.
// Precision argument specifies how precise the result must be (number of digits after decimal point).
// Negative precision is allowed.
//
// Example:
//
//	d1, err := NewFromFloat(13.3).Ln(2)
//	d1.String()  // output: "2.59"
//
//	d2, err := NewFromFloat(579.161).Ln(10)
//	d2.String()  // output: "6.3615805046"
func (d Decimal) Ln(precision int32) (Decimal, error) {
	// Algorithm based on The Use of Iteration Methods for Approximating the Natural Logarithm,
	// James F. Epperson, The American Mathematical Monthly, Vol. 96, No. 9, November 1989, pp. 831-835.
	if d.IsNegative() {
		return Decimal{}, errors.New("cannot calculate natural logarithm for negative decimals")
	}

	if d.IsZero() {
		return Decimal{}, errors.New("cannot represent natural logarithm of 0, result: -infinity")
	}

	calcPrecision := precision + 2
	z := d.Copy()

	var comp1, comp3, comp2, comp4, reduceAdjust Decimal
	comp1 = z.Sub(Decimal{oneInt, 0})
	comp3 = Decimal{oneInt, -1}

	// for decimal in range [0.9, 1.1] where ln(d) is close to 0
	usePowerSeries := false

	if comp1.Abs().Cmp(comp3) <= 0 {
		usePowerSeries = true
	} else {
		// reduce input decimal to range [0.1, 1)
		expDelta := int32(z.NumDigits()) + z.exp //nolint:gosec
		z.exp -= expDelta

		// Input decimal was reduced by factor of 10^expDelta, thus we will need to add
		// ln(10^expDelta) = expDelta * ln(10)
		// to the result to compensate that
		ln10 := ln10.withPrecision(calcPrecision)
		reduceAdjust = NewDecimalFromInt(int64(expDelta))
		reduceAdjust = reduceAdjust.Mul(ln10)

		comp1 = z.Sub(Decimal{oneInt, 0})

		if comp1.Abs().Cmp(comp3) <= 0 {
			usePowerSeries = true
		} else {
			// initial estimate using floats
			zFloat := z.InexactFloat64()
			comp1 = NewDecimalFromFloat(math.Log(zFloat))
		}
	}

	epsilon := Decimal{oneInt, -calcPrecision}

	if usePowerSeries {
		comp2 = comp1.Add(Decimal{twoInt, 0})
		// z / (z + 2)
		comp3 = comp1.DivRound(comp2, calcPrecision)
		// 2 * (z / (z + 2))
		comp1 = comp3.Add(comp3)
		comp2 = comp1.Copy()

		for n := 1; ; n++ {
			// 2 * (z / (z+2))^(2n+1)
			comp2 = comp2.Mul(comp3).Mul(comp3)

			// 1 / (2n+1) * 2 * (z / (z+2))^(2n+1)
			comp4 = NewDecimalFromInt(int64(2*n + 1))
			comp4 = comp2.DivRound(comp4, calcPrecision)

			// comp1 = 2 sum [ 1 / (2n+1) * (z / (z+2))^(2n+1) ]
			comp1 = comp1.Add(comp4)

			if comp4.Abs().Cmp(epsilon) <= 0 {
				break
			}
		}
	} else {
		var prevStep Decimal

		maxIters := calcPrecision*2 + 10

		for i := int32(0); i < maxIters; i++ { //nolint:intrange
			// exp(a_n)
			comp3, _ = comp1.ExpTaylor(calcPrecision)
			// exp(a_n) - z
			comp2 = comp3.Sub(z)
			// 2 * (exp(a_n) - z)
			comp2 = comp2.Add(comp2)
			// exp(a_n) + z
			comp4 = comp3.Add(z)
			// 2 * (exp(a_n) - z) / (exp(a_n) + z)
			comp3 = comp2.DivRound(comp4, calcPrecision)
			// comp1 = a_(n+1) = a_n - 2 * (exp(a_n) - z) / (exp(a_n) + z)
			comp1 = comp1.Sub(comp3)

			if prevStep.Add(comp3).IsZero() {
				// If iteration steps oscillate we should return early and prevent an infinity loop
				// NOTE(mwoss): This should be quite a rare case, returning error is not necessary
				break
			}

			if comp3.Abs().Cmp(epsilon) <= 0 {
				break
			}

			prevStep = comp3
		}
	}

	comp1 = comp1.Add(reduceAdjust)

	return comp1.Round(precision), nil
}

// ExpTaylor calculates the natural exponent of decimal (e to the power of d) using Taylor series expansion.
// Precision argument specifies how precise the result must be (number of digits after decimal point).
// Negative precision is allowed.
//
// ExpTaylor is much faster for large precision values than ExpHullAbrham.
//
// Example:
//
//	d, err := NewFromFloat(26.1).ExpTaylor(2).String()
//	d.String()  // output: "216314672147.06"
//
//	NewFromFloat(26.1).ExpTaylor(20).String()
//	d.String()  // output: "216314672147.05767284062928674083"
//
//	NewFromFloat(26.1).ExpTaylor(-10).String()
//	d.String()  // output: "220000000000"
func (d Decimal) ExpTaylor(precision int32) (Decimal, error) {
	// Note(mwoss): Implementation can be optimized by exclusively using big.Int API only
	if d.IsZero() {
		return Decimal{oneInt, 0}.Round(precision), nil
	}

	var epsilon Decimal

	var divPrecision int32

	if precision < 0 {
		epsilon = NewDecimal(1, -1)
		divPrecision = 8
	} else {
		epsilon = NewDecimal(1, -precision-1)
		divPrecision = precision + 1
	}

	decAbs := d.Abs()
	pow := d.Abs()
	factorial := NewDecimal(1, 0)

	result := NewDecimal(1, 0)

	for i := int64(1); ; {
		step := pow.DivRound(factorial, divPrecision)
		result = result.Add(step)

		// Stop Taylor series when current step is smaller than epsilon
		if step.Cmp(epsilon) < 0 {
			break
		}

		pow = pow.Mul(decAbs)

		i++

		// Calculate next factorial number or retrieve cached value
		if len(factorials) >= int(i) && !factorials[i-1].IsZero() {
			factorial = factorials[i-1]
		} else {
			// To avoid any race conditions, firstly the zero value is appended to a slice to create
			// a spot for newly calculated factorial. After that, the zero value is replaced by calculated
			// factorial using the index notation.
			factorial = factorials[i-2].Mul(NewDecimal(i, 0))
			factorials = append(factorials, Zero)
			factorials[i-1] = factorial
		}
	}

	if d.Sign() < 0 {
		result = NewDecimal(1, 0).DivRound(result, precision+1)
	}

	result = result.Round(precision)

	return result, nil
}

// Cmp compares the numbers represented by d and d2 and returns:
//
//	-1 if d <  d2
//	 0 if d == d2
//	+1 if d >  d2
func (d Decimal) Cmp(d2 Decimal) int {
	d.ensureInitialized()
	d2.ensureInitialized()

	if d.exp == d2.exp {
		return d.value.Cmp(d2.value)
	}

	baseExp := min(d.exp, d2.exp)
	rd := d.rescale(baseExp)
	rd2 := d2.rescale(baseExp)

	return rd.value.Cmp(rd2.value)
}

// Equal returns whether the numbers represented by d and d2 are equal.
func (d Decimal) Equal(d2 Decimal) bool {
	return d.Cmp(d2) == 0
}

// Greater Than (GT) returns true when d is greater than d2.
func (d Decimal) GreaterThan(d2 Decimal) bool {
	return d.Cmp(d2) == 1
}

// Greater Than or Equal (GTE) returns true when d is greater than or equal to d2.
func (d Decimal) GreaterThanOrEqual(d2 Decimal) bool {
	cmp := d.Cmp(d2)

	return cmp == 1 || cmp == 0
}

// Less Than (LT) returns true when d is less than d2.
func (d Decimal) LessThan(d2 Decimal) bool {
	return d.Cmp(d2) == -1
}

// Less Than or Equal (LTE) returns true when d is less than or equal to d2.
func (d Decimal) LessThanOrEqual(d2 Decimal) bool {
	cmp := d.Cmp(d2)

	return cmp == -1 || cmp == 0
}

// Sign returns:
//
// -1 if d <  0
//
//	0 if d == 0
//
// +1 if d >  0
// .
func (d Decimal) Sign() int {
	if d.value == nil {
		return 0
	}

	return d.value.Sign()
}

// Exponent returns the exponent, or scale component of the decimal.
func (d Decimal) Exponent() int32 {
	return d.exp
}

// IntPart returns the integer component of the decimal.
func (d Decimal) IntPart() int64 {
	scaledD := d.rescale(0)

	return scaledD.value.Int64()
}

// Rat returns a rational number representation of the decimal.
func (d Decimal) Rat() *big.Rat {
	d.ensureInitialized()

	if d.exp <= 0 {
		// NOTE(vadim): must negate after casting to prevent int32 overflow
		denom := new(big.Int).Exp(tenInt, big.NewInt(-int64(d.exp)), nil)

		return new(big.Rat).SetFrac(d.value, denom)
	}

	mul := new(big.Int).Exp(tenInt, big.NewInt(int64(d.exp)), nil)
	num := new(big.Int).Mul(d.value, mul)

	return new(big.Rat).SetFrac(num, oneInt)
}

// Float64 returns the nearest float64 value for d and a bool indicating
// whether f represents d exactly.
// For more details, see the documentation for big.Rat.Float64.
func (d Decimal) Float64() (float64, bool) {
	d.ensureInitialized()

	return d.Rat().Float64()
}

func RequireFromString(value string) Decimal {
	dec, err := NewDecimalFromString(value)
	if err != nil {
		panic(err)
	}

	return dec
}

// String returns the string representation of the decimal
// with the fixed point.
//
// Example:
//
//	d := New(-12345, -3)
//	println(d.String())
//
// Output:
//
//	-12.345
func (d Decimal) String() string {
	return d.string(true)
}

// StringFixed returns a rounded fixed-point string with places digits after
// the decimal point.
//
// Example:
//
//	NewDecimalFromFloat(0).StringFixed(2) // output: "0.00"
//	NewDecimalFromFloat(0).StringFixed(0) // output: "0"
//	NewDecimalFromFloat(5.45).StringFixed(0) // output: "5"
//	NewDecimalFromFloat(5.45).StringFixed(1) // output: "5.5"
//	NewDecimalFromFloat(5.45).StringFixed(2) // output: "5.45"
//	NewDecimalFromFloat(5.45).StringFixed(3) // output: "5.450"
//	NewDecimalFromFloat(545).StringFixed(-1) // output: "550"
func (d Decimal) StringFixed(places int32) string {
	rounded := d.Round(places)

	return rounded.string(false)
}

// Round rounds the decimal to places decimal places.
// If places < 0, it will round the integer part to the nearest 10^(-places).
//
// Example:
//
//	NewDecimalFromFloat(5.45).Round(1).String() // output: "5.5"
//	NewDecimalFromFloat(545).Round(-1).String() // output: "550"
func (d Decimal) Round(places int32) Decimal {
	d.ensureInitialized()

	if d.exp == -places {
		return d
	}
	// truncate to places + 1
	ret := d.rescale(-places - 1)

	// add sign(d) * 0.5
	if ret.value.Sign() < 0 {
		ret.value.Sub(ret.value, fiveInt)
	} else {
		ret.value.Add(ret.value, fiveInt)
	}

	// floor for positive numbers, ceil for negative numbers
	_, m := ret.value.DivMod(ret.value, tenInt, new(big.Int))
	ret.exp++

	if ret.value.Sign() < 0 && m.Cmp(zeroInt) != 0 {
		ret.value.Add(ret.value, oneInt)
	}

	return ret
}

func (d Decimal) RoundUp(places int32) Decimal {
	if d.exp >= -places {
		return d
	}

	rescaled := d.rescale(-places)
	if d.Equal(rescaled) {
		return d
	}

	if d.value.Sign() > 0 {
		rescaled.value.Add(rescaled.value, oneInt)
	} else if d.value.Sign() < 0 {
		rescaled.value.Sub(rescaled.value, oneInt)
	}

	return rescaled
}

func (d Decimal) RoundDown(places int32) Decimal {
	if d.exp >= -places {
		return d
	}

	rescaled := d.rescale(-places)
	if d.Equal(rescaled) {
		return d
	}

	return rescaled
}

func (d Decimal) RoundBank(places int32) Decimal {
	round := d.Round(places)
	remainder := d.Sub(round).Abs()

	half := NewDecimal(5, -places-1)
	if remainder.Cmp(half) == 0 && round.value.Bit(0) != 0 {
		if round.value.Sign() < 0 {
			round.value.Add(round.value, oneInt)
		} else {
			round.value.Sub(round.value, oneInt)
		}
	}

	return round
}

// Floor returns the nearest integer value less than or equal to d.
func (d Decimal) Floor() Decimal {
	d.ensureInitialized()

	if d.exp >= 0 {
		return d
	}

	exp := big.NewInt(10)

	exp.Exp(exp, big.NewInt(-int64(d.exp)), nil)

	z := new(big.Int).Div(d.value, exp)

	return Decimal{value: z, exp: 0}
}

// Ceil returns the nearest integer value greater than or equal to d.
func (d Decimal) Ceil() Decimal {
	d.ensureInitialized()

	if d.exp >= 0 {
		return d
	}

	exp := big.NewInt(10)

	exp.Exp(exp, big.NewInt(-int64(d.exp)), nil)

	z, m := new(big.Int).DivMod(d.value, exp, new(big.Int))
	if m.Cmp(zeroInt) != 0 {
		z.Add(z, oneInt)
	}

	return Decimal{value: z, exp: 0}
}

// Truncate truncates off digits from the number, without rounding.
//
// NOTE: precision is the last digit that will not be truncated (must be >= 0).
//
// Example:
//
//	decimal.NewDecimalFromString("123.456").Truncate(2).String() // "123.45"
func (d Decimal) Truncate(precision int32) Decimal {
	d.ensureInitialized()

	if precision >= 0 && -precision > d.exp {
		return d.rescale(-precision)
	}

	return d
}

func (d Decimal) IsInteger() bool {
	// The most typical case, all decimal with exponent higher or equal 0 can be represented as integer
	if d.exp >= 0 {
		return true
	}
	// When the exponent is negative we have to check every number after the decimal place
	// If all of them are zeroes, we are sure that given decimal can be represented as an integer
	var r big.Int

	q := new(big.Int).Set(d.value)

	for z := abs(int64(d.exp)); z > 0; z-- {
		q.QuoRem(q, tenInt, &r)

		if r.Cmp(zeroInt) != 0 {
			return false
		}
	}

	return true
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (d *Decimal) UnmarshalJSON(decimalBytes []byte) error {
	str, err := unquoteIfQuoted(decimalBytes)
	if err != nil {
		return fmt.Errorf("error decoding string '%s': %s", decimalBytes, err)
	}

	decimal, err := NewDecimalFromString(str)
	if err != nil {
		return fmt.Errorf("error decoding string '%s': %s", str, err)
	}

	*d = decimal

	return nil
}

// For mail.ru easy json.
func (d Decimal) MarshalEasyJSON(w *jwriter.Writer) {
	d.ensureInitialized()

	if MarshalJSONWithoutQuotes {
		w.RawString(d.String())
	} else {
		w.RawString("\"")
		w.RawString(d.String())
		w.RawString("\"")
	}
}

func (d Decimal) IsDefined() bool {
	if d.value == nil {
		return false
	}

	if d.value.Sign() == 0 {
		return false
	}

	return true
}

func (d *Decimal) UnmarshalEasyJSON(l *jlexer.Lexer) {
	if l.IsNull() {
		l.Skip()
	} else {
		*d, _ = NewDecimalFromString(string(l.JSONNumber()))
	}
}

// MarshalJSON implements the json.Marshaler interface.
func (d Decimal) MarshalJSON() ([]byte, error) {
	d.ensureInitialized()

	var str string
	if MarshalJSONWithoutQuotes {
		str = d.String()
	} else {
		str = "\"" + d.String() + "\""
	}

	return []byte(str), nil
}

func (d Decimal) BigInt() *big.Int {
	scaledD := d.rescale(0)

	return scaledD.value
}

func (d Decimal) Copy() Decimal {
	d.ensureInitialized()

	return Decimal{
		value: new(big.Int).Set(d.value),
		exp:   d.exp,
	}
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface. As a string representation
// is already used when encoding to text, this method stores that string as []byte.
func (d *Decimal) UnmarshalBinary(data []byte) error {
	// Extract the exponent
	d.exp = int32(binary.BigEndian.Uint32(data[:4])) //nolint:gosec

	// Extract the value
	d.value = new(big.Int)

	if err := d.value.GobDecode(data[4:]); err != nil {
		return fmt.Errorf("unmarshal binary error: %w", err)
	}

	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (d Decimal) MarshalBinary() ([]byte, error) {
	d.ensureInitialized()

	// Write the exponent first since it's a fixed size
	v1 := make([]byte, 4)

	binary.BigEndian.PutUint32(v1, uint32(d.exp)) //nolint:gosec

	// Add the value
	var v2 []byte

	var err error

	if v2, err = d.value.GobEncode(); err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}

	//nolint: makezero
	v1 = append(v1, v2...)

	return v1, nil
}

func (d Decimal) ToMoney(currency string) *money.Money {
	u, n := d.ToMoneyParts()

	return &money.Money{
		CurrencyCode: currency,
		Units:        u,
		Nanos:        n,
	}
}

// UnmarshalText implements the encoding.TextUnmarshaler interface for XML
// deserialization.
func (d *Decimal) UnmarshalText(text []byte) error {
	d.ensureInitialized()

	str := string(text)

	dec, err := NewDecimalFromString(str)
	if err != nil {
		return fmt.Errorf("error decoding string '%s': %s", str, err)
	}

	*d = dec

	return nil
}

func (d Decimal) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	e.EncodeToken(start)
	e.EncodeToken(xml.CharData(d.String()))
	e.EncodeToken(xml.EndElement{Name: start.Name})

	return nil
}

func (d OptDecimal) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if d.Defined {
		e.EncodeToken(start)
		e.EncodeToken(xml.CharData(d.V.String()))
		e.EncodeToken(xml.EndElement{Name: start.Name})
	}

	return nil
}

func (d Decimal) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return xml.Attr{Name: name, Value: d.String()}, nil
}

func (d OptDecimal) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	if d.Defined {
		return xml.Attr{Name: name, Value: d.V.String()}, nil
	}

	return xml.Attr{}, nil
}

func (d *Decimal) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	var s string

	err := dec.DecodeElement(&s, &start)

	if err != nil {
		return err
	}

	t, err := NewDecimalFromString(s)
	if err != nil {
		return err
	}

	*d = t

	return nil
}

func (d *Decimal) UnmarshalXMLAttr(attr xml.Attr) error {
	v, err := NewDecimalFromString(attr.Value)
	if err != nil {
		return err
	}

	*d = v

	return nil
}

func (d *OptDecimal) UnmarshalXMLAttr(attr xml.Attr) error {
	if attr.Value == "" {
		d.Undefine()

		return nil
	}

	v, err := NewDecimalFromString(attr.Value)
	if err != nil {
		return err
	}

	d.V, d.Defined = v, true

	return nil
}

func (d Decimal) LogValue() slog.Value {
	return slog.StringValue(d.String())
}

// MarshalText implements the encoding.TextMarshaler interface for XML
// serialization.
func (d Decimal) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

// GobEncode implements the gob.GobEncoder interface for gob serialization.
func (d Decimal) GobEncode() ([]byte, error) {
	return d.MarshalBinary()
}

// GobDecode implements the gob.GobDecoder interface for gob serialization.
func (d *Decimal) GobDecode(data []byte) error {
	return d.UnmarshalBinary(data)
}

// StringScaled first scales the decimal then calls .String() on it.
func (d Decimal) StringScaled(exp int32) string {
	return d.rescale(exp).String()
}

func (d Decimal) StringFixedBank(places int32) string {
	return d.RoundBank(places).string(false)
}

func (d Decimal) string(trimTrailingZeros bool) string {
	d.ensureInitialized()

	if d.exp >= 0 {
		return d.rescale(0).value.String()
	}

	var str string

	if d.Sign() >= 0 {
		if d.value.IsInt64() {
			str = strconv.FormatInt(d.value.Int64(), 10)
		} else {
			str = d.value.String()
		}
	} else {
		abs := new(big.Int).Abs(d.value)
		if abs.IsInt64() {
			str = strconv.FormatInt(abs.Int64(), 10)
		} else {
			str = abs.String()
		}
	}

	var intPart, fractionalPart string

	dExpInt := int(d.exp)
	if len(str) > -dExpInt {
		intPart = str[:len(str)+dExpInt]
		fractionalPart = str[len(str)+dExpInt:]
	} else {
		intPart = "0"

		num0s := -dExpInt - len(str)
		fractionalPart = strings.Repeat("0", num0s) + str
	}

	if trimTrailingZeros {
		i := len(fractionalPart) - 1
		for ; i >= 0; i-- {
			if fractionalPart[i] != '0' {
				break
			}
		}

		fractionalPart = fractionalPart[:i+1]
	}

	buff := strings.Builder{}
	buff.Grow(len(intPart) + len(fractionalPart) + 2)

	if d.value.Sign() < 0 {
		buff.WriteByte('-')
	}

	buff.WriteString(intPart)

	if len(fractionalPart) > 0 {
		buff.WriteByte('.')
		buff.WriteString(fractionalPart)
	}

	return buff.String()
}

func (d *Decimal) ensureInitialized() {
	if d.value == nil {
		d.value = new(big.Int)
	}
}

func (d Decimal) Coefficient() *big.Int {
	d.ensureInitialized()

	return new(big.Int).Set(d.value)
}

func (d Decimal) InexactFloat64() float64 {
	f, _ := d.Float64()

	return f
}

// Min returns the smallest Decimal that was passed in the arguments.
//
// To call this function with an array, you must do:
//
//	Min(arr[0], arr[1:]...)
//
// This makes it harder to accidentally call Min with 0 arguments.
func Min(first Decimal, rest ...Decimal) Decimal {
	ans := first

	for _, item := range rest {
		if item.Cmp(ans) < 0 {
			ans = item
		}
	}

	return ans
}

// Max returns the largest Decimal that was passed in the arguments.
//
// To call this function with an array, you must do:
//
//	Max(arr[0], arr[1:]...)
//
// This makes it harder to accidentally call Max with 0 arguments.
func Max(first Decimal, rest ...Decimal) Decimal {
	ans := first

	for _, item := range rest {
		if item.Cmp(ans) > 0 {
			ans = item
		}
	}

	return ans
}

func (d *Decimal) Scan(value any) error {
	// first try to see if the data is stored in database as a Numeric datatype
	switch v := value.(type) {
	case float32:
		*d = NewDecimalFromFloat(float64(v))

		return nil

	case float64:
		// numeric in sqlite3 sends us float64
		*d = NewDecimalFromFloat(v)

		return nil

	case int64:
		// at least in sqlite3 when the value is 0 in db, the data is sent
		// to us as an int64 instead of a float64 ...
		*d = NewDecimal(v, 0)

		return nil

	case uint64:
		*d = NewDecimalFromUint(v)

		return nil

	default:
		// default is trying to interpret value stored as string
		str, err := unquoteIfQuoted(v)
		if err != nil {
			return err
		}

		*d, err = NewDecimalFromString(str)

		return err
	}
}

// Value implements the driver.Valuer interface for database serialization.
func (d Decimal) Value() (driver.Value, error) {
	return d.String(), nil
}

func (d Decimal) StringAFT() string {
	str := d.StringFixedBank(4)

	dotIdx := strings.IndexByte(str, '.')

	if dotIdx == -1 {
		return str + ".00"
	}

	if (len(str)-1)-dotIdx == 0 {
		return str + "00"
	}

	if (len(str)-1)-dotIdx == 1 {
		return str + "0"
	}

	hasLastZero := str[len(str)-1] == '0'

	if hasLastZero {
		str = str[:len(str)-1]
	}

	if hasLastZero && str[len(str)-1] == '0' {
		str = str[:len(str)-1]
	}

	return str
}

type OptDecimal struct {
	V       Decimal
	Defined bool
}

func (d *OptDecimal) SetValue(val Decimal) {
	d.V, d.Defined = val, true
}

func (d *OptDecimal) Undefine() {
	d.V, d.Defined = NewDecimal(0, 0), false
}

func (d OptDecimal) MarshalEasyJSON(w *jwriter.Writer) {
	if d.Defined {
		d.V.MarshalEasyJSON(w)
	} else {
		w.RawString("null")
	}
}

func (d OptDecimal) String() string {
	if d.Defined {
		return d.V.String()
	}

	return undef
}

func (d OptDecimal) Equal(d2 OptDecimal) bool {
	if !d.Defined && !d2.Defined {
		return true
	}

	if d.Defined && d2.Defined {
		return d.V.Cmp(d2.V) == 0
	}

	return false
}

func (d *OptDecimal) UnmarshalEasyJSON(l *jlexer.Lexer) {
	if l.IsNull() {
		l.Skip()

		*d = OptDecimal{NewDecimal(0, 0), false}
	} else {
		v, _ := NewDecimalFromString(string(l.JSONNumber()))

		*d = OptDecimal{v, true}
	}
}

func (d OptDecimal) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	d.MarshalEasyJSON(&w)

	return w.Buffer.BuildBytes(), w.Error
}

func (d OptDecimal) ToMoney(currency string) *money.Money {
	if d.IsDefined() {
		return nil
	}

	return d.V.ToMoney(currency)
}

func (d *OptDecimal) UnmarshalText(text []byte) error {
	if str := string(text); str == "" {
		d.Defined = false

		return nil
	}

	if err := d.V.UnmarshalText(text); err != nil {
		d.Defined = false

		return err
	}

	d.Defined = true

	return nil
}

func (d OptDecimal) MarshalText() ([]byte, error) {
	if !d.Defined {
		return []byte{}, nil
	}

	return d.V.MarshalText()
}

func (d *OptDecimal) Scan(value any) error {
	if value == nil {
		d.Defined = false

		return nil
	}

	d.Defined = true

	return d.V.Scan(value)
}

func (d OptDecimal) IsDefined() bool {
	return d.Defined
}

// Value implements the driver.Valuer interface for database serialization.
func (d OptDecimal) Value() (driver.Value, error) {
	if !d.Defined {
		return nil, nil
	}

	return d.V.Value()
}

func (d OptDecimal) LogValue() slog.Value {
	if d.Defined {
		return slog.StringValue(d.V.String())
	}

	return slog.AnyValue(nil)
}

type decimal struct {
	d     [800]byte // digits, big-endian representation
	nd    int       // number of digits used
	dp    int       // decimal point
	neg   bool      // negative flag
	trunc bool      // discarded nonzero digits beyond d[:nd]
}

func (a *decimal) String() string {
	n := 10 + a.nd
	if a.dp > 0 {
		n += a.dp
	}

	if a.dp < 0 {
		n += -a.dp
	}

	buf := make([]byte, n)
	w := 0

	switch {
	case a.nd == 0:
		return "0"

	case a.dp <= 0:
		// zeros fill space between decimal point and digits
		buf[w] = '0'
		w++

		buf[w] = '.'

		w++
		w += digitZero(buf[w : w+-a.dp])
		w += copy(buf[w:], a.d[0:a.nd])

	case a.dp < a.nd:
		// decimal point in middle of digits
		w += copy(buf[w:], a.d[0:a.dp])
		buf[w] = '.'

		w++
		w += copy(buf[w:], a.d[a.dp:a.nd])

	default:
		// zeros fill space between digits and decimal point
		w += copy(buf[w:], a.d[0:a.nd])
		w += digitZero(buf[w : w+a.dp-a.nd])
	}

	return string(buf[0:w])
}

func digitZero(dst []byte) int {
	for i := range dst {
		dst[i] = '0'
	}

	return len(dst)
}

// trim trailing zeros from number.
// (They are meaningless; the decimal point is tracked
// independent of the number of digits.)
func trim(a *decimal) {
	for a.nd > 0 && a.d[a.nd-1] == '0' {
		a.nd--
	}

	if a.nd == 0 {
		a.dp = 0
	}
}

// Assign v to a.
func (a *decimal) Assign(v uint64) {
	var buf [24]byte

	// Write reversed decimal in buf.
	n := 0

	for v > 0 {
		v1 := v / 10
		v -= 10 * v1

		buf[n] = byte(v + '0')

		n++

		v = v1
	}

	// Reverse again to produce forward decimal in a.d.
	a.nd = 0
	for n--; n >= 0; n-- {
		a.d[a.nd] = buf[n]
		a.nd++
	}

	a.dp = a.nd
	trim(a)
}

// Maximum shift that we can do in one pass without overflow.
// A uint has 32 or 64 bits, and we have to be able to accommodate 9<<k.
const uintSize = 32 << (^uint(0) >> 63)
const maxShift = uintSize - 4

// Binary shift right (/ 2) by k bits.  k <= maxShift to avoid overflow.
func rightShift(a *decimal, k uint) {
	r := 0 // read pointer
	w := 0 // write pointer

	// Pick up enough leading digits to cover first shift.
	var n uint

	for ; n>>k == 0; r++ {
		if r >= a.nd {
			if n == 0 {
				// a == 0; shouldn't get here, but handle anyway.
				a.nd = 0

				return
			}

			for n>>k == 0 {
				n = n * 10
				r++
			}

			break
		}

		c := uint(a.d[r])

		n = n*10 + c - '0'
	}

	a.dp -= r - 1

	var mask uint = (1 << k) - 1

	// Pick up a digit, put down a digit.
	for ; r < a.nd; r++ {
		c := uint(a.d[r])
		dig := n >> k

		n &= mask
		a.d[w] = byte(dig + '0')

		w++

		n = n*10 + c - '0'
	}

	// Put down extra digits.
	for n > 0 {
		dig := n >> k
		n &= mask

		if w < len(a.d) {
			a.d[w] = byte(dig + '0')
			w++
		} else if dig > 0 {
			a.trunc = true
		}

		n = n * 10
	}

	a.nd = w
	trim(a)
}

// Cheat sheet for left shift: table indexed by shift count giving
// number of new digits that will be introduced by that shift.
//
// For example, leftcheats[4] = {2, "625"}.  That means that
// if we are shifting by 4 (multiplying by 16), it will add 2 digits
// when the string prefix is "625" through "999", and one fewer digit
// if the string prefix is "000" through "624".
//
// Credit for this trick goes to Ken.

// //nolint:govet
type leftCheat struct {
	delta  int    // number of new digits
	cutoff string // minus one digit if original < a.
}

var leftcheats = []leftCheat{
	// Leading digits of 1/2^i = 5^i.
	// 5^23 is not an exact 64-bit floating point number,
	// so have to use bc for the math.
	// Go up to 60 to be large enough for 32bit and 64bit platforms.
	/*
		seq 60 | sed 's/^/5^/' | bc |
		awk 'BEGIN{ print "\t{ 0, \"\" }," }
		{
			log2 = log(2)/log(10)
			printf("\t{ %d, \"%s\" },\t// * %d\n",
				int(log2*NR+1), $0, 2**NR)
		}'
	*/
	{0, ""},
	{1, "5"},                                           // * 2
	{1, "25"},                                          // * 4
	{1, "125"},                                         // * 8
	{2, "625"},                                         // * 16
	{2, "3125"},                                        // * 32
	{2, "15625"},                                       // * 64
	{3, "78125"},                                       // * 128
	{3, "390625"},                                      // * 256
	{3, "1953125"},                                     // * 512
	{4, "9765625"},                                     // * 1024
	{4, "48828125"},                                    // * 2048
	{4, "244140625"},                                   // * 4096
	{4, "1220703125"},                                  // * 8192
	{5, "6103515625"},                                  // * 16384
	{5, "30517578125"},                                 // * 32768
	{5, "152587890625"},                                // * 65536
	{6, "762939453125"},                                // * 131072
	{6, "3814697265625"},                               // * 262144
	{6, "19073486328125"},                              // * 524288
	{7, "95367431640625"},                              // * 1048576
	{7, "476837158203125"},                             // * 2097152
	{7, "2384185791015625"},                            // * 4194304
	{7, "11920928955078125"},                           // * 8388608
	{8, "59604644775390625"},                           // * 16777216
	{8, "298023223876953125"},                          // * 33554432
	{8, "1490116119384765625"},                         // * 67108864
	{9, "7450580596923828125"},                         // * 134217728
	{9, "37252902984619140625"},                        // * 268435456
	{9, "186264514923095703125"},                       // * 536870912
	{10, "931322574615478515625"},                      // * 1073741824
	{10, "4656612873077392578125"},                     // * 2147483648
	{10, "23283064365386962890625"},                    // * 4294967296
	{10, "116415321826934814453125"},                   // * 8589934592
	{11, "582076609134674072265625"},                   // * 17179869184
	{11, "2910383045673370361328125"},                  // * 34359738368
	{11, "14551915228366851806640625"},                 // * 68719476736
	{12, "72759576141834259033203125"},                 // * 137438953472
	{12, "363797880709171295166015625"},                // * 274877906944
	{12, "1818989403545856475830078125"},               // * 549755813888
	{13, "9094947017729282379150390625"},               // * 1099511627776
	{13, "45474735088646411895751953125"},              // * 2199023255552
	{13, "227373675443232059478759765625"},             // * 4398046511104
	{13, "1136868377216160297393798828125"},            // * 8796093022208
	{14, "5684341886080801486968994140625"},            // * 17592186044416
	{14, "28421709430404007434844970703125"},           // * 35184372088832
	{14, "142108547152020037174224853515625"},          // * 70368744177664
	{15, "710542735760100185871124267578125"},          // * 140737488355328
	{15, "3552713678800500929355621337890625"},         // * 281474976710656
	{15, "17763568394002504646778106689453125"},        // * 562949953421312
	{16, "88817841970012523233890533447265625"},        // * 1125899906842624
	{16, "444089209850062616169452667236328125"},       // * 2251799813685248
	{16, "2220446049250313080847263336181640625"},      // * 4503599627370496
	{16, "11102230246251565404236316680908203125"},     // * 9007199254740992
	{17, "55511151231257827021181583404541015625"},     // * 18014398509481984
	{17, "277555756156289135105907917022705078125"},    // * 36028797018963968
	{17, "1387778780781445675529539585113525390625"},   // * 72057594037927936
	{18, "6938893903907228377647697925567626953125"},   // * 144115188075855872
	{18, "34694469519536141888238489627838134765625"},  // * 288230376151711744
	{18, "173472347597680709441192448139190673828125"}, // * 576460752303423488
	{19, "867361737988403547205962240695953369140625"}, // * 1152921504606846976
}

// Is the leading prefix of b lexicographically less than s?
func prefixIsLessThan(b []byte, s string) bool {
	for i := range len(s) {
		if i >= len(b) {
			return true
		}

		if b[i] != s[i] {
			return b[i] < s[i]
		}
	}

	return false
}

// Binary shift left (* 2) by k bits.  k <= maxShift to avoid overflow.
func leftShift(a *decimal, k uint) {
	delta := leftcheats[k].delta
	if prefixIsLessThan(a.d[0:a.nd], leftcheats[k].cutoff) {
		delta--
	}

	r := a.nd         // read index
	w := a.nd + delta // write index

	// Pick up a digit, put down a digit.
	var n uint
	for r--; r >= 0; r-- {
		n += (uint(a.d[r]) - '0') << k

		quo := n / 10
		rem := n - 10*quo

		w--
		if w < len(a.d) {
			a.d[w] = byte(rem + '0')
		} else if rem != 0 {
			a.trunc = true
		}

		n = quo
	}

	// Put down extra digits.
	for n > 0 {
		quo := n / 10
		rem := n - 10*quo

		w--
		if w < len(a.d) {
			a.d[w] = byte(rem + '0')
		} else if rem != 0 {
			a.trunc = true
		}

		n = quo
	}

	a.nd += delta
	if a.nd >= len(a.d) {
		a.nd = len(a.d)
	}

	a.dp += delta
	trim(a)
}

// Binary shift left (k > 0) or right (k < 0).
func (a *decimal) Shift(k int) {
	switch {
	case a.nd == 0:
		// nothing to do: a == 0
	case k > 0:
		for k > maxShift {
			leftShift(a, maxShift)
			k -= maxShift
		}

		leftShift(a, uint(k)) //nolint:gosec
	case k < 0:
		for k < -maxShift {
			rightShift(a, maxShift)
			k += maxShift
		}

		rightShift(a, uint(-k)) //nolint:gosec
	}
}

// If we chop a at nd digits, should we round up?
func shouldRoundUp(a *decimal, nd int) bool {
	if nd < 0 || nd >= a.nd {
		return false
	}

	if a.d[nd] == '5' && nd+1 == a.nd { // exactly halfway - round to even
		// if we truncated, a little higher than what's recorded - always round up
		if a.trunc {
			return true
		}

		return nd > 0 && (a.d[nd-1]-'0')%2 != 0
	}
	// not halfway - digit tells all
	return a.d[nd] >= '5'
}

// Round a to nd digits (or fewer).
// If nd is zero, it means we're rounding
// just to the left of the digits, as in
// 0.09 -> 0.1.
func (a *decimal) Round(nd int) {
	if nd < 0 || nd >= a.nd {
		return
	}

	if shouldRoundUp(a, nd) {
		a.RoundUp(nd)
	} else {
		a.RoundDown(nd)
	}
}

// Round a down to nd digits (or fewer).
func (a *decimal) RoundDown(nd int) {
	if nd < 0 || nd >= a.nd {
		return
	}

	a.nd = nd
	trim(a)
}

// RoundUp a up to nd digits (or fewer).
func (a *decimal) RoundUp(nd int) {
	if nd < 0 || nd >= a.nd {
		return
	}

	// round up
	for i := nd - 1; i >= 0; i-- {
		c := a.d[i]
		if c < '9' { // can stop after this digit
			a.d[i]++
			a.nd = i + 1

			return
		}
	}

	// Number is all 9s.
	// Change to single 1 with adjusted decimal point.
	a.d[0] = '1'
	a.nd = 1
	a.dp++
}

func (d Decimal) RoundCeil(places int32) Decimal {
	if d.exp >= -places {
		return d
	}

	rescaled := d.rescale(-places)
	if d.Equal(rescaled) {
		return d
	}

	if d.value.Sign() > 0 {
		rescaled.value.Add(rescaled.value, oneInt)
	}

	return rescaled
}

func (d Decimal) RoundFloor(places int32) Decimal {
	if d.exp >= -places {
		return d
	}

	rescaled := d.rescale(-places)
	if d.Equal(rescaled) {
		return d
	}

	if d.value.Sign() < 0 {
		rescaled.value.Sub(rescaled.value, oneInt)
	}

	return rescaled
}

// Extract integer part, rounded appropriately.
// No guarantees about overflow.
func (a *decimal) RoundedInteger() uint64 {
	if a.dp > 20 {
		return 0xFFFFFFFFFFFFFFFF
	}

	var i int

	n := uint64(0)
	for i = 0; i < a.dp && i < a.nd; i++ {
		n = n*10 + uint64(a.d[i]-'0')
	}

	for ; i < a.dp; i++ {
		n *= 10
	}

	if shouldRoundUp(a, a.dp) {
		n++
	}

	return n
}

type floatInfo struct {
	mantbits uint
	expbits  uint
	bias     int
}

var (
	float32info = floatInfo{23, 8, -127}
	float64info = floatInfo{52, 11, -1023}
)

// roundShortest rounds d (= mant * 2^exp) to the shortest number of digits
// that will let the original floating point value be precisely reconstructed.
func roundShortest(d *decimal, mant uint64, exp int, flt *floatInfo) {
	// If mantissa is zero, the number is zero; stop now.
	if mant == 0 {
		d.nd = 0

		return
	}

	// Compute upper and lower such that any decimal number
	// between upper and lower (possibly inclusive)
	// will round to the original floating point number.

	// We may see at once that the number is already shortest.
	//
	// Suppose d is not denormal, so that 2^exp <= d < 10^dp.
	// The closest shorter number is at least 10^(dp-nd) away.
	// The lower/upper bounds computed below are at distance
	// at most 2^(exp-mantbits).
	//
	// So the number is already shortest if 10^(dp-nd) > 2^(exp-mantbits),
	// or equivalently log2(10)*(dp-nd) > exp-mantbits.
	// It is true if 332/100*(dp-nd) >= exp-mantbits (log2(10) > 3.32).
	minexp := flt.bias + 1                                              // minimum possible exponent
	if exp > minexp && 332*(d.dp-d.nd) >= 100*(exp-int(flt.mantbits)) { //nolint:gosec
		// The number is already shortest.
		return
	}

	// d = mant << (exp - mantbits)
	// Next highest floating point number is mant+1 << exp-mantbits.
	// Our upper bound is halfway between, mant*2+1 << exp-mantbits-1.
	upper := new(decimal)

	upper.Assign(mant*2 + 1)
	upper.Shift(exp - int(flt.mantbits) - 1) //nolint:gosec

	// d = mant << (exp - mantbits)
	// Next lowest floating point number is mant-1 << exp-mantbits,
	// unless mant-1 drops the significant bit and exp is not the minimum exp,
	// in which case the next lowest is mant*2-1 << exp-mantbits-1.
	// Either way, call it mantlo << explo-mantbits.
	// Our lower bound is halfway between, mantlo*2+1 << explo-mantbits-1.
	var mantlo uint64

	var explo int

	if mant > 1<<flt.mantbits || exp == minexp {
		mantlo = mant - 1
		explo = exp
	} else {
		mantlo = mant*2 - 1
		explo = exp - 1
	}

	lower := new(decimal)

	lower.Assign(mantlo*2 + 1)
	lower.Shift(explo - int(flt.mantbits) - 1) //nolint:gosec

	// The upper and lower bounds are possible outputs only if
	// the original mantissa is even, so that IEEE round-to-even
	// would round to the original mantissa and not the neighbors.
	inclusive := mant%2 == 0

	// As we walk the digits we want to know whether rounding up would fall
	// within the upper bound. This is tracked by upperdelta:
	//
	// If upperdelta == 0, the digits of d and upper are the same so far.
	//
	// If upperdelta == 1, we saw a difference of 1 between d and upper on a
	// previous digit and subsequently only 9s for d and 0s for upper.
	// (Thus rounding up may fall outside the bound, if it is exclusive.)
	//
	// If upperdelta == 2, then the difference is greater than 1
	// and we know that rounding up falls within the bound.
	var upperdelta uint8

	// Now we can figure out the minimum number of digits required.
	// Walk along until d has distinguished itself from upper and lower.
	for ui := 0; ; ui++ {
		// lower, d, and upper may have the decimal points at different
		// places. In this case upper is the longest, so we iterate from
		// ui==0 and start li and mi at (possibly) -1.
		mi := ui - upper.dp + d.dp
		if mi >= d.nd {
			break
		}

		l := byte('0') // lower digit

		li := ui - upper.dp + lower.dp
		if li >= 0 && li < lower.nd {
			l = lower.d[li]
		}

		m := byte('0') // middle digit
		if mi >= 0 {
			m = d.d[mi]
		}

		u := byte('0') // upper digit
		if ui < upper.nd {
			u = upper.d[ui]
		}

		// Okay to round down (truncate) if lower has a different digit
		// or if lower is inclusive and is exactly the result of rounding
		// down (i.e., and we have reached the final digit of lower).
		okdown := l != m || inclusive && li+1 == lower.nd

		switch {
		case upperdelta == 0 && m+1 < u:
			// Example:
			// m = 12345xxx
			// u = 12347xxx
			upperdelta = 2
		case upperdelta == 0 && m != u:
			// Example:
			// m = 12345xxx
			// u = 12346xxx
			upperdelta = 1
		case upperdelta == 1 && (m != '9' || u != '0'):
			// Example:
			// m = 1234598x
			// u = 1234600x
			upperdelta = 2
		}
		// Okay to round up if upper has a different digit and either upper
		// is inclusive or upper is bigger than the result of rounding up.
		okup := upperdelta > 0 && (inclusive || upperdelta > 1 || ui+1 < upper.nd)

		// If it's okay to do either, then round to the nearest one.
		// If it's okay to do only one, do it.
		switch {
		case okdown && okup:
			d.Round(mi + 1)

			return
		case okdown:
			d.RoundDown(mi + 1)

			return
		case okup:
			d.RoundUp(mi + 1)

			return
		}
	}
}

// DecimalMax returns maximal decimal out of given values.
func DecimalMax(x Decimal, y Decimal) Decimal {
	if x.GreaterThanOrEqual(y) {
		return x
	}

	return y
}

// DecimalMin returns minimal decimal out of given values.
func DecimalMin(x Decimal, y Decimal) Decimal {
	if x.LessThanOrEqual(y) {
		return x
	}

	return y
}

func (d *Decimal) ScanNumeric(v pgtype.Numeric) error {
	if !v.Valid {
		return errors.New("cannot scan NULL into *decimal.Decimal")
	}

	if v.NaN {
		return errors.New("cannot scan NaN into *decimal.Decimal")
	}

	if v.InfinityModifier != pgtype.Finite {
		return fmt.Errorf("cannot scan %v into *decimal.Decimal", v.InfinityModifier)
	}

	*d = Decimal{
		value: v.Int,
		exp:   v.Exp,
	}

	d.Normalize()

	return nil
}

func (d Decimal) NumericValue() (pgtype.Numeric, error) {
	return pgtype.Numeric{Int: d.Coefficient(), Exp: d.Exponent(), Valid: true}, nil
}

func (d *Decimal) ScanFloat64(v pgtype.Float8) error {
	if !v.Valid {
		return errors.New("cannot scan NULL into *decimal.Decimal")
	}

	if math.IsNaN(v.Float64) {
		return errors.New("cannot scan NaN into *decimal.Decimal")
	}

	if math.IsInf(v.Float64, 0) {
		return fmt.Errorf("cannot scan %v into *decimal.Decimal", v.Float64)
	}

	*d = NewDecimalFromFloat(v.Float64)

	return nil
}

func (d Decimal) Float64Value() (pgtype.Float8, error) {
	return pgtype.Float8{Float64: d.InexactFloat64(), Valid: true}, nil
}

func (d *Decimal) ScanInt64(v pgtype.Int8) error {
	if !v.Valid {
		return errors.New("cannot scan NULL into *decimal.Decimal")
	}

	*d = NewDecimalFromInt(v.Int64)

	return nil
}

func (d Decimal) Int64Value() (pgtype.Int8, error) {
	if !d.IsInteger() {
		return pgtype.Int8{}, fmt.Errorf("cannot convert %v to int64", d)
	}

	bi := d.BigInt()
	if !bi.IsInt64() {
		return pgtype.Int8{}, fmt.Errorf("cannot convert %v to int64", d)
	}

	return pgtype.Int8{Int64: bi.Int64(), Valid: true}, nil
}

func (d *OptDecimal) ScanNumeric(v pgtype.Numeric) error {
	if !v.Valid {
		*d = OptDecimal{}

		return nil
	}

	if v.NaN {
		return errors.New("cannot scan NaN into *decimal.NullDecimal")
	}

	if v.InfinityModifier != pgtype.Finite {
		return fmt.Errorf("cannot scan %v into *decimal.NullDecimal", v.InfinityModifier)
	}

	*d = OptDecimal{V: Decimal{value: v.Int, exp: v.Exp}, Defined: true}

	d.V.Normalize()

	return nil
}

func (d OptDecimal) NumericValue() (pgtype.Numeric, error) {
	if !d.Defined {
		return pgtype.Numeric{}, nil
	}

	return pgtype.Numeric{Int: d.V.Coefficient(), Exp: d.V.Exponent(), Valid: true}, nil
}

func (d *OptDecimal) ScanFloat64(v pgtype.Float8) error {
	if !v.Valid {
		*d = OptDecimal{}

		return nil
	}

	if math.IsNaN(v.Float64) {
		return errors.New("cannot scan NaN into *decimal.NullDecimal")
	}

	if math.IsInf(v.Float64, 0) {
		return fmt.Errorf("cannot scan %v into *decimal.NullDecimal", v.Float64)
	}

	*d = OptDecimal{V: NewDecimalFromFloat(v.Float64), Defined: true}

	return nil
}

func (d OptDecimal) Float64Value() (pgtype.Float8, error) {
	if !d.Defined {
		return pgtype.Float8{}, nil
	}

	return pgtype.Float8{Float64: d.V.InexactFloat64(), Valid: true}, nil
}

func (d *OptDecimal) ScanInt64(v pgtype.Int8) error {
	if !v.Valid {
		*d = OptDecimal{}

		return nil
	}

	*d = OptDecimal{V: NewDecimalFromInt(v.Int64), Defined: true}

	return nil
}

func (d OptDecimal) Int64Value() (pgtype.Int8, error) {
	if !d.Defined {
		return pgtype.Int8{}, nil
	}

	if !d.V.IsInteger() {
		return pgtype.Int8{}, fmt.Errorf("cannot convert %v to int64", d)
	}

	bi := d.V.BigInt()
	if !bi.IsInt64() {
		return pgtype.Int8{}, fmt.Errorf("cannot convert %v to int64", d)
	}

	return pgtype.Int8{Int64: bi.Int64(), Valid: true}, nil
}

type NumericCodec struct {
	pgtype.NumericCodec
}

func (NumericCodec) DecodeValue(tm *pgtype.Map, oid uint32, format int16, src []byte) (any, error) {
	if src == nil {
		//nolint: nilnil
		return nil, nil
	}

	var target Decimal

	scanPlan := tm.PlanScan(oid, format, &target)
	if scanPlan == nil {
		return nil, errors.New("PlanScan did not find a plan")
	}

	err := scanPlan.Scan(src, &target)
	if err != nil {
		return nil, err
	}

	return target, nil
}

// Register registers the shopspring/decimal integration with a pgtype.ConnInfo.
func RegisterDecimal(m *pgtype.Map) {
	m.RegisterType(&pgtype.Type{
		Name:  "numeric",
		OID:   pgtype.NumericOID,
		Codec: NumericCodec{},
	})

	registerDefaultPgTypeVariants := func(name, arrayName string, value any) {
		// T
		m.RegisterDefaultPgType(value, name)

		// *T
		valueType := reflect.TypeOf(value)
		m.RegisterDefaultPgType(reflect.New(valueType).Interface(), name)

		// []T
		sliceType := reflect.SliceOf(valueType)
		m.RegisterDefaultPgType(reflect.MakeSlice(sliceType, 0, 0).Interface(), arrayName)

		// *[]T
		m.RegisterDefaultPgType(reflect.New(sliceType).Interface(), arrayName)

		// []*T
		sliceOfPointerType := reflect.SliceOf(reflect.TypeOf(reflect.New(valueType).Interface()))
		m.RegisterDefaultPgType(reflect.MakeSlice(sliceOfPointerType, 0, 0).Interface(), arrayName)

		// *[]*T
		m.RegisterDefaultPgType(reflect.New(sliceOfPointerType).Interface(), arrayName)
	}

	registerDefaultPgTypeVariants("numeric", "_numeric", Decimal{})
	registerDefaultPgTypeVariants("numeric", "_numeric", OptDecimal{})
	registerDefaultPgTypeVariants("numeric", "_numeric", Decimal{})
	registerDefaultPgTypeVariants("numeric", "_numeric", OptDecimal{})
}
