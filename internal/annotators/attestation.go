package annotators

import (
	"context"
	"errors"
	"os"

	"github.com/KarimElghamry/alvarium-sdk-go/pkg/config"
	"github.com/KarimElghamry/alvarium-sdk-go/pkg/contracts"
	"github.com/KarimElghamry/alvarium-sdk-go/pkg/interfaces"
)

type AttestationAnnotator struct {
	hash contracts.HashType
	kind contracts.AnnotationType
	sign config.SignatureInfo
}

func NewAttestationAnnotator(cfg config.SdkInfo) interfaces.Annotator {
	a := AttestationAnnotator{}
	a.hash = cfg.Hash.Type
	a.kind = contracts.AnnotationAttestation
	a.sign = cfg.Signature
	return &a
}

func (a *AttestationAnnotator) Do(ctx context.Context, data []byte) (contracts.Annotation, error) {
	deviceId, ok := ctx.Value(contracts.DeviceIdKey).(string)
	if !ok {
		return contracts.Annotation{}, errors.New("`deviceId` not found in attestation annotator's context")
	}

	key := DeriveHash(a.hash, data)
	hostname, _ := os.Hostname()

	annotation := contracts.NewAnnotation(key, a.hash, hostname, a.kind, true)
	sig, err := SignAnnotation(a.sign.PrivateKey, annotation)

	if err != nil {
		return contracts.Annotation{}, err
	}

	annotation.DeviceId = deviceId
	annotation.Signature = string(sig)
	return annotation, nil
}
