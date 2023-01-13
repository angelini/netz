package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/slices"
)

type Protocol int

func (p Protocol) DefaultPort() uint32 {
	switch p {
	case Https:
		return 443
	case Grpc:
		return 443
	case Http:
		return 80
	default:
		return 80
	}
}

func (p Protocol) String() string {
	switch p {
	case Https:
		return "https"
	case Grpc:
		return "grpc"
	case Http:
		return "http"
	default:
		return "unknown"
	}
}

const (
	Http Protocol = iota
	Https
	Grpc
)

func parseProtocol(str string) (Protocol, error) {
	switch str {
	case "http":
		return Http, nil
	case "https":
		return Https, nil
	case "grpc":
		return Grpc, nil
	}

	return -1, fmt.Errorf("failed to parse protocol %v", str)
}

type FlowDirection int

const (
	Transmit FlowDirection = iota
	Receive
)

func (f FlowDirection) String() string {
	switch f {
	case Transmit:
		return "tx"
	case Receive:
		return "rx"
	default:
		return "unknown"
	}
}

type Host struct {
	Address string
	Port    uint32
}

func parseHost(str string, protocol Protocol) (Host, error) {
	if strings.Contains(str, ":") {
		split := strings.Split(str, ":")

		port, err := strconv.Atoi(split[1])
		if err != nil {
			return Host{}, fmt.Errorf("failed to parse port %s: %w", split[1], err)
		}

		return Host{
			Address: split[0],
			Port:    uint32(port),
		}, nil
	}

	return Host{
		Address: str,
		Port:    protocol.DefaultPort(),
	}, nil
}

type Global struct {
	LogDirectory      string
	IngressPort       uint32
	ConnectionTimeout time.Duration
	KeepaliveInterval time.Duration
}

func globalFromHcl(config *HclGlobalConfig) (*Global, error) {
	var err error

	ingressPort := uint32(8080)
	if config.IngressPort != 0 {
		ingressPort = uint32(config.IngressPort)
	}

	connectionTimeout := 1 * time.Second
	if config.ConnectionTimeout != "" {
		connectionTimeout, err = time.ParseDuration(config.ConnectionTimeout)
		if err != nil {
			return nil, fmt.Errorf("failed to parse connection timeout: %w", err)
		}
	}

	keepaliveInterval := 10 * time.Second
	if config.KeepaliveInterval != "" {
		keepaliveInterval, err = time.ParseDuration(config.KeepaliveInterval)
		if err != nil {
			return nil, fmt.Errorf("failed to parsed keepalive interval: %w", err)
		}
	}

	return &Global{
		LogDirectory:      config.LogDirectory,
		IngressPort:       ingressPort,
		ConnectionTimeout: connectionTimeout,
		KeepaliveInterval: keepaliveInterval,
	}, nil
}

type Ingress struct {
	Domains []string
}

type Service struct {
	Name       string
	Protocol   Protocol
	Host       Host
	LocalPort  uint32
	Ingresses  []Ingress
	ServiceMap map[string]bool
}

func serviceFromHcl(config *HclServiceConfig, root *HclRoot) (*Service, error) {
	protocol, err := parseProtocol(config.Protocol)
	if err != nil {
		return nil, err
	}

	host, err := parseHost(config.Address, protocol)
	if err != nil {
		return nil, err
	}

	serviceMap := make(map[string]bool, len(root.Services))
	for _, otherService := range root.Services {
		canConnect := false

		if otherService.Name != config.Name {
			if otherService.AllowAllServices {
				canConnect = true
			} else {
				canConnect = slices.Index(otherService.ConnectingServices, config.Name) != -1
			}
		}

		serviceMap[otherService.Name] = canConnect
	}

	var ingresses []Ingress

	for _, ingress := range root.Ingresses {
		if ingress.Name == config.Name {
			ingresses = append(ingresses, Ingress{
				Domains: ingress.Domains,
			})
		}
	}

	return &Service{
		Name:       config.Name,
		Protocol:   protocol,
		Host:       host,
		LocalPort:  uint32(config.LocalPort),
		Ingresses:  ingresses,
		ServiceMap: serviceMap,
	}, nil
}

type Root struct {
	Global   *Global
	Services map[string]*Service
}

func rootFromHcl(root *HclRoot) (*Root, error) {
	services := make(map[string]*Service, len(root.Services))

	for _, serviceConfig := range root.Services {
		service, err := serviceFromHcl(serviceConfig, root)
		if err != nil {
			return nil, fmt.Errorf("cannot build service from HCL config %s/%s: %w", serviceConfig.Protocol, serviceConfig.Name, err)
		}

		services[service.Name] = service
	}

	global, err := globalFromHcl(root.Global)
	if err != nil {
		return nil, fmt.Errorf("cannot build global from HCL config: %w", err)
	}

	return &Root{
		Global:   global,
		Services: services,
	}, nil
}

func Parse(path string) (*Root, error) {
	hclRoot, err := parseHcl(path)
	if err != nil {
		return nil, err
	}

	return rootFromHcl(hclRoot)
}
