package config

import (
	"github.com/rocket-pool/smartnode/shared/types/config"
	"github.com/rocket-pool/smartnode/shared/utils/sys"
)

const (
	lighthouseTagPortableTest string = "sigp/lighthouse:v4.4.1"
	lighthouseTagPortableProd string = "sigp/lighthouse:v4.4.1"
	lighthouseTagModernTest   string = "sigp/lighthouse:v4.4.1-modern"
	lighthouseTagModernProd   string = "sigp/lighthouse:v4.4.1-modern"
	defaultLhMaxPeers         uint16 = 80
)

// Configuration for Lighthouse
type LighthouseConfig struct {
	Title string `yaml:"-"`

	// The max number of P2P peers to connect to
	MaxPeers config.Parameter `yaml:"maxPeers,omitempty"`

	// Common parameters that Lighthouse doesn't support and should be hidden
	UnsupportedCommonParams []string `yaml:"-"`

	// The Docker Hub tag for Lighthouse
	ContainerTag config.Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags for the BN
	AdditionalBnFlags config.Parameter `yaml:"additionalBnFlags,omitempty"`

	// Custom command line flags for the VC
	AdditionalVcFlags config.Parameter `yaml:"additionalVcFlags,omitempty"`
}

// Generates a new Lighthouse configuration
func NewLighthouseConfig(cfg *RocketPoolConfig) *LighthouseConfig {
	return &LighthouseConfig{
		Title: "Lighthouse Settings",

		MaxPeers: config.Parameter{
			ID:                   "maxPeers",
			Name:                 "Max Peers",
			Description:          "The maximum number of peers your client should try to maintain. You can try lowering this if you have a low-resource system or a constrained network.",
			Type:                 config.ParameterType_Uint16,
			Default:              map[config.Network]interface{}{config.Network_All: defaultLhMaxPeers},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth2},
			EnvironmentVariables: []string{"BN_MAX_PEERS"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		ContainerTag: config.Parameter{
			ID:          "containerTag",
			Name:        "Container Tag",
			Description: "The tag name of the Lighthouse container you want to use from Docker Hub.",
			Type:        config.ParameterType_String,
			Default: map[config.Network]interface{}{
				config.Network_Mainnet: getLighthouseTagProd(),
				config.Network_Prater:  getLighthouseTagTest(),
				config.Network_Devnet:  getLighthouseTagTest(),
			},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth2, config.ContainerID_Validator},
			EnvironmentVariables: []string{"BN_CONTAINER_TAG", "VC_CONTAINER_TAG"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},

		AdditionalBnFlags: config.Parameter{
			ID:                   "additionalBnFlags",
			Name:                 "Additional Beacon Client Flags",
			Description:          "Additional custom command line flags you want to pass Lighthouse's Beacon Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth2},
			EnvironmentVariables: []string{"BN_ADDITIONAL_FLAGS"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},

		AdditionalVcFlags: config.Parameter{
			ID:                   "additionalVcFlags",
			Name:                 "Additional Validator Client Flags",
			Description:          "Additional custom command line flags you want to pass Lighthouse's Validator Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Validator},
			EnvironmentVariables: []string{"VC_ADDITIONAL_FLAGS"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Get the parameters for this config
func (cfg *LighthouseConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.MaxPeers,
		&cfg.ContainerTag,
		&cfg.AdditionalBnFlags,
		&cfg.AdditionalVcFlags,
	}
}

// Get the common params that this client doesn't support
func (cfg *LighthouseConfig) GetUnsupportedCommonParams() []string {
	return cfg.UnsupportedCommonParams
}

// Get the Docker container name of the validator client
func (cfg *LighthouseConfig) GetValidatorImage() string {
	return cfg.ContainerTag.Value.(string)
}

// Get the name of the client
func (cfg *LighthouseConfig) GetName() string {
	return "Lighthouse"
}

// The the title for the config
func (cfg *LighthouseConfig) GetConfigTitle() string {
	return cfg.Title
}

// Get the appropriate LH default tag for production
func getLighthouseTagProd() string {
	missingFeatures := sys.GetMissingModernCpuFeatures()
	if len(missingFeatures) > 0 {
		return lighthouseTagPortableProd
	}
	return lighthouseTagModernProd
}

// Get the appropriate LH default tag for testnets
func getLighthouseTagTest() string {
	missingFeatures := sys.GetMissingModernCpuFeatures()
	if len(missingFeatures) > 0 {
		return lighthouseTagPortableTest
	}
	return lighthouseTagModernTest
}
