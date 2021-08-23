package presign

import (
	"errors"

	"github.com/taurusgroup/multi-party-sig/internal/round"
	"github.com/taurusgroup/multi-party-sig/pkg/ecdsa"
	"github.com/taurusgroup/multi-party-sig/pkg/math/curve"
	"github.com/taurusgroup/multi-party-sig/pkg/party"
	zkelog "github.com/taurusgroup/multi-party-sig/pkg/zk/elog"
	zklog "github.com/taurusgroup/multi-party-sig/pkg/zk/log"
)

var _ round.Round = (*presign7)(nil)

type presign7 struct {
	*presign6
	// Delta = δ = ∑ⱼ δⱼ
	Delta curve.Scalar

	// S[j] = Sⱼ
	S map[party.ID]curve.Point

	// R = [δ⁻¹] Γ
	R curve.Point

	// RBar = {R̄ⱼ = δ⁻¹⋅Δⱼ}ⱼ
	RBar map[party.ID]curve.Point
}

type message7 struct {
	// S = Sᵢ
	S     curve.Point
	Proof *zkelog.Proof
}

// VerifyMessage implements round.Round.
func (r *presign7) VerifyMessage(msg round.Message) error {
	from := msg.From
	body, ok := msg.Content.(*message7)
	if !ok || body == nil {
		return round.ErrInvalidContent
	}

	if body.S.IsIdentity() {
		return round.ErrNilFields
	}

	if !body.Proof.Verify(r.HashForID(from), zkelog.Public{
		E:             r.ElGamalChi[from],
		ElGamalPublic: r.ElGamal[from],
		Base:          r.R,
		Y:             body.S,
	}) {
		return errors.New("failed to validate elog proof for S")
	}

	return nil
}

// StoreMessage implements round.Round.
//
// - save Sⱼ
func (r *presign7) StoreMessage(msg round.Message) error {
	from, body := msg.From, msg.Content.(*message7)
	r.S[from] = body.S
	return nil
}

// Finalize implements round.Round
//
// - verify ∑ⱼ Sⱼ = X
func (r *presign7) Finalize(out chan<- *round.Message) (round.Round, error) {
	// compute ∑ⱼ Sⱼ
	PublicKeyComputed := r.Group().NewPoint()
	for _, Sj := range r.S {
		PublicKeyComputed = PublicKeyComputed.Add(Sj)
	}

	// ∑ⱼ Sⱼ ?= X
	if !r.PublicKey.Equal(PublicKeyComputed) {
		YHat := r.ElGamalKNonce.Act(r.ElGamal[r.SelfID()])
		YHatProof := zklog.NewProof(r.Group(), r.HashForID(r.SelfID()), zklog.Public{
			H: r.ElGamalKNonce.ActOnBase(),
			X: r.ElGamal[r.SelfID()],
			Y: YHat,
		}, zklog.Private{
			A: r.ElGamalKNonce,
			B: r.SecretElGamal,
		})

		ChiProofs := make(map[party.ID]*abortNth, r.N()-1)
		for _, j := range r.OtherPartyIDs() {
			chiCiphertext := r.ChiCiphertext[j][r.SelfID()] // D̂ᵢⱼ
			ChiProofs[j] = proveNth(r.HashForID(r.SelfID()), r.SecretPaillier, chiCiphertext)
		}
		msg := &messageAbort2{
			YHat:      YHat,
			YHatProof: YHatProof,
			KProof:    proveNth(r.HashForID(r.SelfID()), r.SecretPaillier, r.K[r.SelfID()]),
			ChiProofs: ChiProofs,
		}
		if err := r.SendMessage(out, msg, ""); err != nil {
			return r, err
		}
		ChiAlphas := make(map[party.ID]curve.Scalar, r.N())
		for id, chiAlpha := range r.ChiShareAlpha {
			ChiAlphas[id] = r.Group().NewScalar().SetNat(chiAlpha.Mod(r.Group().Order()))
		}
		return &abort2{
			presign7:  r,
			YHat:      map[party.ID]curve.Point{r.SelfID(): YHat},
			KShares:   map[party.ID]curve.Scalar{r.SelfID(): r.KShare},
			ChiAlphas: map[party.ID]map[party.ID]curve.Scalar{r.SelfID(): ChiAlphas},
		}, nil
	}

	preSignature := &ecdsa.PreSignature{
		Group:    r.Group(),
		R:        r.R,
		RBar:     r.RBar,
		S:        r.S,
		KShare:   r.KShare,
		ChiShare: r.ChiShare,
	}
	if r.Message == nil {
		return &round.Output{Result: preSignature}, nil
	}

	rSign1 := &sign1{
		Helper:       r.Helper,
		PublicKey:    r.PublicKey,
		Message:      r.Message,
		PreSignature: preSignature,
	}
	nextRound, err := rSign1.Finalize(out)
	rSign2, ok := nextRound.(*sign2)
	if !ok || err != nil {
		return r, err
	}
	return &presign8{rSign2}, nil
}

// MessageContent implements round.Round.
func (presign7) MessageContent() round.Content { return &message7{} }

// Number implements round.Round.
func (presign7) Number() round.Number { return 7 }

// Init implements round.Content.
func (m *message7) Init(group curve.Curve) {
	m.S = group.NewPoint()
	m.Proof = zkelog.Empty(group)
}
