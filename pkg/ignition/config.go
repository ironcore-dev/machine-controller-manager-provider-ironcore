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
	"encoding/base64"
	"fmt"
	"github.com/Masterminds/sprig"
	buconfig "github.com/coreos/butane/config"
	"github.com/coreos/butane/config/common"
	"strings"
	"text/template"
)

var (
	//go:embed template.yaml
	IgnitionTemplate string
)

type Config struct {
	Hostname string
	UserData string
}

func File(config *Config) (string, error) {
	tmpl, err := template.New("ignition").Funcs(sprig.HermeticTxtFuncMap()).Parse(IgnitionTemplate)
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

func PrepareUserData(userdata string, sshKeys []string) (string, error) {
	s := userdata
	if strings.HasPrefix(userdata, "#!/") {
		// assume it's a shell script and the ssh keys are appended directly to the authorized keys
		s = packageInCloudInit(userdata)
	}
	return addSSHKeysSection(s, sshKeys)
}

func packageInCloudInit(userdata string) string {
	content := base64.StdEncoding.EncodeToString([]byte(userdata))
	rewrittenUserdata := fmt.Sprintf(`#cloud-config
write_files:
- encoding: b64
  content: %s
  owner: root:root
  path: /root/cloud-init-script
  permissions: '0555'
runcmd:
- /root/cloud-init-script
- rm /root/cloud-init-script
`, content)
	return rewrittenUserdata
}

func addSSHKeysSection(userdata string, sshKeys []string) (string, error) {
	if len(sshKeys) == 0 {
		return userdata, nil
	}
	s := userdata
	if strings.Contains(s, "ssh_authorized_keys:") {
		return "", fmt.Errorf("userdata already contains key `ssh_authorized_keys`")
	}
	s = s + "\nssh_authorized_keys:\n"
	for _, key := range sshKeys {
		s = s + fmt.Sprintf("- %q\n", key)
	}
	return s, nil
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
