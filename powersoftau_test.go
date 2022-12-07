package kzgceremony

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestContribute(t *testing.T) {
	c := qt.New(t)

	srs_0 := newEmptySRS(10, 10)

	contr_1, err := Contribute(srs_0, []byte("1111111111111111111111111111111111111111111111111111111111111111"))
	c.Assert(err, qt.IsNil)

	_, err = Contribute(contr_1.SRS, []byte("2222222222222222222222222222222222222222222222222222222222222222"))
	c.Assert(err, qt.IsNil)
}
