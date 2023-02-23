/*******************************************************************************
 * Copyright 2021 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/
package annotators

import (
	"context"
	"os"

	"github.com/KarimElghamry/alvarium-sdk-go/pkg/config"
	"github.com/KarimElghamry/alvarium-sdk-go/pkg/contracts"
	"github.com/KarimElghamry/alvarium-sdk-go/pkg/interfaces"
	"github.com/google/go-tpm/tpm2"
)

// annotator to check on the TPM attestation quote coming from a remote device
type RemoteTpmAnnotator struct {
	hash contracts.HashType
	kind contracts.AnnotationType
	sign config.SignatureInfo
}

func NewRemoteTpmAnnotator(cfg config.SdkInfo) interfaces.Annotator {
	a := RemoteTpmAnnotator{}
	a.hash = cfg.Hash.Type
	a.kind = contracts.AnnotationRemoteTPM
	a.sign = cfg.Signature
	return &a
}

func (a *RemoteTpmAnnotator) Do(ctx context.Context, data []byte) (contracts.Annotation, error) {
	key := DeriveHash(a.hash, data)
	hostname, _ := os.Hostname()
	isSatisfied := validateTpmQuote(data)

	annotation := contracts.NewAnnotation(key, a.hash, hostname, a.kind, isSatisfied)
	sig, err := SignAnnotation(a.sign.PrivateKey, annotation)
	if err != nil {
		return contracts.Annotation{}, err
	}
	annotation.Signature = string(sig)
	return annotation, nil
}

func validateTpmQuote(data []byte) bool {
	if data == nil {
		return false
	}

	_, err := tpm2.DecodeAttestationData(data)
	if err != nil {
		return false
	}

	return true
}
