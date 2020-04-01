package gocircomprover

import (
	"crypto/rand"
	"math/big"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
)

type Proof struct {
	A *bn256.G1
	B *bn256.G2
	C *bn256.G1
}

type ProvingKey struct {
	A          []*bn256.G1
	B2         []*bn256.G2
	B1         []*bn256.G1
	C          []*bn256.G1
	NVars      int
	NPublic    int
	VkAlpha1   *bn256.G1
	VkDelta1   *bn256.G1
	VkBeta1    *bn256.G1
	VkBeta2    *bn256.G2
	VkDelta2   *bn256.G2
	HExps      []*bn256.G1
	DomainSize int
	PolsA      []map[int]*big.Int
	PolsB      []map[int]*big.Int
	PolsC      []map[int]*big.Int
}

type Witness []*big.Int

var R, _ = new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)

func RandBigInt() (*big.Int, error) {
	maxbits := R.BitLen()
	b := make([]byte, (maxbits/8)-1)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	r := new(big.Int).SetBytes(b)
	rq := new(big.Int).Mod(r, R)

	return rq, nil
}

func Prove(pk *ProvingKey, w Witness) (*Proof, []*big.Int, error) {
	var proof Proof

	r, err := RandBigInt()
	if err != nil {
		return nil, nil, err
	}
	s, err := RandBigInt()
	if err != nil {
		return nil, nil, err
	}

	proof.A = new(bn256.G1).ScalarBaseMult(big.NewInt(0))
	proof.B = new(bn256.G2).ScalarBaseMult(big.NewInt(0))
	proof.C = new(bn256.G1).ScalarBaseMult(big.NewInt(0))
	proofBG1 := new(bn256.G1).ScalarBaseMult(big.NewInt(0))

	for i := 0; i < pk.NVars; i++ {
		proof.A = new(bn256.G1).Add(proof.A, new(bn256.G1).ScalarMult(pk.A[i], w[i]))
		proof.B = new(bn256.G2).Add(proof.B, new(bn256.G2).ScalarMult(pk.B2[i], w[i]))
		proofBG1 = new(bn256.G1).Add(proofBG1, new(bn256.G1).ScalarMult(pk.B1[i], w[i]))
	}

	for i := pk.NPublic + 1; i < pk.NVars; i++ {
		proof.C = new(bn256.G1).Add(proof.C, new(bn256.G1).ScalarMult(pk.C[i], w[i]))
	}

	proof.A = new(bn256.G1).Add(proof.A, pk.VkAlpha1)
	proof.A = new(bn256.G1).Add(proof.A, new(bn256.G1).ScalarMult(pk.VkDelta1, r))

	proof.B = new(bn256.G2).Add(proof.B, pk.VkBeta2)
	proof.B = new(bn256.G2).Add(proof.B, new(bn256.G2).ScalarMult(pk.VkDelta2, s))

	proofBG1 = new(bn256.G1).Add(proofBG1, pk.VkBeta1)
	proofBG1 = new(bn256.G1).Add(proofBG1, new(bn256.G1).ScalarMult(pk.VkDelta1, s))

	// TODO
	// h := calculateH(pk, w)
	h := []*big.Int{} // TMP

	for i := 0; i < len(h); i++ {
		proof.C = new(bn256.G1).Add(proof.C, new(bn256.G1).ScalarMult(pk.HExps[i], h[i]))
	}
	proof.C = new(bn256.G1).Add(proof.C, new(bn256.G1).ScalarMult(proof.A, s))
	proof.C = new(bn256.G1).Add(proof.C, new(bn256.G1).ScalarMult(proofBG1, r))
	rsneg := new(big.Int).Mod(new(big.Int).Neg(new(big.Int).Mul(r, s)), R) // FAdd & FMul
	proof.C = new(bn256.G1).Add(proof.C, new(bn256.G1).ScalarMult(pk.VkDelta1, rsneg))

	pubSignals := w[1 : pk.NPublic+1]

	return &proof, pubSignals, nil
}