package kzgceremony

import (
	"encoding/hex"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	bls12381 "github.com/kilic/bls12-381"
)

func TestParseSRS(t *testing.T) {
	c := qt.New(t)

	srs_0 := newEmptySRS(2, 2)

	contr_1, err := Contribute(srs_0, []byte("1111111111111111111111111111111111111111111111111111111111111111"))
	c.Assert(err, qt.IsNil)

	g1s, g2s, err := contr_1.SRS.ToEthJSON()
	c.Assert(err, qt.IsNil)

	parsedSRS, err := ParseEthJSON(g1s, g2s)
	c.Assert(err, qt.IsNil)

	g1 := bls12381.NewG1()
	g2 := bls12381.NewG2()
	for i := 0; i < len(contr_1.SRS.G1s); i++ {
		c.Assert(g1.Equal(parsedSRS.G1s[i], contr_1.SRS.G1s[i]), qt.IsTrue)
	}
	for i := 0; i < len(contr_1.SRS.G2s); i++ {
		c.Assert(g2.Equal(parsedSRS.G2s[i], contr_1.SRS.G2s[i]), qt.IsTrue)
	}
}

func TestParseCompressedG1Point(t *testing.T) {
	// this test is just to check that github.com/kilic/bls12-381 is
	// compatible with the compressed points from zkcrypto/bls12-381
	c := qt.New(t)

	g1 := bls12381.NewG1()

	p1Str := "0x97f1d3a73197d7942695638c4fa9ac0fc3688c4f9774b905a14e3a3f171bac586c55e83ff97a1aeffb3af00adb22c6bb"
	g1Bytes, err := hex.DecodeString(strings.TrimPrefix(p1Str, "0x"))
	g1Point, err := g1.FromCompressed(g1Bytes)
	c.Assert(err, qt.IsNil)

	recompressed := g1.ToCompressed(g1Point)
	c.Assert("0x"+hex.EncodeToString(recompressed), qt.Equals, p1Str)

	p1Str = "0xc00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	g1Bytes, err = hex.DecodeString(strings.TrimPrefix(p1Str, "0x"))
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
	g2Point, err := g2.FromCompressed(g2Bytes)
	c.Assert(err, qt.IsNil)

	recompressed := g2.ToCompressed(g2Point)
	c.Assert("0x"+hex.EncodeToString(recompressed), qt.Equals, p2Str)

	p2Str = "0xc00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	g2Bytes, err = hex.DecodeString(strings.TrimPrefix(p2Str, "0x"))
	g2Point, err = g2.FromCompressed(g2Bytes)
	c.Assert(err, qt.IsNil)

	recompressed = g2.ToCompressed(g2Point)
	c.Assert("0x"+hex.EncodeToString(recompressed), qt.Equals, p2Str)
	// additionally check that g1Point is zero
	c.Assert(g2.Equal(g2Point, g2.Zero()), qt.IsTrue)
}

func TestVectorFromRustImpl(t *testing.T) {
	c := qt.New(t)
	testVectorG1s := []string{
		"0x97f1d3a73197d7942695638c4fa9ac0fc3688c4f9774b905a14e3a3f171bac586c55e83ff97a1aeffb3af00adb22c6bb",
		"0xae4a5e0fe4947fd5551664fc49453d805ce6af2dd6f053e147d95d9ede825190b06028f228f47c0d3118b71313235aa7",
		"0x888a1073366b3e974c318be5862f607c18fb5241cfefc3c8b82c31c4db10f866c546804db6e44bf61b61d8c31d006099",
		"0x84afca7ce5cc42aa998725610d14cd47078e523a9cb96b66856f563d62d44ff6489f0b1827b853e94d9474b75a138ec6",
		"0x8ae778a5b534c6eadcce4c4e7dd1dab3674d20bfe686f70506fba1695bbdae2455169b035505e5246d726a9e0c982335",
	}
	testVectorG2s := []string{
		"0x93e02b6052719f607dacd3a088274f65596bd0d09920b61ab5da61bbdc7f5049334cf11213945d57e5ac7d055d042b7e024aa2b2f08f0a91260805272dc51051c6e47ad4fa403b02b4510b647ae3d1770bac0326a805bbefd48056c8c121bdb8",
		"0xa5d4fc7bfa882ef674fbbcdda73c66c2627a5a723ddf6fc0f242b3f71301a6e7181708a8c63756787a2c3e47c2b347ca18c64b103461036830880cd26bacad80cdbfe00c69e9214b7c90bf30ab0804f520bed3ce608aefeb6f85011f0d745575",
		"0x87b93440866545adf37f5119d640036b2856a790bd09c2b7f69a1a1bdc7846c5f48f622fcbf5c46a51dede46a4682b7b118542548df17ff7c7e6f63049e1d1c56ca1a637f8a55136d0ed7ab466e78550f635dc06d78d87a3003dab133b8e275a",
	}

	parsedSRS, err := ParseEthJSON(testVectorG1s, testVectorG2s)
	c.Assert(err, qt.IsNil)

	g1s, g2s, err := parsedSRS.ToEthJSON()
	c.Assert(err, qt.IsNil)

	c.Assert(len(testVectorG1s), qt.Equals, len(g1s))
	c.Assert(len(testVectorG2s), qt.Equals, len(g2s))

	for i := 0; i < len(g1s); i++ {
		c.Assert(g1s[i], qt.Equals, testVectorG1s[i])
	}
	for i := 0; i < len(g2s); i++ {
		c.Assert(g2s[i], qt.Equals, testVectorG2s[i])
	}
}
