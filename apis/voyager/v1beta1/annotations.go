package v1beta1

import (
	"encoding/json"
	"net"
	"strconv"
	"strings"

	stringz "github.com/appscode/go/strings"
	"github.com/appscode/voyager/apis/voyager"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

const (
	IngressKey = "ingress.kubernetes.io"
	EngressKey = "ingress.appscode.com"

	APISchema        = EngressKey + "/" + "api-schema" // APISchema = {APIGroup}/{APIVersion}
	APISchemaEngress = voyager.GroupName + "/v1beta1"
	APISchemaIngress = "extension/v1beta1"

	VoyagerPrefix = "voyager-"

	// LB stats options
	StatsOn          = EngressKey + "/" + "stats"
	StatsPort        = EngressKey + "/" + "stats-port"
	StatsSecret      = EngressKey + "/" + "stats-secret-name"
	StatsServiceName = EngressKey + "/" + "stats-service-name"
	DefaultStatsPort = 56789

	LBTypeHostPort     = "HostPort"
	LBTypeNodePort     = "NodePort"
	LBTypeLoadBalancer = "LoadBalancer" // default
	LBType             = EngressKey + "/" + "type"

	// Runs HAProxy on a specific set of a hosts.
	NodeSelector = EngressKey + "/" + "node-selector"

	// Replicas specify # of HAProxy pods run (default 1)
	Replicas = EngressKey + "/" + "replicas"

	// IP to be assigned to cloud load balancer
	LoadBalancerIP = EngressKey + "/" + "load-balancer-ip" // IP or empty

	// BackendWeight is the weight value of a Pod that was
	// addressed by the Endpoint, this weight will be added to server backend.
	// Traffic will be forwarded according to there weight.
	BackendWeight = EngressKey + "/" + "backend-weight"

	// https://github.com/appscode/voyager/issues/103
	// ServiceAnnotations is user provided annotations map that will be
	// applied to the service of that LoadBalancer.
	// ex: "ingress.appscode.com/annotations-service": {"key": "val"}
	ServiceAnnotations = EngressKey + "/" + "annotations-service"

	// PodAnnotations is user provided annotations map that will be
	// applied to the Pods (Deployment/ DaemonSet) of that LoadBalancer.
	// ex: "ingress.appscode.com/annotations-pod": {"key": "val"}
	PodAnnotations = EngressKey + "/" + "annotations-pod"

	// Preserves source IP for LoadBalancer type ingresses. The actual configuration
	// generated depends on the underlying cloud provider.
	//
	//  - gce, gke, azure: Adds annotation service.beta.kubernetes.io/external-traffic: OnlyLocal
	// to services used to expose HAProxy.
	// ref: https://kubernetes.io/docs/tutorials/services/source-ip/#source-ip-for-services-with-typeloadbalancer
	//
	// - aws: Enforces the use of the PROXY protocol over any connection accepted by any of
	// the sockets declared on the same line. Versions 1 and 2 of the PROXY protocol
	// are supported and correctly detected. The PROXY protocol dictates the layer
	// 3/4 addresses of the incoming connection to be used everywhere an address is
	// used, with the only exception of "tcp-request connection" rules which will
	// only see the real connection address. Logs will reflect the addresses
	// indicated in the protocol, unless it is violated, in which case the real
	// address will still be used.  This keyword combined with support from external
	// components can be used as an efficient and reliable alternative to the
	// X-Forwarded-For mechanism which is not always reliable and not even always
	// usable. See also "tcp-request connection expect-proxy" for a finer-grained
	// setting of which client is allowed to use the protocol.
	// ref: https://github.com/kubernetes/kubernetes/blob/release-1.5/pkg/cloudprovider/providers/aws/aws.go#L79
	KeepSourceIP = EngressKey + "/" + "keep-source-ip"

	// Enforces the use of the PROXY protocol over any connection accepted by HAProxy.
	AcceptProxy = EngressKey + "/" + "accept-proxy"

	// Annotations applied to resources offshoot from an ingress
	OriginAPISchema = EngressKey + "/" + "origin-api-schema" // APISchema = {APIGroup}/{APIVersion}
	OriginName      = EngressKey + "/" + "origin-name"

	EgressPoints = EngressKey + "/" + "egress-points"

	// https://github.com/appscode/voyager/issues/280
	// Supports all valid timeout option for defaults section of HAProxy
	// https://cbonte.github.io/haproxy-dconv/1.7/configuration.html#4.2-timeout%20check
	// expects a json encoded map
	// ie: "ingress.appscode.com/default-timeout": {"client": "5s"}
	//
	// If the annotation is not set default values used to config defaults section will be:
	//
	// timeout  connect         50s
	// timeout  client          50s
	// timeout  client-fin      50s
	// timeout  server          50s
	// timeout  tunnel          50s
	DefaultsTimeOut = EngressKey + "/" + "default-timeout"

	// https://github.com/appscode/voyager/issues/343
	// Supports all valid options for defaults section of HAProxy config
	// https://cbonte.github.io/haproxy-dconv/1.7/configuration.html#4.2-option%20abortonclose
	// from the list from here
	// expects a json encoded map
	// ie: "ingress.appscode.com/default-option": '{"http-keep-alive": "true", "dontlognull": "true", "clitcpka": "false"}'
	// This will be appended in the defaults section of HAProxy as
	//
	//   option http-keep-alive
	//   option dontlognull
	//   no option clitcpka
	//
	DefaultsOption = EngressKey + "/" + "default-option"

	// Available Options
	//   ssl:
	//    Creates a TLS/SSL socket when connecting to this server in order to cipher/decipher the traffic
	//
	//    if verify not set the following error may occurred
	//    [/etc/haproxy/haproxy.cfg:49] verify is enabled by default but no CA file specified.
	//    If you're running on a LAN where you're certain to trust the server's certificate,
	//    please set an explicit 'verify none' statement on the 'server' line, or use
	//    'ssl-server-verify none' in the global section to disable server-side verifications by default.
	//
	//   verify [none|required]:
	//    Sets HAProxy‘s behavior regarding the certificated presented by the server:
	//   none :
	//    doesn’t verify the certificate of the server
	//
	//   required (default value) :
	//    TLS handshake is aborted if the validation of the certificate presented by the server returns an error.
	//
	//   veryfyhost <hostname>:
	//    Sets a <hostname> to look for in the Subject and SubjectAlternateNames fields provided in the
	//    certificate sent by the server. If <hostname> can’t be found, then the TLS handshake is aborted.
	// ie.
	// ingress.appscode.com/backend-tls: "ssl verify none"
	//
	// If this annotation is not set HAProxy will connect to backend as http,
	// This value should not be set if the backend do not support https resolution.
	BackendTLSOptions = EngressKey + "/backend-tls"

	certificateAnnotationKeyEnabled                      = "certificate.appscode.com/enabled"
	certificateAnnotationKeyName                         = "certificate.appscode.com/name"
	certificateAnnotationKeyProvider                     = "certificate.appscode.com/provider"
	certificateAnnotationKeyEmail                        = "certificate.appscode.com/email"
	certificateAnnotationKeyProviderCredentialSecretName = "certificate.appscode.com/provider-secret"
	certificateAnnotationKeyACMEUserSecretName           = "certificate.appscode.com/user-secret"
	certificateAnnotationKeyACMEServerURL                = "certificate.appscode.com/server-url"

	// StickyIngress configures HAProxy to use sticky connection
	// to the backend servers.
	// Annotations could  be applied to either Ingress or backend Service (since 3.2+).
	// ie: ingress.appscode.com/sticky-session: "true"
	// If applied to Ingress, all the backend connections would be sticky
	// If applied to Service and Ingress do not have this annotation only
	// connection to that backend service will be sticky.
	// Deprecated
	StickySession                    = EngressKey + "/" + "sticky-session"
	IngressAffinity                  = IngressKey + "/affinity"
	IngressAffinitySessionCookieName = IngressKey + "/session-cookie-name"

	// Basic Auth: Follows ingress controller standard
	// https://github.com/kubernetes/ingress/tree/master/examples/auth/basic/haproxy
	// HAProxy Ingress read user and password from auth file stored on secrets, one
	// user and password per line.
	// Each line of the auth file should have:
	// user and insecure password separated with a pair of colons: <username>::<plain-text-passwd>; or
	// user and an encrypted password separated with colons: <username>:<encrypted-passwd>
	// Secret name, realm and type are configured with annotations in the ingress
	// Auth can only be applied to HTTP backends.
	// Only supported type is basic
	AuthType = IngressKey + "/auth-type"

	// an optional string with authentication realm
	AuthRealm = IngressKey + "/auth-realm"

	// name of the auth secret
	AuthSecret = IngressKey + "/auth-secret"
)

func (r Ingress) OffshootName() string {
	return VoyagerPrefix + r.Name
}

func (r Ingress) OffshootLabels() map[string]string {
	lbl := map[string]string{
		"origin":      "voyager",
		"origin-name": r.Name,
	}

	gv := strings.SplitN(r.APISchema(), "/", 2)
	if len(gv) == 2 {
		lbl["origin-api-group"] = gv[0]
	}
	return lbl
}

func (r Ingress) StatsLabels() map[string]string {
	lbl := r.OffshootLabels()
	lbl["feature"] = "stats"
	return lbl
}

func (r Ingress) APISchema() string {
	if v := GetString(r.Annotations, APISchema); v != "" {
		return v
	}
	return APISchemaEngress
}

func (r Ingress) Sticky() bool {
	// Specify a method to stick clients to origins across requests.
	// Like nginx HAProxy only supports the value cookie.
	if len(GetString(r.Annotations, IngressAffinity)) > 0 {
		return true
	}
	v, _ := GetBool(r.Annotations, StickySession)
	return v
}

func (r Ingress) StickySessionCookieName() string {
	// When affinity is set to cookie, the name of the cookie to use.
	cookieName := GetString(r.Annotations, IngressAffinitySessionCookieName)
	if len(cookieName) > 0 {
		return cookieName
	}
	return "SERVERID"
}

func (r Ingress) Stats() bool {
	v, _ := GetBool(r.Annotations, StatsOn)
	return v
}

func (r Ingress) StatsSecretName() string {
	return GetString(r.Annotations, StatsSecret)
}

func (r Ingress) StatsPort() int {
	if v, _ := GetInt(r.Annotations, StatsPort); v > 0 {
		return v
	}
	return DefaultStatsPort
}

func (r Ingress) StatsServiceName() string {
	if v := GetString(r.Annotations, StatsServiceName); v != "" {
		return v
	}
	return VoyagerPrefix + r.Name + "-stats"
}

func (r Ingress) LBType() string {
	if v := GetString(r.Annotations, LBType); v != "" {
		return v
	}
	return LBTypeLoadBalancer
}

func (r Ingress) Replicas() int32 {
	if v, _ := GetInt(r.Annotations, Replicas); v > 0 {
		return int32(v)
	}
	return 1
}

func (r Ingress) NodeSelector() map[string]string {
	if v, _ := GetMap(r.Annotations, NodeSelector); len(v) > 0 {
		return v
	}
	return ParseDaemonNodeSelector(GetString(r.Annotations, EngressKey+"/"+"daemon.nodeSelector"))
}

func (r Ingress) LoadBalancerIP() net.IP {
	if v := GetString(r.Annotations, LoadBalancerIP); v != "" {
		return net.ParseIP(v)
	}
	return nil
}

func (r Ingress) ServiceAnnotations(provider string) (map[string]string, bool) {
	ans, err := GetMap(r.Annotations, ServiceAnnotations)
	if err == nil {
		filteredMap := make(map[string]string)
		for k, v := range ans {
			if !strings.HasPrefix(strings.TrimSpace(k), EngressKey+"/") {
				filteredMap[k] = v
			}
		}
		if r.LBType() == LBTypeLoadBalancer && r.KeepSourceIP() {
			switch provider {
			case "aws":
				// ref: https://github.com/kubernetes/kubernetes/blob/release-1.5/pkg/cloudprovider/providers/aws/aws.go#L79
				filteredMap["service.beta.kubernetes.io/aws-load-balancer-proxy-protocol"] = "*"
			}
		}
		return filteredMap, true
	}
	return ans, false
}

func (r Ingress) PodsAnnotations() (map[string]string, bool) {
	ans, err := GetMap(r.Annotations, PodAnnotations)
	if err == nil {
		filteredMap := make(map[string]string)
		for k, v := range ans {
			if !strings.HasPrefix(strings.TrimSpace(k), EngressKey+"/") {
				filteredMap[k] = v
			}
		}
		return filteredMap, true
	}
	return ans, false
}

func (r Ingress) KeepSourceIP() bool {
	v, _ := GetBool(r.Annotations, KeepSourceIP)
	return v
}

func (r Ingress) AcceptProxy() bool {
	v, _ := GetBool(r.Annotations, AcceptProxy)
	return v
}

var timeoutDefaults = map[string]string{
	// Maximum time to wait for a connection attempt to a server to succeed.
	"connect": "50s",

	// Maximum inactivity time on the client side.
	// Applies when the client is expected to acknowledge or send data.
	"client": "50s",

	// Inactivity timeout on the client side for half-closed connections.
	// Applies when the client is expected to acknowledge or send data
	// while one direction is already shut down.
	"client-fin": "50s",

	// Maximum inactivity time on the server side.
	"server": "50s",

	// Timeout to use with WebSocket and CONNECT
	"tunnel": "50s",
}

func (r Ingress) Timeouts() map[string]string {
	ans, _ := GetMap(r.Annotations, DefaultsTimeOut)
	if ans == nil {
		ans = make(map[string]string)
	}

	// If the timeouts specified in `defaultTimeoutValues` are not set specifically set
	// we need to set default timeout values.
	// An unspecified timeout results in an infinite timeout, which
	// is not recommended. Such a usage is accepted and works but reports a warning
	// during startup because it may results in accumulation of expired sessions in
	// the system if the system's timeouts are not configured either.
	for k, v := range timeoutDefaults {
		if _, ok := ans[k]; !ok {
			ans[k] = v
		}
	}

	return ans
}

func (r Ingress) HAProxyOptions() map[string]bool {
	ans, _ := GetMap(r.Annotations, DefaultsOption)
	if ans == nil {
		ans = make(map[string]string)
	}

	ret := make(map[string]bool)
	for k := range ans {
		val, err := GetBool(ans, k)
		if err != nil {
			continue
		}
		ret[k] = val
	}

	if len(ret) == 0 {
		ret["http-server-close"] = true
		ret["dontlognull"] = true
	}

	return ret
}

func (r Ingress) CertificateSpec() (*Certificate, bool) {
	if r.Annotations == nil {
		return nil, false
	}
	if val, ok := r.Annotations[certificateAnnotationKeyEnabled]; ok && val == "true" {
		certificateName := r.Annotations[certificateAnnotationKeyName]
		// Check if a certificate already exists.
		newCertificate := &Certificate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      certificateName,
				Namespace: r.Namespace,
			},
			Spec: CertificateSpec{
				Provider: r.Annotations[certificateAnnotationKeyProvider],
				Email:    r.Annotations[certificateAnnotationKeyEmail],
				ProviderCredentialSecretName: r.Annotations[certificateAnnotationKeyProviderCredentialSecretName],
				HTTPProviderIngressReference: apiv1.ObjectReference{
					Kind:            "Ingress",
					Name:            r.Name,
					Namespace:       r.Namespace,
					ResourceVersion: r.ResourceVersion,
					UID:             r.UID,
				},
				ACMEUserSecretName: r.Annotations[certificateAnnotationKeyACMEUserSecretName],
				ACMEServerURL:      r.Annotations[certificateAnnotationKeyACMEServerURL],
			},
		}

		if v, ok := r.Annotations[APISchema]; ok {
			if v == APISchemaIngress {
				newCertificate.Spec.HTTPProviderIngressReference.APIVersion = APISchemaIngress
			} else {
				newCertificate.Spec.HTTPProviderIngressReference.APIVersion = APISchemaEngress
			}
		}

		for _, rule := range r.Spec.Rules {
			found := false
			for _, tls := range r.Spec.TLS {
				if stringz.Contains(tls.Hosts, rule.Host) {
					found = true
				}
			}
			if !found {
				newCertificate.Spec.Domains = append(newCertificate.Spec.Domains, rule.Host)
			}
		}
		return newCertificate, true
	}
	return nil, false
}

