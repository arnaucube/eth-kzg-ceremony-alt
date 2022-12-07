package kzgceremony

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestContribution(t *testing.T) {
	c := qt.New(t)

	srs_0 := newEmptySRS(10, 10)

	contr_1, err := Contribute(srs_0, []byte("1111111111111111111111111111111111111111111111111111111111111111"))
	c.Assert(err, qt.IsNil)

	c.Assert(Verify(srs_0, contr_1.SRS, contr_1.Proof), qt.IsTrue)

	contr_2, err := Contribute(contr_1.SRS, []byte("2222222222222222222222222222222222222222222222222222222222222222"))
	c.Assert(err, qt.IsNil)
	c.Assert(Verify(contr_1.SRS, contr_2.SRS, contr_2.Proof), qt.IsTrue)
}
