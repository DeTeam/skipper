/*
This command provides an executable version of skipper with the default
set of filters.

For the list of command line options, run:

    skipper -help

For details about the usage and extensibility of skipper, please see the
documentation of the root skipper package.

To see which built-in filters are available, see the skipper/filters
package documentation.
*/
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/zalando/skipper"
	"github.com/zalando/skipper/proxy"
)

const (
	defaultAddress              = ":9090"
	defaultEtcdPrefix           = "/skipper"
	defaultSourcePollTimeout    = int64(3000)
	defaultMetricsListener      = ":9911"
	defaultMetricsPrefix        = "skipper."
	defaultRuntimeMetrics       = true
	defaultApplicationLogPrefix = "[APP]"
	defaultApplicationLogLevel  = "INFO"
	defaultBackendFlushInterval = 20 * time.Millisecond
	defaultExperimentalUpgrade  = false

	addressUsage                   = "network address that skipper should listen on"
	etcdUrlsUsage                  = "urls of nodes in an etcd cluster, storing route definitions"
	etcdPrefixUsage                = "path prefix for skipper related data in etcd"
	kubernetesUsage                = "enables skipper to generate routes for ingress resources in kubernetes cluster"
	kubernetesInClusterUsage       = "specify if skipper is running inside kubernetes cluster"
	kubernetesURLUsage             = "kubernetes API base URL for the ingress data client; requires kubectl proxy running; omit if kubernetes-in-cluster is set to true"
	kubernetesHealthcheckUsage     = "automatic healthcheck route for internal IPs with path /kube-system/healthz; valid only with kubernetes"
	kubernetesHTTPSRedirectUsage   = "automatic HTTP->HTTPS redirect route; valid only with kubernetes"
	innkeeperUrlUsage              = "API endpoint of the Innkeeper service, storing route definitions"
	innkeeperAuthTokenUsage        = "fixed token for innkeeper authentication"
	innkeeperPreRouteFiltersUsage  = "filters to be prepended to each route loaded from Innkeeper"
	innkeeperPostRouteFiltersUsage = "filters to be appended to each route loaded from Innkeeper"
	oauthUrlUsage                  = "OAuth2 URL for Innkeeper authentication"
	oauthCredentialsDirUsage       = "directory where oauth credentials are stored: client.json and user.json"
	oauthScopeUsage                = "the whitespace separated list of oauth scopes"
	routesFileUsage                = "file containing static route definitions"
	sourcePollTimeoutUsage         = "polling timeout of the routing data sources, in milliseconds"
	insecureUsage                  = "flag indicating to ignore the verification of the TLS certificates of the backend services"
	proxyPreserveHostUsage         = "flag indicating to preserve the incoming request 'Host' header in the outgoing requests"
	idleConnsPerHostUsage          = "maximum idle connections per backend host"
	closeIdleConnsPeriodUsage      = "period of closing all idle connections in seconds or as a duration string. Not closing when less than 0"
	devModeUsage                   = "enables developer time behavior, like ubuffered routing updates"
	metricsListenerUsage           = "network address used for exposing the /metrics endpoint. An empty value disables metrics."
	metricsPrefixUsage             = "allows setting a custom path prefix for metrics export"
	enableProfileUsage             = "enable profile information on the metrics endpoint with path /pprof"
	debugGcMetricsUsage            = "enables reporting of the Go garbage collector statistics exported in debug.GCStats"
	runtimeMetricsUsage            = "enables reporting of the Go runtime statistics exported in runtime and specifically runtime.MemStats"
	serveRouteMetricsUsage         = "enables reporting total serve time metrics for each route"
	serveHostMetricsUsage          = "enables reporting total serve time metrics for each host"
	backendHostMetricsUsage        = "enables reporting total serve time metrics for each backend"
	applicationLogUsage            = "output file for the application log. When not set, /dev/stderr is used"
	applicationLogLevelUsage       = "log level for application logs, possible values: PANIC, FATAL, ERROR, WARN, INFO, DEBUG"
	applicationLogPrefixUsage      = "prefix for each log entry"
	accessLogUsage                 = "output file for the access log, When not set, /dev/stderr is used"
	accessLogDisabledUsage         = "when this flag is set, no access log is printed"
	debugEndpointUsage             = "when this address is set, skipper starts an additional listener returning the original and transformed requests"
	certPathTLSUsage               = "the path on the local filesystem to the certificate file (including any intermediates)"
	keyPathTLSUsage                = "the path on the local filesystem to the certificate's private key file"
	backendFlushIntervalUsage      = "flush interval for upgraded proxy connections"
	experimentalUpgradeUsage       = "enable experimental feature to handle upgrade protocol requests"
	versionUsage                   = "print Skipper version"
)

