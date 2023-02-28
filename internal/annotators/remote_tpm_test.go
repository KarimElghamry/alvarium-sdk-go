package annotators

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/KarimElghamry/alvarium-sdk-go/pkg/config"
	"github.com/KarimElghamry/alvarium-sdk-go/pkg/contracts"
	"github.com/KarimElghamry/alvarium-sdk-go/test"
)

func TestRemoteTpmAnnotator_DO(t *testing.T) {
	b, err := ioutil.ReadFile("../../test/res/config.json")
	if err != nil {
		t.Fatalf(err.Error())
	}

	var cfg config.SdkInfo
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		t.Fatalf(err.Error())
	}

	validAttestData := []byte{
		255, 84, 67, 71, 128, 24, 0, 34, 0, 11, 178, 247, 234, 144, 84,
		235, 92, 112, 28, 229, 84, 210, 2, 184, 175, 218, 188, 204, 2, 139, 114,
		251, 175, 123, 91, 82, 179, 24, 142, 127, 67, 101, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		97, 0, 0, 0, 29, 0, 0, 0, 0, 1, 32, 23, 6, 25, 0, 22, 54, 54, 0, 0, 0, 1, 0, 4,
		3, 128, 0, 0, 0, 32, 222, 71, 201, 178, 126, 184, 211, 0, 219, 181, 242, 195, 83,
		230, 50, 195, 147, 38, 44, 240, 99, 64, 196, 250, 127, 27, 64, 196, 203, 211, 111, 144,
	}

	invalidAttestData := []byte{
		230, 50, 195, 147, 38, 44, 240, 99, 64, 196, 250, 127, 27, 64, 196, 203, 211, 111, 144,
	}

	tests := []struct {
		name          string
		data          []byte
		cfg           config.SdkInfo
		expectedError bool
		isSatisfied   bool
	}{
		{"nil attest data input", nil, cfg, false, false},
		{"empty attest data input", []byte(""), cfg, false, false},
		{"valid attest data input", validAttestData, cfg, false, true},
		{"invalid attest data input", invalidAttestData, cfg, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			remoteTpm := NewRemoteTpmAnnotator(tt.cfg)
			annotation, err := remoteTpm.Do(context.WithValue(context.Background(), contracts.DeviceIdKey, "foo"), tt.data)
			test.CheckError(err, tt.expectedError, tt.name, t)
			if annotation.IsSatisfied != tt.isSatisfied {
				t.Errorf("output annotation isSatisfied is: %t. expected value is: %t", annotation.IsSatisfied, tt.isSatisfied)
			}
		})
	}
}
