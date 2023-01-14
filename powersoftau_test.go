package kzgceremony

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestContribution(t *testing.T) {
	c := qt.New(t)

	srs_0 := newEmptySRS(10, 10)

	srs_1, proof_1, err := Contribute(srs_0, 0,
		[]byte("1111111111111111111111111111111111111111111111111111111111111111"))
	c.Assert(err, qt.IsNil)

	c.Assert(VerifyNewSRSFromPrevSRS(srs_0, srs_1, proof_1), qt.IsTrue)

	srs_2, proof_2, err := Contribute(srs_1, 0,
		[]byte("2222222222222222222222222222222222222222222222222222222222222222"))
	c.Assert(err, qt.IsNil)
	c.Assert(VerifyNewSRSFromPrevSRS(srs_1, srs_2, proof_2), qt.IsTrue)
}

func TestComputeNewState(t *testing.T) {
	c := qt.New(t)
	j, err := ioutil.ReadFile("current_state_10.json")
	c.Assert(err, qt.IsNil)

	cs := &State{}
	err = json.Unmarshal(j, cs)
	c.Assert(err, qt.IsNil)

	newState, err :=
		cs.Contribute([]byte("1111111111111111111111111111111111111111111111111111111111111111"))
	c.Assert(err, qt.IsNil)

	b, err := json.Marshal(newState)
	c.Assert(err, qt.IsNil)
	err = ioutil.WriteFile("new_state.json", b, 0600)
	c.Assert(err, qt.IsNil)
}

func TestBatchContribution(t *testing.T) {
	c := qt.New(t)
	j, err := ioutil.ReadFile("batch_contribution_10.json")
	c.Assert(err, qt.IsNil)

	bc := &BatchContribution{}
	err = json.Unmarshal(j, bc)
	c.Assert(err, qt.IsNil)

	nb, err :=
		bc.Contribute([]byte("1111111111111111111111111111111111111111111111111111111111111111"))
	c.Assert(err, qt.IsNil)

	c.Assert(len(nb.Contributions), qt.Equals, 4)
	c.Assert(g2.Equal(nb.Contributions[0].PotPubKey, nb.Contributions[1].PotPubKey), qt.IsFalse)
	c.Assert(g2.Equal(nb.Contributions[0].PotPubKey, nb.Contributions[2].PotPubKey), qt.IsFalse)
	c.Assert(g2.Equal(nb.Contributions[0].PotPubKey, nb.Contributions[3].PotPubKey), qt.IsFalse)

	_, err = json.Marshal(nb)
	c.Assert(err, qt.IsNil)
}
