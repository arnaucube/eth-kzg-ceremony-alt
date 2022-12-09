package kzgceremony

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	bls12381 "github.com/kilic/bls12-381"
)

func (s *State) UnmarshalJSON(b []byte) error {
	var sStr stateStr
	if err := json.Unmarshal(b, &sStr); err != nil {
		return err
	}
	var err error
	s.ParticipantIDs = sStr.ParticipantIDs
	s.ParticipantECDSASignatures = sStr.ParticipantECDSASignatures

	s.Transcripts = make([]Transcript, len(sStr.Transcripts))
	for i := 0; i < len(sStr.Transcripts); i++ {
		if sStr.Transcripts[i].NumG1Powers != uint64(len(sStr.Transcripts[i].PowersOfTau.G1Powers)) {
			return fmt.Errorf("wrong NumG1Powers")
		}
		if sStr.Transcripts[i].NumG2Powers != uint64(len(sStr.Transcripts[i].PowersOfTau.G2Powers)) {
			return fmt.Errorf("wrong NumG2Powers")
		}
		s.Transcripts[i].NumG1Powers = sStr.Transcripts[i].NumG1Powers
		s.Transcripts[i].NumG2Powers = sStr.Transcripts[i].NumG2Powers
		s.Transcripts[i].PowersOfTau = &SRS{}
		s.Transcripts[i].PowersOfTau.G1Powers, err =
			stringsToPointsG1(sStr.Transcripts[i].PowersOfTau.G1Powers)
		if err != nil {
			return err
		}
		s.Transcripts[i].PowersOfTau.G2Powers, err =
			stringsToPointsG2(sStr.Transcripts[i].PowersOfTau.G2Powers)
		if err != nil {
			return err
		}

		s.Transcripts[i].Witness = &Witness{}
		s.Transcripts[i].Witness.RunningProducts, err =
			stringsToPointsG1(sStr.Transcripts[i].Witness.RunningProducts)
		if err != nil {
			return err
		}
		s.Transcripts[i].Witness.PotPubKeys, err =
			stringsToPointsG2(sStr.Transcripts[i].Witness.PotPubKeys)
		if err != nil {
			return err
		}
		s.Transcripts[i].Witness.BLSSignatures, err =
			stringsToPointsG1(sStr.Transcripts[i].Witness.BLSSignatures)
		if err != nil {
			return err
		}
	}
	// TODO validate data (G1 & G2 subgroup checks, etc)
	return err
}

func (s State) MarshalJSON() ([]byte, error) {
	var sStr stateStr
	sStr.ParticipantIDs = s.ParticipantIDs
	sStr.ParticipantECDSASignatures = s.ParticipantECDSASignatures

	sStr.Transcripts = make([]transcriptStr, len(s.Transcripts))
	for i := 0; i < len(s.Transcripts); i++ {
		if s.Transcripts[i].NumG1Powers != uint64(len(s.Transcripts[i].PowersOfTau.G1Powers)) {
			return nil, fmt.Errorf("wrong NumG1Powers")
		}
		if s.Transcripts[i].NumG2Powers != uint64(len(s.Transcripts[i].PowersOfTau.G2Powers)) {
			return nil, fmt.Errorf("wrong NumG2Powers")
		}
		sStr.Transcripts[i].NumG1Powers = s.Transcripts[i].NumG1Powers
		sStr.Transcripts[i].NumG2Powers = s.Transcripts[i].NumG2Powers
		sStr.Transcripts[i].PowersOfTau = powersOfTauStr{}
		sStr.Transcripts[i].PowersOfTau.G1Powers =
			g1PointsToStrings(s.Transcripts[i].PowersOfTau.G1Powers)
		sStr.Transcripts[i].PowersOfTau.G2Powers =
			g2PointsToStrings(s.Transcripts[i].PowersOfTau.G2Powers)

		sStr.Transcripts[i].Witness = witnessStr{}
		sStr.Transcripts[i].Witness.RunningProducts =
			g1PointsToStrings(s.Transcripts[i].Witness.RunningProducts)
		sStr.Transcripts[i].Witness.PotPubKeys =
			g2PointsToStrings(s.Transcripts[i].Witness.PotPubKeys)
		sStr.Transcripts[i].Witness.BLSSignatures =
			g1PointsToStrings(s.Transcripts[i].Witness.BLSSignatures)
	}
	return json.Marshal(sStr)
}

func (c *BatchContribution) UnmarshalJSON(b []byte) error {
	var cStr batchContributionStr
	if err := json.Unmarshal(b, &cStr); err != nil {
		return err
	}
	var err error
	g2 := bls12381.NewG2()

	c.Contributions = make([]Contribution, len(cStr.Contributions))
	for i := 0; i < len(cStr.Contributions); i++ {
		c.Contributions[i].NumG1Powers = cStr.Contributions[i].NumG1Powers
		c.Contributions[i].NumG2Powers = cStr.Contributions[i].NumG2Powers
		c.Contributions[i].PowersOfTau = &SRS{}
		c.Contributions[i].PowersOfTau.G1Powers, err =
			stringsToPointsG1(cStr.Contributions[i].PowersOfTau.G1Powers)
		if err != nil {
			return err
		}
		c.Contributions[i].PowersOfTau.G2Powers, err =
			stringsToPointsG2(cStr.Contributions[i].PowersOfTau.G2Powers)
		if err != nil {
			return err
		}

		g2sBytes, err := hex.DecodeString(strings.TrimPrefix(cStr.Contributions[i].PotPubKey, "0x"))
		if err != nil {
			return err
		}
		c.Contributions[i].PotPubKey, err = g2.FromCompressed(g2sBytes)
		if err != nil {
			return err
		}
	}
	return err
}