func (r Ingress) AuthEnabled() bool {
	if r.Annotations == nil {
		return false
	}

	// Check auth type is basic; other auth mode is not supported
	if val := GetString(r.Annotations, AuthType); val != "basic" {
		return false
	}

	// Check secret name is not empty
	if val := GetString(r.Annotations, AuthSecret); len(val) == 0 {
		return false
	}

	return true
}

func (r Ingress) AuthRealm() string {
	return GetString(r.Annotations, AuthRealm)
}

func (r Ingress) AuthSecretName() string {
	return GetString(r.Annotations, AuthSecret)
}

// ref: https://github.com/kubernetes/kubernetes/blob/078238a461a0872a8eacb887fbb3d0085714604c/staging/src/k8s.io/apiserver/pkg/apis/example/v1/types.go#L134
// Deprecated, for newer ones use '{"k1":"v1", "k2", "v2"}' form
// This expects the form k1=v1,k2=v2
func ParseDaemonNodeSelector(labels string) map[string]string {
	selectorMap := make(map[string]string)
	for _, label := range strings.Split(labels, ",") {
		label = strings.TrimSpace(label)
		if len(label) > 0 && strings.Contains(label, "=") {
			data := strings.SplitN(label, "=", 2)
			if len(data) >= 2 {
				if len(data[0]) > 0 && len(data[1]) > 0 {
					selectorMap[data[0]] = data[1]
				}
			}
		}
	}
	return selectorMap
}

func GetBool(m map[string]string, key string) (bool, error) {
	if m == nil {
		return false, nil
	}
	return strconv.ParseBool(m[key])
}

func GetInt(m map[string]string, key string) (int, error) {
	if m == nil {
		return 0, nil
	}
	s, ok := m[key]
	if !ok {
		return 0, nil
	}
	return strconv.Atoi(s)
}

func GetString(m map[string]string, key string) string {
	if m == nil {
		return ""
	}
	return m[key]
}

func GetList(m map[string]string, key string) ([]string, error) {
	if m == nil {
		return []string{}, nil
	}
	s, ok := m[key]
	if !ok {
		return []string{}, nil
	}
	v := make([]string, 0)
	err := json.Unmarshal([]byte(s), &v)
	return v, err
}

func GetMap(m map[string]string, key string) (map[string]string, error) {
	if m == nil {
		return map[string]string{}, nil
	}
	s, ok := m[key]
	if !ok {
		return map[string]string{}, nil
	}
	v := make(map[string]string)
	err := json.Unmarshal([]byte(s), &v)
	return v, err
}
