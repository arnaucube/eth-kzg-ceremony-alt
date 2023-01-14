package kzgceremony

import (
	"fmt"
	"math/big"

	"golang.org/x/crypto/blake2b"

	bls12381 "github.com/kilic/bls12-381"
)

// MinRandomnessLen is the minimum accepted length for the user defined
// randomness
const MinRandomnessLen = 64

var g1 *bls12381.G1
var g2 *bls12381.G2

func init() {
	g1 = bls12381.NewG1()
	g2 = bls12381.NewG2()
}

// State represents the data structure obtained from the Sequencer at the
// /info/current_state endpoint
type State struct {
	Transcripts                []Transcript
	ParticipantIDs             []string // WIP
	ParticipantECDSASignatures []string
}

// BatchContribution represents the data structure obtained from the Sequencer
// at the /contribute endpoint
type BatchContribution struct {
	Contributions []Contribution
}

type Contribution struct {
	NumG1Powers uint64
	NumG2Powers uint64
	PowersOfTau *SRS
	PotPubKey   *bls12381.PointG2
}

type Transcript struct {
	NumG1Powers uint64
	NumG2Powers uint64
	PowersOfTau *SRS
	Witness     *Witness
}

type Witness struct {
	RunningProducts []*bls12381.PointG1
	PotPubKeys      []*bls12381.PointG2
	BLSSignatures   []*bls12381.PointG1
}

// SRS contains the powers of tau in G1 & G2, eg.
// [τ'⁰]₁, [τ'¹]₁, [τ'²]₁, ..., [τ'ⁿ⁻¹]₁,
// [τ'⁰]₂, [τ'¹]₂, [τ'²]₂, ..., [τ'ⁿ⁻¹]₂
type SRS struct {
	G1Powers []*bls12381.PointG1
	G2Powers []*bls12381.PointG2
}

type toxicWaste struct {
	tau   *big.Int
	TauG2 *bls12381.PointG2 // Proof.G2P
}

// Proof contains g₂ᵖ and g₂^τ', used by the verifier
type Proof struct {
	G2P    *bls12381.PointG2 // g₂ᵖ
	G1PTau *bls12381.PointG1 // g₂^τ' = g₂^{p ⋅ τ}
}

// Contribute takes the last State and computes a new State using the defined
// randomness
func (cs *State) Contribute(randomness []byte) (*State, error) {
	ns := State{}
	ns.Transcripts = make([]Transcript, len(cs.Transcripts))
	for i := 0; i < len(cs.Transcripts); i++ {
		ns.Transcripts[i].NumG1Powers = cs.Transcripts[i].NumG1Powers
		ns.Transcripts[i].NumG2Powers = cs.Transcripts[i].NumG2Powers

		newSRS, proof, err := Contribute(cs.Transcripts[i].PowersOfTau, i, randomness)
		if err != nil {
			return nil, err
		}
		ns.Transcripts[i].PowersOfTau = newSRS

		ns.Transcripts[i].Witness = &Witness{}
		ns.Transcripts[i].Witness.RunningProducts =
			append(cs.Transcripts[i].Witness.RunningProducts, proof.G1PTau)
		ns.Transcripts[i].Witness.PotPubKeys =
			append(cs.Transcripts[i].Witness.PotPubKeys, proof.G2P)
		ns.Transcripts[i].Witness.BLSSignatures = cs.Transcripts[i].Witness.BLSSignatures
	}
	ns.ParticipantIDs = cs.ParticipantIDs // TODO add github id (id_token.sub)
	ns.ParticipantECDSASignatures = cs.ParticipantECDSASignatures

	return &ns, nil
}

// Contribute takes the last BatchContribution and computes a new
// BatchContribution using the defined randomness
func (pb *BatchContribution) Contribute(randomness []byte) (*BatchContribution, error) {
	nb := BatchContribution{}
	nb.Contributions = make([]Contribution, len(pb.Contributions))
	for i := 0; i < len(pb.Contributions); i++ {
		nb.Contributions[i].NumG1Powers = pb.Contributions[i].NumG1Powers
		nb.Contributions[i].NumG2Powers = pb.Contributions[i].NumG2Powers

		newSRS, proof, err := Contribute(pb.Contributions[i].PowersOfTau, i, randomness)
		if err != nil {
			return nil, err
		}
		nb.Contributions[i].PowersOfTau = newSRS

		nb.Contributions[i].PotPubKey = proof.G2P
	}

	return &nb, nil
}

// newEmptySRS creates an empty SRS filled by [1]₁ & [1]₂ points in all
// respective arrays positions
func newEmptySRS(nG1, nG2 int) *SRS {
	g1s := make([]*bls12381.PointG1, nG1)
	g2s := make([]*bls12381.PointG2, nG2)
	for i := 0; i < nG1; i++ {
		g1s[i] = g1.One()
	}
	for i := 0; i < nG2; i++ {
		g2s[i] = g2.One()
	}
	return &SRS{g1s, g2s}
}