var (
	version string
	commit  string
)

var (
	address                   string
	etcdUrls                  string
	etcdPrefix                string
	insecure                  bool
	proxyPreserveHost         bool
	idleConnsPerHost          int
	closeIdleConnsPeriod      string
	kubernetes                bool
	kubernetesInCluster       bool
	kubernetesURL             string
	kubernetesHealthcheck     bool
	kubernetesHTTPSRedirect   bool
	innkeeperUrl              string
	sourcePollTimeout         int64
	routesFile                string
	oauthUrl                  string
	oauthScope                string
	oauthCredentialsDir       string
	innkeeperAuthToken        string
	innkeeperPreRouteFilters  string
	innkeeperPostRouteFilters string
	devMode                   bool
	metricsListener           string
	metricsPrefix             string
	enableProfile             bool
	debugGcMetrics            bool
	runtimeMetrics            bool
	serveRouteMetrics         bool
	serveHostMetrics          bool
	backendHostMetrics        bool
	applicationLog            string
	applicationLogLevel       string
	applicationLogPrefix      string
	accessLog                 string
	accessLogDisabled         bool
	debugListener             string
	certPathTLS               string
	keyPathTLS                string
	backendFlushInterval      time.Duration
	experimentalUpgrade       bool
	printVersion              bool
)

func init() {
	flag.StringVar(&address, "address", defaultAddress, addressUsage)
	flag.StringVar(&etcdUrls, "etcd-urls", "", etcdUrlsUsage)
	flag.BoolVar(&insecure, "insecure", false, insecureUsage)
	flag.BoolVar(&proxyPreserveHost, "proxy-preserve-host", false, proxyPreserveHostUsage)
	flag.IntVar(&idleConnsPerHost, "idle-conns-num", proxy.DefaultIdleConnsPerHost, idleConnsPerHostUsage)
	flag.StringVar(&closeIdleConnsPeriod, "close-idle-conns-period", strconv.Itoa(int(proxy.DefaultCloseIdleConnsPeriod/time.Second)), closeIdleConnsPeriodUsage)
	flag.StringVar(&etcdPrefix, "etcd-prefix", defaultEtcdPrefix, etcdPrefixUsage)
	flag.BoolVar(&kubernetes, "kubernetes", false, kubernetesUsage)
	flag.BoolVar(&kubernetesInCluster, "kubernetes-in-cluster", false, kubernetesInClusterUsage)
	flag.StringVar(&kubernetesURL, "kubernetes-url", "", kubernetesURLUsage)
	flag.BoolVar(&kubernetesHealthcheck, "kubernetes-healthcheck", true, kubernetesHealthcheckUsage)
	flag.BoolVar(&kubernetesHTTPSRedirect, "kubernetes-https-redirect", true, kubernetesHTTPSRedirectUsage)
	flag.StringVar(&innkeeperUrl, "innkeeper-url", "", innkeeperUrlUsage)
	flag.Int64Var(&sourcePollTimeout, "source-poll-timeout", defaultSourcePollTimeout, sourcePollTimeoutUsage)
	flag.StringVar(&routesFile, "routes-file", "", routesFileUsage)
	flag.StringVar(&oauthUrl, "oauth-url", "", oauthUrlUsage)
	flag.StringVar(&oauthScope, "oauth-scope", "", oauthScopeUsage)
	flag.StringVar(&oauthCredentialsDir, "oauth-credentials-dir", "", oauthCredentialsDirUsage)
	flag.StringVar(&innkeeperAuthToken, "innkeeper-auth-token", "", innkeeperAuthTokenUsage)
	flag.StringVar(&innkeeperPreRouteFilters, "innkeeper-pre-route-filters", "", innkeeperPreRouteFiltersUsage)
	flag.StringVar(&innkeeperPostRouteFilters, "innkeeper-post-route-filters", "", innkeeperPostRouteFiltersUsage)
	flag.BoolVar(&devMode, "dev-mode", false, devModeUsage)
	flag.StringVar(&metricsListener, "metrics-listener", defaultMetricsListener, metricsListenerUsage)
	flag.StringVar(&metricsPrefix, "metrics-prefix", defaultMetricsPrefix, metricsPrefixUsage)
	flag.BoolVar(&enableProfile, "enable-profile", false, enableProfileUsage)
	flag.BoolVar(&debugGcMetrics, "debug-gc-metrics", false, debugGcMetricsUsage)
	flag.BoolVar(&runtimeMetrics, "runtime-metrics", defaultRuntimeMetrics, runtimeMetricsUsage)
	flag.BoolVar(&serveRouteMetrics, "serve-route-metrics", false, serveRouteMetricsUsage)
	flag.BoolVar(&serveHostMetrics, "serve-host-metrics", false, serveHostMetricsUsage)
	flag.BoolVar(&backendHostMetrics, "backend-host-metrics", false, backendHostMetricsUsage)
	flag.StringVar(&applicationLog, "application-log", "", applicationLogUsage)
	flag.StringVar(&applicationLogLevel, "application-log-level", defaultApplicationLogLevel, applicationLogLevelUsage)
	flag.StringVar(&applicationLogPrefix, "application-log-prefix", defaultApplicationLogPrefix, applicationLogPrefixUsage)
	flag.StringVar(&accessLog, "access-log", "", accessLogUsage)
	flag.BoolVar(&accessLogDisabled, "access-log-disabled", false, accessLogDisabledUsage)
	flag.StringVar(&debugListener, "debug-listener", "", debugEndpointUsage)
	flag.StringVar(&certPathTLS, "tls-cert", "", certPathTLSUsage)
	flag.StringVar(&keyPathTLS, "tls-key", "", keyPathTLSUsage)
	flag.DurationVar(&backendFlushInterval, "backend-flush-interval", defaultBackendFlushInterval, backendFlushIntervalUsage)
	flag.BoolVar(&experimentalUpgrade, "experimental-upgrade", defaultExperimentalUpgrade, experimentalUpgradeUsage)
	flag.BoolVar(&printVersion, "version", false, versionUsage)
	flag.Parse()
}

