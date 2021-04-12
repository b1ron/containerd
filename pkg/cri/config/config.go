/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package config

import (
	"context"
	"net/url"
	"time"

	"github.com/containerd/containerd/log"
	"github.com/containerd/containerd/plugin"
	"github.com/pkg/errors"
)

// Runtime struct to contain the type(ID), engine, and root variables for a default runtime
// and a runtime for untrusted worload.
type Runtime struct {
	// Type is the runtime type to use in containerd e.g. io.containerd.runtime.v1.linux
	Type string `toml:"runtime_type" json:"runtimeType"`
	// Engine is the name of the runtime engine used by containerd.
	// This only works for runtime type "io.containerd.runtime.v1.linux".
	// DEPRECATED: use Options instead. Remove when shim v1 is deprecated.
	Engine string `toml:"runtime_engine" json:"runtimeEngine"`
	// PodAnnotations is a list of pod annotations passed to both pod sandbox as well as
	// container OCI annotations.
	PodAnnotations []string `toml:"pod_annotations" json:"PodAnnotations"`
	// ContainerAnnotations is a list of container annotations passed through to the OCI config of the containers.
	// Container annotations in CRI are usually generated by other Kubernetes node components (i.e., not users).
	// Currently, only device plugins populate the annotations.
	ContainerAnnotations []string `toml:"container_annotations" json:"ContainerAnnotations"`
	// Root is the directory used by containerd for runtime state.
	// DEPRECATED: use Options instead. Remove when shim v1 is deprecated.
	// This only works for runtime type "io.containerd.runtime.v1.linux".
	Root string `toml:"runtime_root" json:"runtimeRoot"`
	// Options are config options for the runtime.
	// If options is loaded from toml config, it will be map[string]interface{}.
	// Options can be converted into toml.Tree using toml.TreeFromMap().
	// Using options type as map[string]interface{} helps in correctly marshaling options from Go to JSON.
	Options map[string]interface{} `toml:"options" json:"options"`
	// PrivilegedWithoutHostDevices overloads the default behaviour for adding host devices to the
	// runtime spec when the container is privileged. Defaults to false.
	PrivilegedWithoutHostDevices bool `toml:"privileged_without_host_devices" json:"privileged_without_host_devices"`
	// BaseRuntimeSpec is a json file with OCI spec to use as base spec that all container's will be created from.
	BaseRuntimeSpec string `toml:"base_runtime_spec" json:"baseRuntimeSpec"`
}

// ContainerdConfig contains toml config related to containerd
type ContainerdConfig struct {
	// Snapshotter is the snapshotter used by containerd.
	Snapshotter string `toml:"snapshotter" json:"snapshotter"`
	// DefaultRuntimeName is the default runtime name to use from the runtimes table.
	DefaultRuntimeName string `toml:"default_runtime_name" json:"defaultRuntimeName"`
	// DefaultRuntime is the default runtime to use in containerd.
	// This runtime is used when no runtime handler (or the empty string) is provided.
	// DEPRECATED: use DefaultRuntimeName instead. Remove in containerd 1.4.
	DefaultRuntime Runtime `toml:"default_runtime" json:"defaultRuntime"`
	// UntrustedWorkloadRuntime is a runtime to run untrusted workloads on it.
	// DEPRECATED: use `untrusted` runtime in Runtimes instead. Remove in containerd 1.4.
	UntrustedWorkloadRuntime Runtime `toml:"untrusted_workload_runtime" json:"untrustedWorkloadRuntime"`
	// Runtimes is a map from CRI RuntimeHandler strings, which specify types of runtime
	// configurations, to the matching configurations.
	Runtimes map[string]Runtime `toml:"runtimes" json:"runtimes"`
	// NoPivot disables pivot-root (linux only), required when running a container in a RamDisk with runc
	// This only works for runtime type "io.containerd.runtime.v1.linux".
	NoPivot bool `toml:"no_pivot" json:"noPivot"`

	// DisableSnapshotAnnotations disables to pass additional annotations (image
	// related information) to snapshotters. These annotations are required by
	// stargz snapshotter (https://github.com/containerd/stargz-snapshotter).
	DisableSnapshotAnnotations bool `toml:"disable_snapshot_annotations" json:"disableSnapshotAnnotations"`

	// DiscardUnpackedLayers is a boolean flag to specify whether to allow GC to
	// remove layers from the content store after successfully unpacking these
	// layers to the snapshotter.
	DiscardUnpackedLayers bool `toml:"discard_unpacked_layers" json:"discardUnpackedLayers"`
}