func tau(round int, randomness []byte) *toxicWaste {
	val := blake2b.Sum256(randomness)
	tau := new(big.Int).Mod(
		new(big.Int).SetBytes(val[:]),
		g2.Q())
	tau_Fr := bls12381.NewFr().FromBytes(tau.Bytes())
	TauG2 := g2.New()
	g2.MulScalar(TauG2, g2.One(), tau_Fr)

	return &toxicWaste{tau, TauG2}
}

func computeContribution(t *toxicWaste, prevSRS *SRS) *SRS {
	srs := newEmptySRS(len(prevSRS.G1Powers), len(prevSRS.G2Powers))
	Q := g1.Q() // Q = |G1| == |G2|

	// fmt.Println("Computing [τ'⁰]₁, [τ'¹]₁, [τ'²]₁, ..., [τ'ⁿ⁻¹]₁, for n =", len(prevSRS.G1s))
	for i := 0; i < len(prevSRS.G1Powers); i++ {
		tau_i := new(big.Int).Exp(t.tau, big.NewInt(int64(i)), Q)
		tau_i_Fr := bls12381.NewFr().FromBytes(tau_i.Bytes())
		g1.MulScalar(srs.G1Powers[i], prevSRS.G1Powers[i], tau_i_Fr)
	}
	// fmt.Println("Computing [τ'⁰]₂, [τ'¹]₂, [τ'²]₂, ..., [τ'ⁿ⁻¹]₂, for n =", len(prevSRS.G2s))
	for i := 0; i < len(prevSRS.G2Powers); i++ {
		tau_i := new(big.Int).Exp(t.tau, big.NewInt(int64(i)), Q)
		tau_i_Fr := bls12381.NewFr().FromBytes(tau_i.Bytes())
		g2.MulScalar(srs.G2Powers[i], prevSRS.G2Powers[i], tau_i_Fr)
	}

	return srs
}

func genProof(toxicWaste *toxicWaste, prevSRS, newSRS *SRS) *Proof {
	G1_p := g1.New()
	tau_Fr := bls12381.NewFr().FromBytes(toxicWaste.tau.Bytes())
	g1.MulScalar(G1_p, prevSRS.G1Powers[1], tau_Fr) // g_1^{tau'} = g_1^{p * tau}, where p=toxicWaste.tau

	return &Proof{toxicWaste.TauG2, G1_p}
}

// Contribute takes as input the previous SRS and a random
// byte slice, and returns the new SRS together with the Proof
func Contribute(prevSRS *SRS, round int, randomness []byte) (*SRS, *Proof, error) {
	if len(randomness) < MinRandomnessLen {
		return nil, nil, fmt.Errorf("err randomness") // WIP
	}
	// set tau from randomness
	tw := tau(round, randomness)

	newSRS := computeContribution(tw, prevSRS)

	proof := genProof(tw, prevSRS, newSRS)

	return newSRS, proof, nil
}

func checkG1PointCorrectness(p *bls12381.PointG1) error {
	// i) non-empty
	if p == nil {
		return fmt.Errorf("empty point value")
	}
	// ii) non-zero
	if g1.IsZero(p) {
		return fmt.Errorf("point can not be zero")
	}
	// iii) in the correct prime order of subgroups
	if !g1.IsOnCurve(p) {
		return fmt.Errorf("point not on curve")
	}
	if !g1.InCorrectSubgroup(p) {
		return fmt.Errorf("point not in the correct prime order of subgroups")
	}
	return nil
}

func checkG2PointCorrectness(p *bls12381.PointG2) error {
	// i) non-empty
	if p == nil {
		return fmt.Errorf("empty point value")
	}
	// ii) non-zero
	if g2.IsZero(p) {
		return fmt.Errorf("point can not be zero")
	}
	// iii) in the correct prime order of subgroups
	if !g2.IsOnCurve(p) {
		return fmt.Errorf("point not on curve")
	}
	if !g2.InCorrectSubgroup(p) {
		return fmt.Errorf("point not in the correct prime order of subgroups")
	}
	return nil
}

