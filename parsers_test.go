package kzgceremony

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	bls12381 "github.com/kilic/bls12-381"
)

func TestStateMarshalers(t *testing.T) {
	c := qt.New(t)
	j, err := ioutil.ReadFile("current_state_10.json")
	c.Assert(err, qt.IsNil)

	state := &State{}
	err = json.Unmarshal(j, state)
	c.Assert(err, qt.IsNil)

	b, err := json.Marshal(state)
	c.Assert(err, qt.IsNil)
	err = ioutil.WriteFile("parsed_state.json", b, 0600)
	c.Assert(err, qt.IsNil)
}

func TestParseCompressedG1Point(t *testing.T) {
	// this test is just to check that github.com/kilic/bls12-381 is
	// compatible with the compressed points from zkcrypto/bls12-381
	c := qt.New(t)

	g1 := bls12381.NewG1()

	p1Str := "0x97f1d3a73197d7942695638c4fa9ac0fc3688c4f9774b905a14e3a3f171bac586c55e83ff97a1aeffb3af00adb22c6bb"
	g1Bytes, err := hex.DecodeString(strings.TrimPrefix(p1Str, "0x"))
	c.Assert(err, qt.IsNil)
	g1Point, err := g1.FromCompressed(g1Bytes)
	c.Assert(err, qt.IsNil)

	recompressed := g1.ToCompressed(g1Point)
	c.Assert("0x"+hex.EncodeToString(recompressed), qt.Equals, p1Str)

	p1Str = "0xc00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	g1Bytes, err = hex.DecodeString(strings.TrimPrefix(p1Str, "0x"))
	c.Assert(err, qt.IsNil)
	g1Point, err = g1.FromCompressed(g1Bytes)
	c.Assert(err, qt.IsNil)

	recompressed = g1.ToCompressed(g1Point)
	c.Assert("0x"+hex.EncodeToString(recompressed), qt.Equals, p1Str)
	// additionally check that g1Point is zero
	c.Assert(g1.Equal(g1Point, g1.Zero()), qt.IsTrue)
}

func TestParseCompressedG2Point(t *testing.T) {
	// this test is just to check that github.com/kilic/bls12-381 is
	// compatible with the compressed points from zkcrypto/bls12-381
	c := qt.New(t)

	g2 := bls12381.NewG2()

	p2Str := "0x93e02b6052719f607dacd3a088274f65596bd0d09920b61ab5da61bbdc7f5049334cf11213945d57e5ac7d055d042b7e024aa2b2f08f0a91260805272dc51051c6e47ad4fa403b02b4510b647ae3d1770bac0326a805bbefd48056c8c121bdb8"
	g2Bytes, err := hex.DecodeString(strings.TrimPrefix(p2Str, "0x"))
	c.Assert(err, qt.IsNil)
	g2Point, err := g2.FromCompressed(g2Bytes)
	c.Assert(err, qt.IsNil)

	recompressed := g2.ToCompressed(g2Point)
	c.Assert("0x"+hex.EncodeToString(recompressed), qt.Equals, p2Str)

	p2Str = "0xc00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	g2Bytes, err = hex.DecodeString(strings.TrimPrefix(p2Str, "0x"))
	c.Assert(err, qt.IsNil)
	g2Point, err = g2.FromCompressed(g2Bytes)
	c.Assert(err, qt.IsNil)

	recompressed = g2.ToCompressed(g2Point)
	c.Assert("0x"+hex.EncodeToString(recompressed), qt.Equals, p2Str)
	// additionally check that g1Point is zero
	c.Assert(g2.Equal(g2Point, g2.Zero()), qt.IsTrue)
}
