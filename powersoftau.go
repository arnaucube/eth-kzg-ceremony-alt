package kzgceremony

import (
	"math/big"

	bls12381 "github.com/kilic/bls12-381"
)

// todo: unify addition & multiplicative notation in the comments

// Contribution contains the SRS with its Proof
type Contribution struct {
	SRS   *SRS
	Proof *Proof
}

// SRS contains the powers of tau in G1 & G2, eg.
// [τ'⁰]₁, [τ'¹]₁, [τ'²]₁, ..., [τ'ⁿ⁻¹]₁,
// [τ'⁰]₂, [τ'¹]₂, [τ'²]₂, ..., [τ'ⁿ⁻¹]₂
type SRS struct {
	G1s []*bls12381.PointG1
	G2s []*bls12381.PointG2
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
	srs := newEmptySRS(len(prevSRS.G1s), len(prevSRS.G2s))
	g1 := bls12381.NewG1()
	g2 := bls12381.NewG2()
	Q := g1.Q() // Q = |G1| == |G2|

	// fmt.Println("Computing [τ'⁰]₁, [τ'¹]₁, [τ'²]₁, ..., [τ'ⁿ⁻¹]₁, for n =", len(prevSRS.G1s))
	for i := 0; i < len(prevSRS.G1s); i++ {
		tau_i := new(big.Int).Exp(t.tau, big.NewInt(int64(i)), Q)
		tau_i_Fr := bls12381.NewFr().FromBytes(tau_i.Bytes())
		g1.MulScalar(srs.G1s[i], prevSRS.G1s[i], tau_i_Fr)
	}
	// fmt.Println("Computing [τ'⁰]₂, [τ'¹]₂, [τ'²]₂, ..., [τ'ⁿ⁻¹]₂, for n =", len(prevSRS.G2s))
	for i := 0; i < len(prevSRS.G2s); i++ {
		tau_i := new(big.Int).Exp(t.tau, big.NewInt(int64(i)), Q)
		tau_i_Fr := bls12381.NewFr().FromBytes(tau_i.Bytes())
		g2.MulScalar(srs.G2s[i], prevSRS.G2s[i], tau_i_Fr)
	}

	return srs
}

func genProof(toxicWaste *toxicWaste, prevSRS, newSRS *SRS) *Proof {
	g1 := bls12381.NewG1()
	G1_p := g1.New()
	tau_Fr := bls12381.NewFr().FromBytes(toxicWaste.tau.Bytes())
	g1.MulScalar(G1_p, prevSRS.G1s[1], tau_Fr) // g_1^{tau'} = g_1^{p * tau}, where p=toxicWaste.tau

	return &Proof{toxicWaste.TauG2, G1_p}
}

// Contribute takes as input the previous SRS and a random byte slice, and
// returns the new SRS together with the Proof
func Contribute(prevSRS *SRS, randomness []byte) (Contribution, error) {
	// set tau from randomness
	tw := tau(randomness)

	newSRS := computeContribution(tw, prevSRS)

	proof := genProof(tw, prevSRS, newSRS)

	return Contribution{SRS: newSRS, Proof: proof}, nil
}

// Verify checks the correct computation of the new SRS respectively from the
// previous SRS
func Verify(prevSRS, newSRS *SRS, proof *Proof) bool {
	g1 := bls12381.NewG1()
	g2 := bls12381.NewG2()
	pairing := bls12381.NewEngine()

	// 1. check that elements of the newSRS are valid points
	for i := 0; i < len(newSRS.G1s); i++ {
		// i) non-empty
		if newSRS.G1s[i] == nil {
			return false
		}
		// ii) non-zero
		if g1.IsZero(newSRS.G1s[i]) {
			return false
		}
		// iii) in the correct prime order of subgroups
		if !g1.IsOnCurve(newSRS.G1s[i]) {
			return false
		}
		if !g1.InCorrectSubgroup(newSRS.G1s[i]) {
			return false
		}
	}
	for i := 0; i < len(newSRS.G2s); i++ {
		// i) non-empty
		if newSRS.G2s[i] == nil {
			return false
		}
		// ii) non-zero
		if g2.IsZero(newSRS.G2s[i]) {
			return false
		}
		// iii) in the correct prime order of subgroups
		if !g2.IsOnCurve(newSRS.G2s[i]) {
			return false
		}
		if !g2.InCorrectSubgroup(newSRS.G2s[i]) {
			return false
		}
	}

	// 2. check proof.G1PTau == newSRS.G1s[1]
	if !g1.Equal(proof.G1PTau, newSRS.G1s[1]) {
		return false
	}

	// 3. check newSRS.G1s[1] (g₁^τ'), is correctly related to prevSRS.G1s[1] (g₁^τ)
	//   e([τ]₁, [p]₂) == e([τ']₁, [1]₂)
	e0 := pairing.AddPair(prevSRS.G1s[1], proof.G2P).Result()
	e1 := pairing.AddPair(newSRS.G1s[1], g2.One()).Result()
	if !e0.Equal(e1) {
		return false
	}

	// 4. check newSRS following the powers of tau structure
	for i := 0; i < len(newSRS.G1s)-1; i++ {
		// i) e([τ'ⁱ]₁, [τ']₂) == e([τ'ⁱ⁺¹]₁, [1]₂), for i ∈ [1, n−1]
		e0 := pairing.AddPair(newSRS.G1s[i], newSRS.G2s[1]).Result()
		e1 := pairing.AddPair(newSRS.G1s[i+1], g2.One()).Result()
		if !e0.Equal(e1) {
			return false
		}
	}

	for i := 0; i < len(newSRS.G2s)-1; i++ {
		// ii) e([τ']₁, [τ'ʲ]₂) == e([1]₁, [τ'ʲ⁺¹]₂), for j ∈ [1, m−1]
		e3 := pairing.AddPair(newSRS.G1s[1], newSRS.G2s[i]).Result()
		e4 := pairing.AddPair(g1.One(), newSRS.G2s[i+1]).Result()
		if !e3.Equal(e4) {
			return false
		}
	}

	return true
}
