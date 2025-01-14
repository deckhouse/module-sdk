package hookinfolder

import (
	"fmt"

	tlscertificate "github.com/deckhouse/module-sdk/common-hooks/tls-certificate"
)

const (
	MODULE_NAME string = "exampleModule"
	MODULE_NAMESPACE string = "d8-example-module"

	EXAMPLE_WEBHOOK_CERT_CN string = "example-webhook"
)

var _ = tlscertificate.RegisterInternalTLSHookEM(tlscertificate.GenSelfSignedTLSHookConf{
	CN: EXAMPLE_WEBHOOK_CERT_CN,
	TLSSecretName: fmt.Sprintf("%s-webhook-cert", EXAMPLE_WEBHOOK_CERT_CN),
	Namespace: MODULE_NAMESPACE,
	SANs: tlscertificate.DefaultSANs([]string{
		// example-webhook
		EXAMPLE_WEBHOOK_CERT_CN,
		// example-webhook.d8-example-module
		fmt.Sprintf("%s.%s", EXAMPLE_WEBHOOK_CERT_CN, MODULE_NAMESPACE),
		// example-webhook.d8-example-module.svc
		fmt.Sprintf("%s.%s.svc", EXAMPLE_WEBHOOK_CERT_CN, MODULE_NAMESPACE),
		// %CLUSTER_DOMAIN%:// is a special value to generate SAN like 'example-webhook.d8-example-module.svc.cluster.local'
		fmt.Sprintf("%%CLUSTER_DOMAIN%%://%s.%s.svc", EXAMPLE_WEBHOOK_CERT_CN, MODULE_NAMESPACE),
	}),

	FullValuesPathPrefix: fmt.Sprintf("%s.internal.webhookCert", MODULE_NAME),
})