func parseDurationFlag(ds string) (time.Duration, error) {
	d, perr := time.ParseDuration(ds)
	if perr == nil {
		return d, nil
	}

	if i, serr := strconv.Atoi(ds); serr == nil {
		return time.Duration(i) * time.Second, nil
	} else {
		// returning the first parse error as more informative
		return 0, perr
	}
}

func main() {
	if printVersion {
		fmt.Printf(
			"Skipper version %s (commit: %s)\n",
			version, commit,
		)

		return
	}

	if logLevel, err := log.ParseLevel(applicationLogLevel); err != nil {
		log.Fatal(err)
	} else {
		log.SetLevel(logLevel)
	}

	var eus []string
	if len(etcdUrls) > 0 {
		eus = strings.Split(etcdUrls, ",")
	}

	clsic, err := parseDurationFlag(closeIdleConnsPeriod)
	if err != nil {
		flag.PrintDefaults()
		os.Exit(2)
	}

	options := skipper.Options{
		Address:                   address,
		EtcdUrls:                  eus,
		EtcdPrefix:                etcdPrefix,
		Kubernetes:                kubernetes,
		KubernetesInCluster:       kubernetesInCluster,
		KubernetesURL:             kubernetesURL,
		KubernetesHealthcheck:     kubernetesHealthcheck,
		KubernetesHTTPSRedirect:   kubernetesHTTPSRedirect,
		InnkeeperUrl:              innkeeperUrl,
		SourcePollTimeout:         time.Duration(sourcePollTimeout) * time.Millisecond,
		RoutesFile:                routesFile,
		IdleConnectionsPerHost:    idleConnsPerHost,
		CloseIdleConnsPeriod:      time.Duration(clsic) * time.Second,
		IgnoreTrailingSlash:       false,
		OAuthUrl:                  oauthUrl,
		OAuthScope:                oauthScope,
		OAuthCredentialsDir:       oauthCredentialsDir,
		InnkeeperAuthToken:        innkeeperAuthToken,
		InnkeeperPreRouteFilters:  innkeeperPreRouteFilters,
		InnkeeperPostRouteFilters: innkeeperPostRouteFilters,
		DevMode:                   devMode,
		MetricsListener:           metricsListener,
		MetricsPrefix:             metricsPrefix,
		EnableProfile:             enableProfile,
		EnableDebugGcMetrics:      debugGcMetrics,
		EnableRuntimeMetrics:      runtimeMetrics,
		EnableServeRouteMetrics:   serveRouteMetrics,
		EnableServeHostMetrics:    serveHostMetrics,
		EnableBackendHostMetrics:  backendHostMetrics,
		ApplicationLogOutput:      applicationLog,
		ApplicationLogPrefix:      applicationLogPrefix,
		AccessLogOutput:           accessLog,
		AccessLogDisabled:         accessLogDisabled,
		DebugListener:             debugListener,
		CertPathTLS:               certPathTLS,
		KeyPathTLS:                keyPathTLS,
		BackendFlushInterval:      backendFlushInterval,
		ExperimentalUpgrade:       experimentalUpgrade,
	}

	if insecure {
		options.ProxyFlags |= proxy.Insecure
	}

	if proxyPreserveHost {
		options.ProxyFlags |= proxy.PreserveHost
	}

	log.Fatal(skipper.Run(options))
}
