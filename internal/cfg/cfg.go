// Copyright (c) 2020. Stefan Kiss
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

package cfg

import (
	"fmt"
	"github.com/goccy/go-yaml"
	"github.com/spf13/viper"
	"io/ioutil"
)

type Source struct {
	Source        string `yaml:"source"`
	Insecure      bool   `yaml:"insecure"`
	ApiAddress    string `yaml:"apiaddress"`
	AutodetectApi bool   `yaml:"-"`
	OverrideIp    string `yaml:"-"`
}

type Cfg struct {
	Sources           map[string]Source `yaml:"sources"`
	Destination       string            `yaml:"destination"`
	DefaultSourcePath string
}

func Add(config *Cfg, label string, source Source) error {
	config.Sources[label] = source
	return Save(config)
}

func Delete(config *Cfg, label string) error {
	if _, ok := config.Sources[label]; ok {
		delete(config.Sources, label)
	} else {
		return fmt.Errorf("label: %s not found in config", label)
	}
	return nil
}

func Save(config *Cfg) error {
	configBytes, err := yaml.Marshal(*config)
	if err != nil {
		return fmt.Errorf("unable marshal the config file: %v", err)
	}

	err = ioutil.WriteFile(viper.ConfigFileUsed(), configBytes, 600)
	if err != nil {
		return fmt.Errorf("unable write the config file %s: %v", viper.ConfigFileUsed(), err)
	}
	return nil
}