// CniConfig contains toml config related to cni
type CniConfig struct {
	// NetworkPluginBinDir is the directory in which the binaries for the plugin is kept.
	NetworkPluginBinDir string `toml:"bin_dir" json:"binDir"`
	// NetworkPluginConfDir is the directory in which the admin places a CNI conf.
	NetworkPluginConfDir string `toml:"conf_dir" json:"confDir"`
	// NetworkPluginMaxConfNum is the max number of plugin config files that will
	// be loaded from the cni config directory by go-cni. Set the value to 0 to
	// load all config files (no arbitrary limit). The legacy default value is 1.
	NetworkPluginMaxConfNum int `toml:"max_conf_num" json:"maxConfNum"`
	// NetworkPluginConfTemplate is the file path of golang template used to generate
	// cni config.
	// When it is set, containerd will get cidr(s) from kubelet to replace {{.PodCIDR}},
	// {{.PodCIDRRanges}} or {{.Routes}} in the template, and write the config into
	// NetworkPluginConfDir.
	// Ideally the cni config should be placed by system admin or cni daemon like calico,
	// weaveworks etc. However, there are still users using kubenet
	// (https://kubernetes.io/docs/concepts/cluster-administration/network-plugins/#kubenet)
	// today, who don't have a cni daemonset in production. NetworkPluginConfTemplate is
	// a temporary backward-compatible solution for them.
	// TODO(random-liu): Deprecate this option when kubenet is deprecated.
	NetworkPluginConfTemplate string `toml:"conf_template" json:"confTemplate"`
}

// Mirror contains the config related to the registry mirror
type Mirror struct {
	// Endpoints are endpoints for a namespace. CRI plugin will try the endpoints
	// one by one until a working one is found. The endpoint must be a valid url
	// with host specified.
	// The scheme, host and path from the endpoint URL will be used.
	Endpoints []string `toml:"endpoint" json:"endpoint"`
}

// AuthConfig contains the config related to authentication to a specific registry
type AuthConfig struct {
	// Username is the username to login the registry.
	Username string `toml:"username" json:"username"`
	// Password is the password to login the registry.
	Password string `toml:"password" json:"password"`
	// Auth is a base64 encoded string from the concatenation of the username,
	// a colon, and the password.
	Auth string `toml:"auth" json:"auth"`
	// IdentityToken is used to authenticate the user and get
	// an access token for the registry.
	IdentityToken string `toml:"identitytoken" json:"identitytoken"`
}

// TLSConfig contains the CA/Cert/Key used for a registry
type TLSConfig struct {
	InsecureSkipVerify bool   `toml:"insecure_skip_verify" json:"insecure_skip_verify"`
	CAFile             string `toml:"ca_file" json:"caFile"`
	CertFile           string `toml:"cert_file" json:"certFile"`
	KeyFile            string `toml:"key_file" json:"keyFile"`
}

// Registry is registry settings configured
type Registry struct {
	// ConfigPath is a path to the root directory containing registry-specific
	// configurations.
	// If ConfigPath is set, the rest of the registry specific options are ignored.
	ConfigPath string `toml:"config_path" json:"configPath"`
	// Mirrors are namespace to mirror mapping for all namespaces.
	// This option will not be used when ConfigPath is provided.
	// DEPRECATED: Use ConfigPath instead. Remove in containerd 1.7.
	Mirrors map[string]Mirror `toml:"mirrors" json:"mirrors"`
	// Configs are configs for each registry.
	// The key is the domain name or IP of the registry.
	// This option will be fully deprecated for ConfigPath in the future.
	Configs map[string]RegistryConfig `toml:"configs" json:"configs"`
	// Auths are registry endpoint to auth config mapping. The registry endpoint must
	// be a valid url with host specified.
	// DEPRECATED: Use ConfigPath instead. Remove in containerd 1.6.
	Auths map[string]AuthConfig `toml:"auths" json:"auths"`
	// Headers adds additional HTTP headers that get sent to all registries
	Headers map[string][]string `toml:"headers" json:"headers"`
}

// RegistryConfig contains configuration used to communicate with the registry.
type RegistryConfig struct {
	// Auth contains information to authenticate to the registry.
	Auth *AuthConfig `toml:"auth" json:"auth"`
	// TLS is a pair of CA/Cert/Key which then are used when creating the transport
	// that communicates with the registry.
	// This field will not be used when ConfigPath is provided.
	// DEPRECATED: Use ConfigPath instead. Remove in containerd 1.7.
	TLS *TLSConfig `toml:"tls" json:"tls"`
}

