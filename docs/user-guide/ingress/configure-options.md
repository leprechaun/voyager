Voyager Supports all valid options for defaults section of HAProxy config
https://cbonte.github.io/haproxy-dconv/1.7/configuration.html#4.2-option%20abortonclose
from the list from here
expects a json encoded map
ie: "ingress.appscode.com/default-option": {"http-keep-alive": "true", "dontlognull": "true", "clitcpka": "false"}
This will be appended in the defaults section of HAProxy as
```
option http-keep-alive
option dontlognull
no option clitcpka

```
Ingress Example:
```yaml
apiVersion: voyager.appscode.com/v1beta1
kind: Ingress
metadata:
  name: test-ingress
  namespace: default
  annotations:
    ingress.appscode.com/default-options: '{"http-keep-alive": "true", "dontlognull": "true", "clitcpka": "false"}'
spec:
  backend:
    serviceName: test-server
    servicePort: '80'
  rules:
  - host: appscode.example.com
    http:
      paths:
      - backend:
          serviceName: test-service
          servicePort: '80'
```

This ingress will generate a HAProxy template with provided timeouts. like
```console
defaults
	log global

	option http-keep-alive
	option dontlognull
	no option clitcpka

```
