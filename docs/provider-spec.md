## Specification
### ProviderSpec Schema
<br>
<h3 id="settings.gardener.cloud/v1alpha1.ProviderSpec">
<b>ProviderSpec</b>
</h3>
<p>
<p>ProviderSpec is the spec to be used while parsing the calls</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Type</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>image</code>
</td>
<td>
<em>
string
</em>
</td>
<td>
<p>Image is the URL pointing to an OCI registry containing the operating system image which should be used to boot the Machine</p>
</td>
</tr>
<tr>
<td>
<code>ignition</code>
</td>
<td>
<em>
string
</em>
</td>
<td>
<p>Ignition contains the ignition configuration which should be run on first boot of a Machine.</p>
</td>
</tr>
<tr>
<td>
<code>ignitionOverride</code>
</td>
<td>
<em>
bool
</em>
</td>
<td>
<p>By default, if ignition is set it will be merged it with our template
If IgnitionOverride is set to true allows to fully override</p>
</td>
</tr>
<tr>
<td>
<code>ignitionSecretKey</code>
</td>
<td>
<em>
string
</em>
</td>
<td>
<p>IgnitionSecretKey is optional key field used to identify the ignition content in the Secret
If the key is empty, the DefaultIgnitionKey will be used as fallback.</p>
</td>
</tr>
<tr>
<td>
<code>rootDisk</code>
</td>
<td>
<em>
<a href="#?id=%23settings.gardener.cloud%2fv1alpha1.RootDisk">
RootDisk
</a>
</em>
</td>
<td>
<p>RootDisk defines the root disk properties of the Machine.</p>
</td>
</tr>
<tr>
<td>
<code>networkName</code>
</td>
<td>
<em>
string
</em>
</td>
<td>
<p>NetworkName is the Network to be used for the Machine&rsquo;s NetworkInterface.</p>
</td>
</tr>
<tr>
<td>
<code>prefixName</code>
</td>
<td>
<em>
string
</em>
</td>
<td>
<p>PrefixName is the parent Prefix from which an IP should be allocated for the Machine&rsquo;s NetworkInterface.</p>
</td>
</tr>
<tr>
<td>
<code>labels</code>
</td>
<td>
<em>
map[string]string
</em>
</td>
<td>
<p>Labels are used to tag resources which the MCM creates, so they can be identified later.</p>
</td>
</tr>
<tr>
<td>
<code>dnsServers</code>
</td>
<td>
<em>
<a href="#?id=https%3a%2f%2fpkg.go.dev%2fnet%2fnetip%23Addr">
[]net/netip.Addr
</a>
</em>
</td>
<td>
<p>DnsServers is a list of DNS resolvers which should be configured on the host.</p>
</td>
</tr>
</tbody>
</table>
<br>
<h3 id="settings.gardener.cloud/v1alpha1.RootDisk">
<b>RootDisk</b>
</h3>
<p>
(<em>Appears on:</em>
<a href="#?id=%23settings.gardener.cloud%2fv1alpha1.ProviderSpec">ProviderSpec</a>)
</p>
<p>
<p>RootDisk defines the root disk properties of the Machine.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Type</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>size</code>
</td>
<td>
<em>
<a href="#?id=https%3a%2f%2fpkg.go.dev%2fk8s.io%2fapimachinery%2fpkg%2fapi%2fresource%23Quantity">
k8s.io/apimachinery/pkg/api/resource.Quantity
</a>
</em>
</td>
<td>
<p>Size defines the volume size of the root disk.</p>
</td>
</tr>
<tr>
<td>
<code>volumeClassName</code>
</td>
<td>
<em>
string
</em>
</td>
<td>
<p>VolumeClassName defines which volume class to use for the root disk.</p>
</td>
</tr>
<tr>
<td>
<code>volumePoolName</code>
</td>
<td>
<em>
string
</em>
</td>
<td>
<p>VolumePoolName defines on which VolumePool a Volume should be scheduled.</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <a href="https://github.com/ahmetb/gen-crd-api-reference-docs">gen-crd-api-reference-docs</a>
</em></p>