// ImageDecryption contains configuration to handling decryption of encrypted container images.
type ImageDecryption struct {
	// KeyModel specifies the trust model of where keys should reside.
	//
	// Details of field usage can be found in:
	// https://github.com/containerd/cri/tree/master/docs/config.md
	//
	// Details of key models can be found in:
	// https://github.com/containerd/cri/tree/master/docs/decryption.md
	KeyModel string `toml:"key_model" json:"keyModel"`
}

// PluginConfig contains toml config related to CRI plugin,
// it is a subset of Config.
type PluginConfig struct {
	// ContainerdConfig contains config related to containerd
	ContainerdConfig `toml:"containerd" json:"containerd"`
	// CniConfig contains config related to cni
	CniConfig `toml:"cni" json:"cni"`
	// Registry contains config related to the registry
	Registry Registry `toml:"registry" json:"registry"`
	// ImageDecryption contains config related to handling decryption of encrypted container images
	ImageDecryption `toml:"image_decryption" json:"imageDecryption"`
	// DisableTCPService disables serving CRI on the TCP server.
	DisableTCPService bool `toml:"disable_tcp_service" json:"disableTCPService"`
	// StreamServerAddress is the ip address streaming server is listening on.
	StreamServerAddress string `toml:"stream_server_address" json:"streamServerAddress"`
	// StreamServerPort is the port streaming server is listening on.
	StreamServerPort string `toml:"stream_server_port" json:"streamServerPort"`
	// StreamIdleTimeout is the maximum time a streaming connection
	// can be idle before the connection is automatically closed.
	// The string is in the golang duration format, see:
	//   https://golang.org/pkg/time/#ParseDuration
	StreamIdleTimeout string `toml:"stream_idle_timeout" json:"streamIdleTimeout"`
	// EnableSelinux indicates to enable the selinux support.
	EnableSelinux bool `toml:"enable_selinux" json:"enableSelinux"`
	// SelinuxCategoryRange allows the upper bound on the category range to be set.
	// If not specified or set to 0, defaults to 1024 from the selinux package.
	SelinuxCategoryRange int `toml:"selinux_category_range" json:"selinuxCategoryRange"`
	// SandboxImage is the image used by sandbox container.
	SandboxImage string `toml:"sandbox_image" json:"sandboxImage"`
	// StatsCollectPeriod is the period (in seconds) of snapshots stats collection.
	StatsCollectPeriod int `toml:"stats_collect_period" json:"statsCollectPeriod"`
	// SystemdCgroup enables systemd cgroup support.
	// This only works for runtime type "io.containerd.runtime.v1.linux".
	// DEPRECATED: config runc runtime handler instead. Remove when shim v1 is deprecated.
	SystemdCgroup bool `toml:"systemd_cgroup" json:"systemdCgroup"`
	// EnableTLSStreaming indicates to enable the TLS streaming support.
	EnableTLSStreaming bool `toml:"enable_tls_streaming" json:"enableTLSStreaming"`
	// X509KeyPairStreaming is a x509 key pair used for TLS streaming
	X509KeyPairStreaming `toml:"x509_key_pair_streaming" json:"x509KeyPairStreaming"`
	// MaxContainerLogLineSize is the maximum log line size in bytes for a container.
	// Log line longer than the limit will be split into multiple lines. Non-positive
	// value means no limit.
	MaxContainerLogLineSize int `toml:"max_container_log_line_size" json:"maxContainerLogSize"`
	// DisableCgroup indicates to disable the cgroup support.
	// This is useful when the containerd does not have permission to access cgroup.
	DisableCgroup bool `toml:"disable_cgroup" json:"disableCgroup"`
	// DisableApparmor indicates to disable the apparmor support.
	// This is useful when the containerd does not have permission to access Apparmor.
	DisableApparmor bool `toml:"disable_apparmor" json:"disableApparmor"`
	// RestrictOOMScoreAdj indicates to limit the lower bound of OOMScoreAdj to the containerd's
	// current OOMScoreADj.
	// This is useful when the containerd does not have permission to decrease OOMScoreAdj.
	RestrictOOMScoreAdj bool `toml:"restrict_oom_score_adj" json:"restrictOOMScoreAdj"`
	// MaxConcurrentDownloads restricts the number of concurrent downloads for each image.
	MaxConcurrentDownloads int `toml:"max_concurrent_downloads" json:"maxConcurrentDownloads"`
	// DisableProcMount disables Kubernetes ProcMount support. This MUST be set to `true`
	// when using containerd with Kubernetes <=1.11.
	DisableProcMount bool `toml:"disable_proc_mount" json:"disableProcMount"`
	// UnsetSeccompProfile is the profile containerd/cri will use If the provided seccomp profile is
	// unset (`""`) for a container (default is `unconfined`)
	UnsetSeccompProfile string `toml:"unset_seccomp_profile" json:"unsetSeccompProfile"`
	// TolerateMissingHugetlbController if set to false will error out on create/update
	// container requests with huge page limits if the cgroup controller for hugepages is not present.
	// This helps with supporting Kubernetes <=1.18 out of the box. (default is `true`)
	TolerateMissingHugetlbController bool `toml:"tolerate_missing_hugetlb_controller" json:"tolerateMissingHugetlbController"`
	// DisableHugetlbController indicates to silently disable the hugetlb controller, even when it is
	// present in /sys/fs/cgroup/cgroup.controllers.
	// This helps with running rootless mode + cgroup v2 + systemd but without hugetlb delegation.
	DisableHugetlbController bool `toml:"disable_hugetlb_controller" json:"disableHugetlbController"`
	// IgnoreImageDefinedVolumes ignores volumes defined by the image. Useful for better resource
	// isolation, security and early detection of issues in the mount configuration when using
	// ReadOnlyRootFilesystem since containers won't silently mount a temporary volume.
	IgnoreImageDefinedVolumes bool `toml:"ignore_image_defined_volumes" json:"ignoreImageDefinedVolumes"`
	// NetNSMountsUnderStateDir places all mounts for network namespaces under StateDir/netns instead
	// of being placed under the hardcoded directory /var/run/netns. Changing this setting requires
	// that all containers are deleted.
	NetNSMountsUnderStateDir bool `toml:"netns_mounts_under_state_dir" json:"netnsMountsUnderStateDir"`
}