// VerifyNewSRSFromPrevSRS checks the correct computation of the new SRS
// respectively from the previous SRS. These are the checks that the Sequencer
// would do.
func VerifyNewSRSFromPrevSRS(prevSRS, newSRS *SRS, proof *Proof) bool {
	pairing := bls12381.NewEngine()

	// 1. check that elements of the newSRS are valid points
	for i := 0; i < len(newSRS.G1Powers); i++ {
		if err := checkG1PointCorrectness(newSRS.G1Powers[i]); err != nil {
			return false
		}
	}
	for i := 0; i < len(newSRS.G2Powers); i++ {
		if err := checkG2PointCorrectness(newSRS.G2Powers[i]); err != nil {
			return false
		}
	}

	// 2. check proof.G1PTau == newSRS.G1Powers[1]
	if !g1.Equal(proof.G1PTau, newSRS.G1Powers[1]) {
		return false
	}

	// 3. check newSRS.G1s[1] (g₁^τ'), is correctly related to prevSRS.G1s[1] (g₁^τ)
	//   e([τ]₁, [p]₂) == e([τ']₁, [1]₂)
	eL := pairing.AddPair(prevSRS.G1Powers[1], proof.G2P).Result()
	eR := pairing.AddPair(newSRS.G1Powers[1], g2.One()).Result()
	if !eL.Equal(eR) {
		return false
	}

	// 4. check newSRS following the powers of tau structure
	for i := 0; i < len(newSRS.G1Powers)-1; i++ {
		// i) e([τ'ⁱ]₁, [τ']₂) == e([τ'ⁱ⁺¹]₁, [1]₂), for i ∈ [1, n−1]
		eL := pairing.AddPair(newSRS.G1Powers[i], newSRS.G2Powers[1]).Result()
		eR := pairing.AddPair(newSRS.G1Powers[i+1], g2.One()).Result()
		if !eL.Equal(eR) {
			return false
		}
	}

	for i := 0; i < len(newSRS.G2Powers)-1; i++ {
		// ii) e([τ']₁, [τ'ʲ]₂) == e([1]₁, [τ'ʲ⁺¹]₂), for j ∈ [1, m−1]
		eL := pairing.AddPair(newSRS.G1Powers[1], newSRS.G2Powers[i]).Result()
		eR := pairing.AddPair(g1.One(), newSRS.G2Powers[i+1]).Result()
		if !eL.Equal(eR) {
			return false
		}
	}

	return true
}

// VerifyState acts similarly to VerifyNewSRSFromPrevSRS, but verifying the
// given State (which can be obtained from the Sequencer)
func VerifyState(s *State) bool {
	pairing := bls12381.NewEngine()

	for _, t := range s.Transcripts {
		// 1. check that elements of the SRS are valid points
		for i := 0; i < len(t.PowersOfTau.G1Powers); i++ {
			if err := checkG1PointCorrectness(t.PowersOfTau.G1Powers[i]); err != nil {
				return false
			}
		}
		for i := 0; i < len(t.PowersOfTau.G2Powers); i++ {
			if err := checkG2PointCorrectness(t.PowersOfTau.G2Powers[i]); err != nil {
				return false
			}
		}

		// 2. check t.Witness.RunningProducts[last] == t.PowersOfTau.G1Powers[1]
		if !g1.Equal(t.Witness.RunningProducts[len(t.Witness.RunningProducts)-1],
			t.PowersOfTau.G1Powers[1]) {
			return false
		}

		// 3. check newSRS.G1s[1] (g₁^τ'), is correctly related to prevSRS.G1s[1] (g₁^τ)
		//   e([τ]₁, [p]₂) == e([τ']₁, [1]₂)
		eL := pairing.AddPair(t.Witness.RunningProducts[len(t.Witness.RunningProducts)-2], t.Witness.PotPubKeys[len(t.Witness.PotPubKeys)-1]).Result()
		eR := pairing.AddPair(t.Witness.RunningProducts[len(t.Witness.RunningProducts)-1], g2.One()).Result()
		if !eL.Equal(eR) {
			return false
		}

		// 4. check newSRS following the powers of tau structure
		for i := 0; i < len(t.PowersOfTau.G1Powers)-1; i++ {
			// i) e([τ'ⁱ]₁, [τ']₂) == e([τ'ⁱ⁺¹]₁, [1]₂), for i ∈ [1, n−1]
			eL := pairing.AddPair(t.PowersOfTau.G1Powers[i], t.PowersOfTau.G2Powers[1]).Result()
			eR := pairing.AddPair(t.PowersOfTau.G1Powers[i+1], g2.One()).Result()
			if !eL.Equal(eR) {
				return false
			}
		}

		for i := 0; i < len(t.PowersOfTau.G2Powers)-1; i++ {
			// ii) e([τ']₁, [τ'ʲ]₂) == e([1]₁, [τ'ʲ⁺¹]₂), for j ∈ [1, m−1]
			eL := pairing.AddPair(t.PowersOfTau.G1Powers[1], t.PowersOfTau.G2Powers[i]).Result()
			eR := pairing.AddPair(g1.One(), t.PowersOfTau.G2Powers[i+1]).Result()
			if !eL.Equal(eR) {
				return false
			}
		}
	}

	return true
}
