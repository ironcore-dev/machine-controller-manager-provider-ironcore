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
	"encoding/base64"
	"fmt"
	"strings"
	"text/template"
)

type Config struct {
	PasswdHash     string
	Hostname       string
	SSHKeys        []string
	UserdataBase64 string
	InstallPath    string
}

func File(config *Config) (string, error) {
	ignitionTemplate := `
{
  "ignition": {"config":{},"timeouts":{},"version":"3.0.0"},
  "networkd":{"units":[{"contents":"[Match]\nName=ens192\n\n[Network]\nDHCP=no\nLinkLocalAddressing=no\nIPv6AcceptRA=no\n","name":"00-ens192.network"}]},
  "passwd":{"users":[{"name":"core","passwordHash":"{{.PasswdHash}}","sshAuthorizedKeys":[{{range $index,$elem := .SSHKeys}}{{if $index}},{{end}}"{{$elem}}"{{end}}]}]},
  "storage": {
	"directories":[{"filesystem":"root","path":"{{.InstallPath}}","mode":493}],
	"files":[
	  {"filesystem":"root","path":"/etc/hostname","contents":{"source":"data:,{{.Hostname}}"},"mode":420},
	  {"filesystem":"root","path":"{{.InstallPath}}/user_data","contents":{"source":"data:text/plain;charset=utf-8;base64,{{.UserdataBase64}}"},"mode":420}
	]
  },
  "systemd":{}
}
`
	tmpl, err := template.New("ignition").Parse(ignitionTemplate)
	if err != nil {
		return "", fmt.Errorf("failed creating ignition file: %w", err)
	}
	buf := bytes.NewBufferString("")
	err = tmpl.Execute(buf, config)
	if err != nil {
		return "", fmt.Errorf("failed creating ignition file while executing template: %w", err)
	}
	return buf.String(), nil
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