// X509KeyPairStreaming contains the x509 configuration for streaming
type X509KeyPairStreaming struct {
	// TLSCertFile is the path to a certificate file
	TLSCertFile string `toml:"tls_cert_file" json:"tlsCertFile"`
	// TLSKeyFile is the path to a private key file
	TLSKeyFile string `toml:"tls_key_file" json:"tlsKeyFile"`
}

// Config contains all configurations for cri server.
type Config struct {
	// PluginConfig is the config for CRI plugin.
	PluginConfig
	// ContainerdRootDir is the root directory path for containerd.
	ContainerdRootDir string `json:"containerdRootDir"`
	// ContainerdEndpoint is the containerd endpoint path.
	ContainerdEndpoint string `json:"containerdEndpoint"`
	// RootDir is the root directory path for managing cri plugin files
	// (metadata checkpoint etc.)
	RootDir string `json:"rootDir"`
	// StateDir is the root directory path for managing volatile pod/container data
	StateDir string `json:"stateDir"`
}

const (
	// RuntimeUntrusted is the implicit runtime defined for ContainerdConfig.UntrustedWorkloadRuntime
	RuntimeUntrusted = "untrusted"
	// RuntimeDefault is the implicit runtime defined for ContainerdConfig.DefaultRuntime
	RuntimeDefault = "default"
	// KeyModelNode is the key model where key for encrypted images reside
	// on the worker nodes
	KeyModelNode = "node"
)

