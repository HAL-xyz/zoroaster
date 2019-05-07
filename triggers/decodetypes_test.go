package trigger

import (
	"reflect"
	"testing"
)

func TestDecode2DBytesArray(t *testing.T) {

	// slice of static array
	var data interface{} = [][4]uint8{{1,2,3,4},{0,0,0,255}}
	dec := Decode2DBytesArray(data)

	if len(dec) != 2 {
		t.Error()
	}

	have := MultArrayToHex(dec)
	want := []string{"01020304", "000000ff"}

	if !reflect.DeepEqual(have, want) {
		t.Error()
	}
}

func TestDecode2DBytesArray2(t *testing.T) {

	// array of arrays
	var data interface{} = [2][4]uint8{{1,2,3,4},{0,0,0,255}}
	dec := Decode2DBytesArray(data)

	if len(dec) != 2 {
		t.Error()
	}

	have := MultArrayToHex(dec)
	want := []string{"01020304", "000000ff"}

	if !reflect.DeepEqual(have, want) {
		t.Error()
	}
}
