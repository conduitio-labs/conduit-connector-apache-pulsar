// Code generated by paramgen. DO NOT EDIT.
// Source: github.com/ConduitIO/conduit-connector-sdk/tree/main/cmd/paramgen

package pulsar

import (
	sdk "github.com/conduitio/conduit-connector-sdk"
)

func (SourceConfig) Parameters() map[string]sdk.Parameter {
	return map[string]sdk.Parameter{
		"connectionTimeout": {
			Default:     "",
			Description: "connectionTimeout specifies the duration for which the client will attempt to establish a connection before timing out.",
			Type:        sdk.ParameterTypeDuration,
			Validations: []sdk.Validation{},
		},
		"disableLogging": {
			Default:     "",
			Description: "disableLogging is for internal use only",
			Type:        sdk.ParameterTypeBool,
			Validations: []sdk.Validation{},
		},
		"enableTransaction": {
			Default:     "",
			Description: "enableTransaction determines if the client should support transactions.",
			Type:        sdk.ParameterTypeBool,
			Validations: []sdk.Validation{},
		},
		"maxConnectionsPerBroker": {
			Default:     "",
			Description: "maxConnectionsPerBroker limits the number of connections to each broker.",
			Type:        sdk.ParameterTypeInt,
			Validations: []sdk.Validation{},
		},
		"memoryLimitBytes": {
			Default:     "",
			Description: "memoryLimitBytes sets the memory limit for the client in bytes. If the limit is exceeded, the client may start to block or fail operations.",
			Type:        sdk.ParameterTypeInt,
			Validations: []sdk.Validation{},
		},
		"operationTimeout": {
			Default:     "",
			Description: "operationTimeout is the duration after which an operation is considered to have timed out.",
			Type:        sdk.ParameterTypeDuration,
			Validations: []sdk.Validation{},
		},
		"subscriptionName": {
			Default:     "",
			Description: "subscriptionName is the name of the subscription to be used for consuming messages.",
			Type:        sdk.ParameterTypeString,
			Validations: []sdk.Validation{
				sdk.ValidationRequired{},
			},
		},
		"tlsAllowInsecureConnection": {
			Default:     "",
			Description: "tlsAllowInsecureConnection configures whether the internal Pulsar client accepts untrusted TLS certificate from broker (default: false)",
			Type:        sdk.ParameterTypeBool,
			Validations: []sdk.Validation{},
		},
		"tlsCertificateFile": {
			Default:     "",
			Description: "tlsCertificateFile sets the path to the TLS certificate file",
			Type:        sdk.ParameterTypeString,
			Validations: []sdk.Validation{},
		},
		"tlsKeyFilePath": {
			Default:     "",
			Description: "tlsKeyFilePath sets the path to the TLS key file",
			Type:        sdk.ParameterTypeString,
			Validations: []sdk.Validation{},
		},
		"tlsTrustCertsFilePath": {
			Default:     "",
			Description: "tlsTrustCertsFilePath sets the path to the trusted TLS certificate file",
			Type:        sdk.ParameterTypeString,
			Validations: []sdk.Validation{},
		},
		"tlsValidateHostname": {
			Default:     "",
			Description: "tlsValidateHostname configures whether the Pulsar client verifies the validity of the host name from broker (default: false)",
			Type:        sdk.ParameterTypeBool,
			Validations: []sdk.Validation{},
		},
		"topic": {
			Default:     "",
			Description: "topic specifies the Pulsar topic from which the source will consume messages.",
			Type:        sdk.ParameterTypeString,
			Validations: []sdk.Validation{
				sdk.ValidationRequired{},
			},
		},
		"url": {
			Default:     "",
			Description: "url of the Pulsar instance to connect to.",
			Type:        sdk.ParameterTypeString,
			Validations: []sdk.Validation{
				sdk.ValidationRequired{},
			},
		},
	}
}
