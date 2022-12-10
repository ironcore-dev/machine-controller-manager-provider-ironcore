// Copyright 2022 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ignition

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/Masterminds/sprig"
	buconfig "github.com/coreos/butane/config"
	"github.com/coreos/butane/config/common"
	"github.com/imdario/mergo"
	"sigs.k8s.io/yaml"
	"text/template"
)

var (
	//go:embed template.yaml
	IgnitionTemplate string
)

type Config struct {
	Hostname         string
	UserData         string
	Ignition         string
	IgnitionOverride bool
}

func File(config *Config) (string, error) {

	ignitionBase := &map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(IgnitionTemplate), ignitionBase); err != nil {
		return "", err
	}
	
	// if ignition was set in providerSpec merge it with our template
	if config.Ignition != "" {
		additional := map[string]interface{}{}

		if err := yaml.Unmarshal([]byte(config.Ignition), &additional); err != nil {
			return "", err
		}

		// default to append ignition
		opt := mergo.WithAppendSlice

		// allow also to fully override
		if config.IgnitionOverride {
			opt = mergo.WithOverride
		}

		// merge both ignitions
		err := mergo.Merge(ignitionBase, additional, opt)
		if err != nil {
			return "", err
		}
	}

	mergedIgnition, err := yaml.Marshal(ignitionBase)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("ignition").Funcs(sprig.HermeticTxtFuncMap()).Parse(string(mergedIgnition))
	if err != nil {
		return "", fmt.Errorf("failed creating ignition file: %w", err)
	}
	buf := bytes.NewBufferString("")
	err = tmpl.Execute(buf, config)
	if err != nil {
		return "", fmt.Errorf("failed creating ignition file while executing template: %w", err)
	}

	ignition, err := renderButane(buf.Bytes())
	if err != nil {
		return "", err
	}

	return ignition, nil
}
func renderButane(dataIn []byte) (string, error) {
	// render by butane to json
	options := common.TranslateBytesOptions{
		Raw:    true,
		Pretty: false,
	}
	options.NoResourceAutoCompression = true
	dataOut, _, err := buconfig.TranslateBytes(dataIn, options)
	if err != nil {
		return "", err
	}
	return string(dataOut), nil
}
