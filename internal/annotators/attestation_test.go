package annotators

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/KarimElghamry/alvarium-sdk-go/pkg/config"
	"github.com/KarimElghamry/alvarium-sdk-go/pkg/contracts"
)

func TestAttestationAnnotation(t *testing.T) {
	b, err := ioutil.ReadFile("../../test/res/config.json")
	if err != nil {
		t.Fatalf(err.Error())
	}
	var cfg config.SdkInfo
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		t.Fatalf(err.Error())
	}

	tests := []struct {
		name        string
		data        string
		cfg         config.SdkInfo
		expectError bool
	}{
		{"check is satisfied", "foo", cfg, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attest := NewAttestationAnnotator(cfg)
			a, err := attest.Do(context.WithValue(context.Background(), contracts.DeviceIdKey, "foo"), []byte(tt.data))
			if err != nil {
				t.Fatalf(err.Error())
			}
			if !a.IsSatisfied {
				t.Error("Attestation Annotation's isSatisfied is not true")
			}
			if a.DeviceId != "foo" {
				t.Error("Invalid device id")
			}
		})
	}
}
