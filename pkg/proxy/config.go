package proxy

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/angelini/netz/pkg/config"
	accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	bootstrap "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	filelog "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/file/v3"
	router "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	httpup "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
)

type BootstrapOptions struct {
	Admin     *bootstrap.Admin
	Listeners []*listener.Listener
	Clusters  []*cluster.Cluster
}

func BuildBootstrapConfig(options BootstrapOptions) *bootstrap.Bootstrap {
	return &bootstrap.Bootstrap{
		Admin: options.Admin,
		StaticResources: &bootstrap.Bootstrap_StaticResources{
			Listeners: options.Listeners,
			Clusters:  options.Clusters,
		},
	}
}

type ClusterOptions struct {
	ConnectionTimeout time.Duration
	KeepaliveInterval time.Duration
}

func buildEndpoint(clusterName string, host config.Host) *endpoint.ClusterLoadAssignment {
	return &endpoint.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints: []*endpoint.LocalityLbEndpoints{{
			LbEndpoints: []*endpoint.LbEndpoint{{
				HostIdentifier: &endpoint.LbEndpoint_Endpoint{
					Endpoint: &endpoint.Endpoint{
						Address: &core.Address{
							Address: &core.Address_SocketAddress{
								SocketAddress: &core.SocketAddress{
									Protocol: core.SocketAddress_TCP,
									Address:  host.Address,
									PortSpecifier: &core.SocketAddress_PortValue{
										PortValue: host.Port,
									},
								},
							},
						},
					},
				},
			}},
		}},
	}
}

func BuildClusterConfig(name string, host config.Host, options ClusterOptions) *cluster.Cluster {
	protocolOptions := &httpup.HttpProtocolOptions{
		CommonHttpProtocolOptions: &core.HttpProtocolOptions{
			IdleTimeout: durationpb.New(60 * time.Second),
		},
		UpstreamProtocolOptions: &httpup.HttpProtocolOptions_ExplicitHttpConfig_{
			ExplicitHttpConfig: &httpup.HttpProtocolOptions_ExplicitHttpConfig{
				ProtocolConfig: &httpup.HttpProtocolOptions_ExplicitHttpConfig_Http2ProtocolOptions{
					Http2ProtocolOptions: &core.Http2ProtocolOptions{
						AllowConnect: true,
						ConnectionKeepalive: &core.KeepaliveSettings{
							Interval: durationpb.New(options.KeepaliveInterval),
							Timeout:  durationpb.New(1 * time.Second),
						},
					},
				},
			},
		},
	}

	protocolOptionsConfig, _ := anypb.New(protocolOptions)

	return &cluster.Cluster{
		Name:                 fmt.Sprintf("%s-service", name),
		ConnectTimeout:       durationpb.New(options.ConnectionTimeout),
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_STRICT_DNS},
		LbPolicy:             cluster.Cluster_ROUND_ROBIN,
		LoadAssignment:       buildEndpoint(name, host),
		DnsLookupFamily:      cluster.Cluster_ALL,
		TypedExtensionProtocolOptions: map[string]*anypb.Any{
			"envoy.extensions.upstreams.http.v3.HttpProtocolOptions": protocolOptionsConfig,
		},
	}
}

func stringMapToStruct(input map[string]string) *structpb.Struct {
	fields := make(map[string]*structpb.Value, len(input))
	for key, value := range input {
		fields[key] = &structpb.Value{
			Kind: &structpb.Value_StringValue{
				StringValue: value,
			},
		}
	}
	return &structpb.Struct{
		Fields: fields,
	}
}

func BuildListenerConfig(name string, protocol config.Protocol, direction config.FlowDirection, port uint32, domains []string, logDir string, repeatCounter int) *listener.Listener {
	routerConfig, _ := anypb.New(&router.Router{})
	fullName := name
	if repeatCounter != -1 {
		fullName = fmt.Sprintf("%s-%d", name, repeatCounter)
	}

	accessLog := &filelog.FileAccessLog{
		Path: filepath.Join(logDir, fmt.Sprintf("%s-%s.log", fullName, direction)),
		AccessLogFormat: &filelog.FileAccessLog_LogFormat{
			LogFormat: &core.SubstitutionFormatString{
				Format: &core.SubstitutionFormatString_JsonFormat{
					JsonFormat: stringMapToStruct(map[string]string{
						"timestamp":      "%START_TIME%",
						"protocol":       "%PROTOCOL%",
						"method":         "%REQ(:METHOD)%",
						"user_agent":     "%REQ(USER-AGENT)%",
						"bytes_received": "%BYTES_RECEIVED%",
						"bytes_sent":     "%BYTES_SENT%",
						"status":         "%RESPONSE_CODE%",
						"duration":       "%DURATION%",
					}),
				},
			},
		},
	}

	accessLogConfig, _ := anypb.New(accessLog)

	manager := &hcm.HttpConnectionManager{
		CodecType:  hcm.HttpConnectionManager_AUTO,
		StatPrefix: fmt.Sprintf("%s_%s", protocol, direction),
		RouteSpecifier: &hcm.HttpConnectionManager_RouteConfig{
			RouteConfig: &route.RouteConfiguration{
				Name: fmt.Sprintf("%s-route", name),
				VirtualHosts: []*route.VirtualHost{{
					Name:    fmt.Sprintf("%s-service", name),
					Domains: domains,
					Routes: []*route.Route{{
						Match: &route.RouteMatch{
							PathSpecifier: &route.RouteMatch_Prefix{
								Prefix: "/",
							},
						},
						Action: &route.Route_Route{
							Route: &route.RouteAction{
								ClusterSpecifier: &route.RouteAction_Cluster{
									Cluster: fmt.Sprintf("%s-service", name),
								},
							},
						},
					}},
				}},
			},
		},
		HttpFilters: []*hcm.HttpFilter{{
			Name:       wellknown.Router,
			ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: routerConfig},
		}},
		AccessLog: []*accesslog.AccessLog{{
			Name: wellknown.FileAccessLog,
			ConfigType: &accesslog.AccessLog_TypedConfig{
				TypedConfig: accessLogConfig,
			},
		}},
	}

	managerConfig, _ := anypb.New(manager)

	address := "127.0.0.1"
	if direction == config.Receive {
		address = "0.0.0.0"
	}

	return &listener.Listener{
		Name: fmt.Sprintf("%s-%s-listener", fullName, protocol),
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  address,
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: port,
					},
				},
			},
		},
		FilterChains: []*listener.FilterChain{{
			Filters: []*listener.Filter{{
				Name: wellknown.HTTPConnectionManager,
				ConfigType: &listener.Filter_TypedConfig{
					TypedConfig: managerConfig,
				},
			}},
		}},
	}
}