// ValidatePluginConfig validates the given plugin configuration.
func ValidatePluginConfig(ctx context.Context, c *PluginConfig) error {
	if c.ContainerdConfig.Runtimes == nil {
		c.ContainerdConfig.Runtimes = make(map[string]Runtime)
	}

	// Validation for deprecated untrusted_workload_runtime.
	if c.ContainerdConfig.UntrustedWorkloadRuntime.Type != "" {
		log.G(ctx).Warning("`untrusted_workload_runtime` is deprecated, please use `untrusted` runtime in `runtimes` instead")
		if _, ok := c.ContainerdConfig.Runtimes[RuntimeUntrusted]; ok {
			return errors.Errorf("conflicting definitions: configuration includes both `untrusted_workload_runtime` and `runtimes[%q]`", RuntimeUntrusted)
		}
		c.ContainerdConfig.Runtimes[RuntimeUntrusted] = c.ContainerdConfig.UntrustedWorkloadRuntime
	}

	// Validation for deprecated default_runtime field.
	if c.ContainerdConfig.DefaultRuntime.Type != "" {
		log.G(ctx).Warning("`default_runtime` is deprecated, please use `default_runtime_name` to reference the default configuration you have defined in `runtimes`")
		c.ContainerdConfig.DefaultRuntimeName = RuntimeDefault
		c.ContainerdConfig.Runtimes[RuntimeDefault] = c.ContainerdConfig.DefaultRuntime
	}

	// Validation for default_runtime_name
	if c.ContainerdConfig.DefaultRuntimeName == "" {
		return errors.New("`default_runtime_name` is empty")
	}
	if _, ok := c.ContainerdConfig.Runtimes[c.ContainerdConfig.DefaultRuntimeName]; !ok {
		return errors.Errorf("no corresponding runtime configured in `containerd.runtimes` for `containerd` `default_runtime_name = \"%s\"", c.ContainerdConfig.DefaultRuntimeName)
	}

	// Validation for deprecated runtime options.
	if c.SystemdCgroup {
		if c.ContainerdConfig.Runtimes[c.ContainerdConfig.DefaultRuntimeName].Type != plugin.RuntimeLinuxV1 {
			return errors.Errorf("`systemd_cgroup` only works for runtime %s", plugin.RuntimeLinuxV1)
		}
		log.G(ctx).Warning("`systemd_cgroup` is deprecated, please use runtime `options` instead")
	}
	if c.NoPivot {
		if c.ContainerdConfig.Runtimes[c.ContainerdConfig.DefaultRuntimeName].Type != plugin.RuntimeLinuxV1 {
			return errors.Errorf("`no_pivot` only works for runtime %s", plugin.RuntimeLinuxV1)
		}
		// NoPivot can't be deprecated yet, because there is no alternative config option
		// for `io.containerd.runtime.v1.linux`.
	}
	for _, r := range c.ContainerdConfig.Runtimes {
		if r.Engine != "" {
			if r.Type != plugin.RuntimeLinuxV1 {
				return errors.Errorf("`runtime_engine` only works for runtime %s", plugin.RuntimeLinuxV1)
			}
			log.G(ctx).Warning("`runtime_engine` is deprecated, please use runtime `options` instead")
		}
		if r.Root != "" {
			if r.Type != plugin.RuntimeLinuxV1 {
				return errors.Errorf("`runtime_root` only works for runtime %s", plugin.RuntimeLinuxV1)
			}
			log.G(ctx).Warning("`runtime_root` is deprecated, please use runtime `options` instead")
		}
	}

	useConfigPath := c.Registry.ConfigPath != ""
	if len(c.Registry.Mirrors) > 0 {
		if useConfigPath {
			return errors.Errorf("`mirrors` cannot be set when `config_path` is provided")
		}
		log.G(ctx).Warning("`mirrors` is deprecated, please use `config_path` instead")
	}
	var hasDeprecatedTLS bool
	for _, r := range c.Registry.Configs {
		if r.TLS != nil {
			hasDeprecatedTLS = true
			break
		}
	}
	if hasDeprecatedTLS {
		if useConfigPath {
			return errors.Errorf("`configs.tls` cannot be set when `config_path` is provided")
		}
		log.G(ctx).Warning("`configs.tls` is deprecated, please use `config_path` instead")
	}

	// Validation for deprecated auths options and mapping it to configs.
	if len(c.Registry.Auths) != 0 {
		if c.Registry.Configs == nil {
			c.Registry.Configs = make(map[string]RegistryConfig)
		}
		for endpoint, auth := range c.Registry.Auths {
			auth := auth
			u, err := url.Parse(endpoint)
			if err != nil {
				return errors.Wrapf(err, "failed to parse registry url %q from `registry.auths`", endpoint)
			}
			if u.Scheme != "" {
				// Do not include the scheme in the new registry config.
				endpoint = u.Host
			}
			config := c.Registry.Configs[endpoint]
			config.Auth = &auth
			c.Registry.Configs[endpoint] = config
		}
		log.G(ctx).Warning("`auths` is deprecated, please use `configs` instead")
	}

	// Validation for stream_idle_timeout
	if c.StreamIdleTimeout != "" {
		if _, err := time.ParseDuration(c.StreamIdleTimeout); err != nil {
			return errors.Wrap(err, "invalid stream idle timeout")
		}
	}
	return nil
}