func (c BatchContribution) MarshalJSON() ([]byte, error) {
	var cStr batchContributionStr
	g2 := bls12381.NewG2()

	cStr.Contributions = make([]contributionStr, len(c.Contributions))
	for i := 0; i < len(c.Contributions); i++ {
		cStr.Contributions[i].NumG1Powers = c.Contributions[i].NumG1Powers
		cStr.Contributions[i].NumG2Powers = c.Contributions[i].NumG2Powers
		cStr.Contributions[i].PowersOfTau = powersOfTauStr{}
		cStr.Contributions[i].PowersOfTau.G1Powers =
			g1PointsToStrings(c.Contributions[i].PowersOfTau.G1Powers)
		cStr.Contributions[i].PowersOfTau.G2Powers =
			g2PointsToStrings(c.Contributions[i].PowersOfTau.G2Powers)

		cStr.Contributions[i].PotPubKey = "0x" + hex.EncodeToString(g2.ToCompressed(c.Contributions[i].PotPubKey))
	}
	return json.Marshal(cStr)
}

type powersOfTauStr struct {
	G1Powers []string `json:"G1Powers"`
	G2Powers []string `json:"G2Powers"`
}

type witnessStr struct {
	RunningProducts []string `json:"runningProducts"`
	PotPubKeys      []string `json:"potPubkeys"`
	BLSSignatures   []string `json:"blsSignatures"`
}

type transcriptStr struct {
	NumG1Powers uint64         `json:"numG1Powers"`
	NumG2Powers uint64         `json:"numG2Powers"`
	PowersOfTau powersOfTauStr `json:"powersOfTau"`
	Witness     witnessStr     `json:"witness"`
}

type contributionStr struct {
	NumG1Powers uint64         `json:"numG1Powers"`
	NumG2Powers uint64         `json:"numG2Powers"`
	PowersOfTau powersOfTauStr `json:"powersOfTau"`
	PotPubKey   string         `json:"potPubkey"`
}

type batchContributionStr struct {
	Contributions []contributionStr `json:"contributions"`
}

type stateStr struct {
	Transcripts                []transcriptStr `json:"transcripts"`
	ParticipantIDs             []string        `json:"participantIds"`
	ParticipantECDSASignatures []string        `json:"participantEcdsaSignatures"`
}

func g1PointsToStrings(points []*bls12381.PointG1) []string {
	g1 := bls12381.NewG1() // TODO unify g1 instantiation (& g2)
	n := len(points)
	g1s := make([]string, n)
	for i := 0; i < n; i++ {
		if points[i] == nil {
			g1s[i] = ""
			continue
		}
		g1s[i] = "0x" + hex.EncodeToString(g1.ToCompressed(points[i]))
	}
	return g1s
}

func g2PointsToStrings(points []*bls12381.PointG2) []string {
	g2 := bls12381.NewG2()
	n := len(points)
	g2s := make([]string, n)
	for i := 0; i < n; i++ {
		if points[i] == nil {
			g2s[i] = ""
			continue
		}
		g2s[i] = "0x" + hex.EncodeToString(g2.ToCompressed(points[i]))
	}
	return g2s
}

func stringsToPointsG1(s []string) ([]*bls12381.PointG1, error) {
	g1 := bls12381.NewG1() // TODO unify g1 instantiation (& g2)
	n := len(s)
	g1s := make([]*bls12381.PointG1, n)
	for i := 0; i < n; i++ {
		if s[i] == "" {
			continue
		}
		g1sBytes, err := hex.DecodeString(strings.TrimPrefix(s[i], "0x"))
		if err != nil {
			return nil, err
		}
		g1s_i, err := g1.FromCompressed(g1sBytes)
		if err != nil {
			return nil, err
		}
		g1s[i] = g1s_i
	}
	return g1s, nil
}
func stringsToPointsG2(s []string) ([]*bls12381.PointG2, error) {
	g2 := bls12381.NewG2()
	n := len(s)
	g2s := make([]*bls12381.PointG2, n)
	for i := 0; i < n; i++ {
		if s[i] == "" {
			continue
		}
		g2sBytes, err := hex.DecodeString(strings.TrimPrefix(s[i], "0x"))
		if err != nil {
			return nil, err
		}
		g2s_i, err := g2.FromCompressed(g2sBytes)
		if err != nil {
			return nil, err
		}
		g2s[i] = g2s_i
	}
	return g2s, nil
}
