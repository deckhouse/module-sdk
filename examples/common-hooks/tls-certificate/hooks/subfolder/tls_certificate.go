package hookinfolder

import (
	"fmt"

	tlscertificate "github.com/deckhouse/module-sdk/common-hooks/tls-certificate"
)

const (
	moduleName      string = "exampleModule"
	moduleNamespace string = "d8-example-module"

	exampleWebhookCertCN string = "example-webhook"
)

var _ = tlscertificate.RegisterInternalTLSHookEM(tlscertificate.GenSelfSignedTLSHookConf{
	CN:            exampleWebhookCertCN,
	TLSSecretName: fmt.Sprintf("%s-webhook-cert", exampleWebhookCertCN),
	Namespace:     moduleNamespace,
	SANs: tlscertificate.DefaultSANs([]string{
		// example-webhook
		exampleWebhookCertCN,
		// example-webhook.d8-example-module
		fmt.Sprintf("%s.%s", exampleWebhookCertCN, moduleNamespace),
		// example-webhook.d8-example-module.svc
		fmt.Sprintf("%s.%s.svc", exampleWebhookCertCN, moduleNamespace),
		// %CLUSTER_DOMAIN%:// is a special value to generate SAN like 'example-webhook.d8-example-module.svc.cluster.local'
		fmt.Sprintf("%%CLUSTER_DOMAIN%%://%s.%s.svc", exampleWebhookCertCN, moduleNamespace),
	}),

	FullValuesPathPrefix: fmt.Sprintf("%s.internal.webhookCert", moduleName),
})
