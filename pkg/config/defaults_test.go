package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type D struct {
	VPStruct  *NestedD
	VInt      int               `default:"1"`
	VBool     bool              `default:"true"`
	VInt8     int8              `default:"2"`
	VInt16    int16             `default:"3"`
	VInt32    int32             `default:"-4"`
	VInt64    int64             `default:"5"`
	VUnit     uint              `default:"6"`
	VUint8    uint8             `default:"7"`
	VUint16   uint16            `default:"8"`
	VUint32   uint32            `default:"9"`
	VUint64   uint64            `default:"10"`
	VFloat32  float32           `default:"1.1"`
	VFloat64  float64           `default:"1.2"`
	VString   string            `default:"s1"`
	VIntSlice []int             `default:"[3,2,1]"`
	VSMap     map[string]string `default:"{\"1\":\"2\"}"`
	VStruct   NestedD
}

type NestedD struct {
	StringVal string         `default:"s2"`
	IntVal    int            `default:"-1024"`
	Map       map[string]int `default:"{\"one\":2}"`
}

func TestDefautsInvalid(t *testing.T) {
	err := SetDefaults(1)

	require.Error(t, err)
	require.ErrorIs(t, err, errInvalidType)

	i := 0

	err = SetDefaults(&i)
	require.Error(t, err)
	require.ErrorIs(t, err, errInvalidType)
}

func TestDefaults(t *testing.T) {
	s := D{VPStruct: new(NestedD)}

	err := SetDefaults(&s)

	require.NoError(t, err)

	expected := D{
		VInt:      1,
		VBool:     true,
		VInt8:     2,
		VInt16:    3,
		VInt32:    -4,
		VInt64:    5,
		VUnit:     6,
		VUint8:    7,
		VUint16:   8,
		VUint32:   9,
		VUint64:   10,
		VFloat32:  1.1,
		VFloat64:  1.2,
		VString:   "s1",
		VIntSlice: []int{3, 2, 1},
		VSMap:     map[string]string{"1": "2"},
		VPStruct: &NestedD{
			StringVal: "s2",
			IntVal:    -1024,
			Map:       map[string]int{"one": 2},
		},
		VStruct: NestedD{
			StringVal: "s2",
			IntVal:    -1024,
			Map:       map[string]int{"one": 2},
		},
	}

	require.Equal(t, expected, s)
}
