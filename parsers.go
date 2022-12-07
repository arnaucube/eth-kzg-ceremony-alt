package kzgceremony

import (
	"encoding/hex"
	"strings"

	bls12381 "github.com/kilic/bls12-381"
)

// WIP

// ParseEthJSON parses the eth-kzg-ceremony SRS json format
func ParseEthJSON(g1sStr, g2sStr []string) (*SRS, error) {
	g1 := bls12381.NewG1()
	g2 := bls12381.NewG2()
	nG1s := len(g1sStr)
	nG2s := len(g2sStr)
	g1s := make([]*bls12381.PointG1, nG1s)
	g2s := make([]*bls12381.PointG2, nG2s)
	for i := 0; i < nG1s; i++ {
		g1sBytes, err := hex.DecodeString(strings.TrimPrefix(g1sStr[i], "0x"))
		if err != nil {
			return nil, err
		}
		g1s_i, err := g1.FromCompressed(g1sBytes)
		if err != nil {
			return nil, err
		}
		g1s[i] = g1s_i
	}
	for i := 0; i < nG2s; i++ {
		g2sBytes, err := hex.DecodeString(strings.TrimPrefix(g2sStr[i], "0x"))
		if err != nil {
			return nil, err
		}
		g2s_i, err := g2.FromCompressed(g2sBytes)
		if err != nil {
			return nil, err
		}
		g2s[i] = g2s_i
	}
	return &SRS{G1s: g1s, G2s: g2s}, nil
}

// ToEthJSON outputs the SRS into the eth-kzg-ceremony SRS json format
func (srs *SRS) ToEthJSON() ([]string, []string, error) {
	g1 := bls12381.NewG1()
	g2 := bls12381.NewG2()
	nG1s := len(srs.G1s)
	nG2s := len(srs.G2s)
	g1s := make([]string, nG1s)
	g2s := make([]string, nG2s)

	for i := 0; i < nG1s; i++ {
		g1s[i] = "0x" + hex.EncodeToString(g1.ToCompressed(srs.G1s[i]))
	}
	for i := 0; i < nG2s; i++ {
		g2s[i] = "0x" + hex.EncodeToString(g2.ToCompressed(srs.G2s[i]))
	}
	return g1s, g2s, nil
}
