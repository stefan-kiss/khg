// Copyright (c) 2021. Stefan Kiss
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package kubeconfig

import (
	"github.com/k0kubun/pp"
	"github.com/stefan-kiss/khg/internal/cfg"
	"net/url"
	"testing"
)

var kubeValidDst = KubeConfig{
	Url: &url.URL{
		Scheme: "file",
		Path:   "../../test/kubeconfig/validConfig.yaml",
	},
}

var kubeValidSrc = KubeConfig{
	Url: &url.URL{
		Scheme: "file",
		Path:   "../../test/kubeconfig/config.src.yaml",
	},
	Label: "externalTest",
	SrcDef: cfg.Source{
		Insecure:   true,
		ApiAddress: "1.1.1.1:443",
	},
}

func TestKubeConfig_ReadConfig(t *testing.T) {
	tests := []struct {
		name    string
		k       KubeConfig
		wantErr bool
	}{
		{
			name:    "ValidSource",
			k:       kubeValidDst,
			wantErr: false,
		},
		{
			name:    "ValidSource",
			k:       kubeValidSrc,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.k
			if err := k.ReadConfig(); (err != nil) != tt.wantErr {
				t.Errorf("ReadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			pp.Println(k.Config)
		})
	}
}

func TestKubeConfig_CopyCurrentContext(t *testing.T) {
	tests := []struct {
		name      string
		labelKube string
		dest      *KubeConfig
		src       *KubeConfig
		wantErr   bool
	}{
		{
			name:      "SuccessTest",
			labelKube: "external",
			src:       &kubeValidSrc,
			dest:      &kubeValidDst,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			err = tt.src.ReadConfig()
			if err != nil {
				t.Errorf("ReadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			err = tt.dest.ReadConfig()
			if err != nil {
				t.Errorf("ReadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}

			k := tt.dest
			if err := k.CopyCurrentContext(tt.src); (err != nil) != tt.wantErr {
				t.Errorf("CopyCurrentContext() error = %v, wantErr %v", err, tt.wantErr)
			}
			pp.Println(k)
		})
	}
}
