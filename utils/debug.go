package utils

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/LoveRetro/nextui-pak-store/models"
	"github.com/LoveRetro/nextui-pak-store/version"
)

const fetchTimeout = 30 * time.Second

// LogNetworkDiagnostics logs details about the device environment that affect
// whether the storefront can be fetched.
func LogNetworkDiagnostics() {
	logger := gabagool.GetLogger()

	build := version.Get()
	logger.Debug("Build info",
		"version", build.Version,
		"commit", build.GitCommit,
		"build_date", build.BuildDate)

	logger.Debug("Environment",
		"platform_env", os.Getenv("PLATFORM"),
		"platform", GetPlatform(),
		"logs_dir", GetLogsDir(),
		"time", time.Now().Format(time.RFC3339),
		"http_proxy", os.Getenv("HTTP_PROXY")+os.Getenv("http_proxy"),
		"https_proxy", os.Getenv("HTTPS_PROXY")+os.Getenv("https_proxy"),
		"no_proxy", os.Getenv("NO_PROXY")+os.Getenv("no_proxy"),
		"ssl_cert_file", os.Getenv("SSL_CERT_FILE"),
		"ssl_cert_dir", os.Getenv("SSL_CERT_DIR"))

	if resolv, err := os.ReadFile("/etc/resolv.conf"); err != nil {
		logger.Debug("Unable to read /etc/resolv.conf", "error", err)
	} else {
		logger.Debug("/etc/resolv.conf", "contents", string(resolv))
	}

	if target, err := os.Readlink("/etc/resolv.conf"); err == nil {
		logger.Debug("/etc/resolv.conf is a symlink", "target", target)
	}

	if hosts, err := os.ReadFile("/etc/hosts"); err != nil {
		logger.Debug("Unable to read /etc/hosts", "error", err)
	} else {
		logger.Debug("/etc/hosts", "contents", string(hosts))
	}

	if ifaces, err := net.Interfaces(); err != nil {
		logger.Debug("Unable to list network interfaces", "error", err)
	} else {
		for _, iface := range ifaces {
			var addrs []string
			if ifaceAddrs, err := iface.Addrs(); err == nil {
				for _, a := range ifaceAddrs {
					addrs = append(addrs, a.String())
				}
			}
			logger.Debug("Network interface",
				"name", iface.Name,
				"flags", iface.Flags.String(),
				"addrs", strings.Join(addrs, ","))
		}
	}

	host := "raw.githubusercontent.com"
	if u, err := url.Parse(storefrontURL()); err == nil && u.Hostname() != "" {
		host = u.Hostname()
	}
	logger.Debug("Storefront URL", "url", storefrontURL())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	start := time.Now()
	addrs, err := net.DefaultResolver.LookupHost(ctx, host)
	logger.Debug("DNS pre-flight lookup",
		"host", host,
		"addrs", strings.Join(addrs, ","),
		"elapsed", time.Since(start).String(),
		"error", errString(err))
}

func storefrontURL() string {
	if override := os.Getenv("STOREFRONT_OVERRIDE"); override != "" {
		return override
	}
	return models.StorefrontJsonURL
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// newTracedRequest builds a GET request that logs each phase of the connection.
func newTracedRequest(rawURL string) (*http.Request, error) {
	logger := gabagool.GetLogger()
	start := time.Now()
	since := func() string { return time.Since(start).String() }

	trace := &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			logger.Debug("HTTP get conn", "host_port", hostPort, "elapsed", since())
		},
		DNSStart: func(info httptrace.DNSStartInfo) {
			logger.Debug("HTTP DNS start", "host", info.Host, "elapsed", since())
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			var addrs []string
			for _, a := range info.Addrs {
				addrs = append(addrs, a.String())
			}
			logger.Debug("HTTP DNS done",
				"addrs", strings.Join(addrs, ","),
				"coalesced", info.Coalesced,
				"error", errString(info.Err),
				"elapsed", since())
		},
		ConnectStart: func(network, addr string) {
			logger.Debug("HTTP connect start", "network", network, "addr", addr, "elapsed", since())
		},
		ConnectDone: func(network, addr string, err error) {
			logger.Debug("HTTP connect done",
				"network", network,
				"addr", addr,
				"error", errString(err),
				"elapsed", since())
		},
		TLSHandshakeStart: func() {
			logger.Debug("HTTP TLS handshake start", "elapsed", since())
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			args := []any{
				"version", tls.VersionName(state.Version),
				"cipher_suite", tls.CipherSuiteName(state.CipherSuite),
				"server_name", state.ServerName,
				"error", errString(err),
				"elapsed", since(),
			}
			for i, cert := range state.PeerCertificates {
				args = append(args, fmt.Sprintf("peer_cert_%d", i), fmt.Sprintf(
					"subject=%q issuer=%q not_before=%s not_after=%s",
					cert.Subject.String(),
					cert.Issuer.String(),
					cert.NotBefore.Format(time.RFC3339),
					cert.NotAfter.Format(time.RFC3339)))
			}
			logger.Debug("HTTP TLS handshake done", args...)
		},
		GotConn: func(info httptrace.GotConnInfo) {
			logger.Debug("HTTP got conn",
				"local", info.Conn.LocalAddr().String(),
				"remote", info.Conn.RemoteAddr().String(),
				"reused", info.Reused,
				"elapsed", since())
		},
		GotFirstResponseByte: func() {
			logger.Debug("HTTP got first response byte", "elapsed", since())
		},
	}

	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}

	return req.WithContext(httptrace.WithClientTrace(req.Context(), trace)), nil
}

// logRequestError logs an HTTP error along with the type of each wrapped error.
func logRequestError(rawURL string, err error) {
	logger := gabagool.GetLogger()

	var chain []string
	for e := err; e != nil; e = errors.Unwrap(e) {
		chain = append(chain, fmt.Sprintf("%T: %v", e, e))
	}

	args := []any{
		"url", rawURL,
		"error", err.Error(),
		"chain", strings.Join(chain, " -> "),
	}

	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		args = append(args,
			"dns_name", dnsErr.Name,
			"dns_server", dnsErr.Server,
			"dns_not_found", dnsErr.IsNotFound,
			"dns_timeout", dnsErr.IsTimeout,
			"dns_temporary", dnsErr.IsTemporary)
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		args = append(args, "op", opErr.Op, "op_net", opErr.Net)
		if opErr.Addr != nil {
			args = append(args, "op_addr", opErr.Addr.String())
		}
	}

	var unknownAuth x509.UnknownAuthorityError
	if errors.As(err, &unknownAuth) {
		args = append(args, "x509", "unknown authority")
	}

	var certInvalid x509.CertificateInvalidError
	if errors.As(err, &certInvalid) {
		args = append(args, "x509", "certificate invalid", "x509_detail", certInvalid.Detail)
	}

	var hostnameErr x509.HostnameError
	if errors.As(err, &hostnameErr) {
		args = append(args, "x509", "hostname mismatch", "x509_host", hostnameErr.Host)
	}

	if errors.Is(err, context.DeadlineExceeded) {
		args = append(args, "timeout", true)
	}

	logger.Error("HTTP request failed", args...)
}
