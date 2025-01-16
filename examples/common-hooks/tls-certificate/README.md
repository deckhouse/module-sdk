# Module Hooks TLS Certificate example
In this example you can build your basic hook binary with common TLS Ceritificate hook.

Note that this hook only adds the (raw) certificate to `FullValuesPathPrefix`, it DOES NOT create a secret with certificate.
You should create the required secret yourself, for example, using the helm template like this:

```yaml
---
apiVersion: v1
kind: Secret
metadata:
  name: example-webhook-cert
  namespace: d8-{{ .Chart.Name }}
  {{- include "helm_lib_module_labels" (list . (dict "app" "seaweedfs-operator")) | nindent 2 }}
type: kubernetes.io/tls
data:
  ca.crt: {{ .Values.exampleModule.internal.webhookCert.ca | b64enc | quote }}
  tls.crt: {{ .Values.exampleModule.internal.webhookCert.crt | b64enc | quote }}
  tls.key: {{ .Values.exampleModule.internal.webhookCert.key | b64enc | quote }}
```

### Run

To get list of your registered hooks
```bash
go run . hook list
```

To get configs of your registered hooks
```bash
go run . hook config
```

To dump configs of your registered hooks in file
```bash
go run . hook dump
```

To run registered hook with index '0' (you can see index of your hook in output of "hook list" command)
```bash
go run . hook run 0
```

By default, all logs in hooks are suppressed and he waiting for files in default folders.
To make them available, you must add env variable LOG_LEVEL and CREATE_FILES.
```bash
CREATE_FILES=true LOG_LEVEL=INFO go run . hook run 0
```

### Build
```bash
go build -o tls-certificate .
```