package kzgceremony

import (
	"fmt"
	"math/big"

	bls12381 "github.com/kilic/bls12-381"
)

// todo: unify addition & multiplicative notation in the comments

type Witness struct {
	RunningProducts []*bls12381.PointG1
	PotPubKeys      []*bls12381.PointG2
	BLSSignatures   []*bls12381.PointG1
}

type Transcript struct {
	NumG1Powers uint64
	NumG2Powers uint64
	PowersOfTau *SRS
	Witness     *Witness
}

type State struct {
	Transcripts                []Transcript
	ParticipantIDs             []string // WIP
	ParticipantECDSASignatures []string
}

func (cs *State) Contribute(randomness []byte) (*State, error) {
	ns := State{}
	ns.Transcripts = make([]Transcript, len(cs.Transcripts))
	for i := 0; i < len(cs.Transcripts); i++ {
		ns.Transcripts[i].NumG1Powers = cs.Transcripts[i].NumG1Powers
		ns.Transcripts[i].NumG2Powers = cs.Transcripts[i].NumG2Powers

		newSRS, proof, err := Contribute(cs.Transcripts[i].PowersOfTau, randomness)
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
	ns.ParticipantIDs = cs.ParticipantIDs
	ns.ParticipantECDSASignatures = cs.ParticipantECDSASignatures

	return &ns, nil
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
	TauG2 *bls12381.PointG2
}

// Proof contains g₂ᵖ and g₂^τ', used by the verifier
type Proof struct {
	G2P    *bls12381.PointG2 // g₂ᵖ
	G1PTau *bls12381.PointG1 // g₂^τ' = g₂^{p ⋅ τ}
}

// newEmptySRS creates an empty SRS
func newEmptySRS(nG1, nG2 int) *SRS {
	g1s := make([]*bls12381.PointG1, nG1)
	g2s := make([]*bls12381.PointG2, nG2)
	g1 := bls12381.NewG1()
	g2 := bls12381.NewG2()
	// one_G1 := g1.One()
	// one_G2 := g2.One()
	for i := 0; i < nG1; i++ {
		g1s[i] = g1.One()
		// g1.MulScalar(g1s[i], one_G1, big.NewInt(int64(i)))
	}
	for i := 0; i < nG2; i++ {
		g2s[i] = g2.One()
		// g2.MulScalar(g2s[i], one_G2, big.NewInt(int64(i)))
	}
	return &SRS{g1s, g2s}
}

func tau(randomness []byte) *toxicWaste {
	g2 := bls12381.NewG2()
	tau := new(big.Int).Mod(
		new(big.Int).SetBytes(randomness),
		g2.Q())
	tau_Fr := bls12381.NewFr().FromBytes(tau.Bytes())
	TauG2 := g2.New()
	g2.MulScalar(TauG2, g2.One(), tau_Fr)

	return &toxicWaste{tau, TauG2}
}

func computeContribution(t *toxicWaste, prevSRS *SRS) *SRS {
	srs := newEmptySRS(len(prevSRS.G1Powers), len(prevSRS.G2Powers))
	g1 := bls12381.NewG1()
	g2 := bls12381.NewG2()
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
	g1 := bls12381.NewG1()
	G1_p := g1.New()
	tau_Fr := bls12381.NewFr().FromBytes(toxicWaste.tau.Bytes())
	g1.MulScalar(G1_p, prevSRS.G1Powers[1], tau_Fr) // g_1^{tau'} = g_1^{p * tau}, where p=toxicWaste.tau

	return &Proof{toxicWaste.TauG2, G1_p}
}

// Contribute takes as input the previous SRS and a random
// byte slice, and returns the new SRS together with the Proof
func Contribute(prevSRS *SRS, randomness []byte) (*SRS, *Proof, error) {
	if len(randomness) < 64 {
		return nil, nil, fmt.Errorf("err randomness") // WIP
	}
	// set tau from randomness
	tw := tau(randomness)

	newSRS := computeContribution(tw, prevSRS)

	proof := genProof(tw, prevSRS, newSRS)

	return newSRS, proof, nil
}

// Verify checks the correct computation of the new SRS respectively from the
// previous SRS
func Verify(prevSRS, newSRS *SRS, proof *Proof) bool {
	g1 := bls12381.NewG1()
	g2 := bls12381.NewG2()
	pairing := bls12381.NewEngine()

	// 1. check that elements of the newSRS are valid points
	for i := 0; i < len(newSRS.G1Powers); i++ {
		// i) non-empty
		if newSRS.G1Powers[i] == nil {
			return false
		}
		// ii) non-zero
		if g1.IsZero(newSRS.G1Powers[i]) {
			return false
		}
		// iii) in the correct prime order of subgroups
		if !g1.IsOnCurve(newSRS.G1Powers[i]) {
			return false
		}
		if !g1.InCorrectSubgroup(newSRS.G1Powers[i]) {
			return false
		}
	}
	for i := 0; i < len(newSRS.G2Powers); i++ {
		// i) non-empty
		if newSRS.G2Powers[i] == nil {
			return false
		}
		// ii) non-zero
		if g2.IsZero(newSRS.G2Powers[i]) {
			return false
		}
		// iii) in the correct prime order of subgroups
		if !g2.IsOnCurve(newSRS.G2Powers[i]) {
			return false
		}
		if !g2.InCorrectSubgroup(newSRS.G2Powers[i]) {
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
