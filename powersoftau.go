package kzgceremony

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto/bls12381"
)

type Contribution struct {
	SRS   *SRS
	Proof *Proof
}

type SRS struct {
	G1s []*bls12381.PointG1
	G2s []*bls12381.PointG2
}

type toxicWaste struct {
	tau   *big.Int
	TauG2 *bls12381.PointG2
}

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
	TauG2 := g2.New()
	g2.MulScalar(TauG2, g2.One(), tau)

	return &toxicWaste{tau, TauG2}
}

func computeContribution(t *toxicWaste, prevSRS *SRS) *SRS {
	srs := newEmptySRS(len(prevSRS.G1s), len(prevSRS.G2s))
	g1 := bls12381.NewG1()
	g2 := bls12381.NewG2()
	Q := g1.Q() // Q = |G1| == |G2|

	fmt.Println("Computing [τ'⁰]₁, [τ'¹]₁, [τ'²]₁, ..., [τ'ⁿ⁻¹]₁, for n =", len(prevSRS.G1s))
	for i := 0; i < len(prevSRS.G1s); i++ {
		tau_i := new(big.Int).Exp(t.tau, big.NewInt(int64(i)), Q)
		g1.MulScalar(srs.G1s[i], prevSRS.G1s[i], tau_i)
	}
	fmt.Println("Computing [τ'⁰]₂, [τ'¹]₂, [τ'²]₂, ..., [τ'ⁿ⁻¹]₂, for n =", len(prevSRS.G2s))
	for i := 0; i < len(prevSRS.G2s); i++ {
		tau_i := new(big.Int).Exp(t.tau, big.NewInt(int64(i)), Q)
		g2.MulScalar(srs.G2s[i], prevSRS.G2s[i], tau_i)
	}

	return srs
}

func genProof(toxicWaste *toxicWaste, prevSRS, newSRS *SRS) *Proof {
	g1 := bls12381.NewG1()
	G1_p := g1.New()
	g1.MulScalar(G1_p, prevSRS.G1s[1], toxicWaste.tau) // g_1^{tau'} = g_1^{p * tau}, where p=toxicWaste.tau

	return &Proof{toxicWaste.TauG2, G1_p}
}

// Contribute
func Contribute(prevSRS *SRS, randomness []byte) (Contribution, error) {
	// set tau from randomness
	tw := tau(randomness)

	newSRS := computeContribution(tw, prevSRS)

	proof := genProof(tw, prevSRS, newSRS)

	return Contribution{SRS: newSRS, Proof: proof}, nil
}

func Verify(prevSRS, newSRS *SRS, proof *Proof) bool {
	g1 := bls12381.NewG1()

	// check proof.G1PTau == newSRS.G1s[1]
	if !g1.Equal(proof.G1PTau, newSRS.G1s[1]) {
		return false
	}

	// WIP!
	return true
}
