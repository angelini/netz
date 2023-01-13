package config

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"golang.org/x/exp/slices"
)

type HclGlobalConfig struct {
	LogDirectory      string `hcl:"log_directory"`
	IngressPort       int    `hcl:"ingress_port,optional"`
	ConnectionTimeout string `hcl:"connection_timeout,optional"`
	KeepaliveInterval string `hcl:"keepalive_interval,optional"`
}

type HclServiceConfig struct {
	Protocol           string   `hcl:"protocol,label"`
	Name               string   `hcl:"name,label"`
	Address            string   `hcl:"address"`
	LocalPort          int      `hcl:"local_port"`
	AllowAllServices   bool     `hcl:"allow_all_services,optional"`
	ConnectingServices []string `hcl:"connecting_services,optional"`
}

func (s *HclServiceConfig) validate(serviceNames []string) error {
	if s.AllowAllServices && len(s.ConnectingServices) > 0 {
		return fmt.Errorf("cannot mix allow_all_services and connecting_services")
	}

	for _, connecting := range s.ConnectingServices {
		if connecting == s.Name {
			return fmt.Errorf("cannot allow connection from self")
		}

		idx := slices.Index(serviceNames, connecting)
		if idx == -1 {
			return fmt.Errorf("cannot receive connections from service %s, not found", connecting)
		}
	}

	return nil
}

type HclIngressConfig struct {
	Name    string   `hcl:"name,label"`
	Domains []string `hcl:"domains"`
}

func (s *HclIngressConfig) validate(serviceNames []string) error {
	if len(s.Domains) == 0 {
		return fmt.Errorf("domains cannot be an empty list")
	}

	idx := slices.Index(serviceNames, s.Name)
	if idx == -1 {
		return fmt.Errorf("cannot direct ingress to service %s, not found", s.Name)
	}

	return nil
}

type HclRoot struct {
	Global    *HclGlobalConfig    `hcl:"global,block"`
	Services  []*HclServiceConfig `hcl:"service,block"`
	Ingresses []*HclIngressConfig `hcl:"ingress,block"`
}

func (r *HclRoot) validate() error {
	var serviceNames []string

	for _, service := range r.Services {
		serviceNames = append(serviceNames, service.Name)
	}

	for _, service := range r.Services {
		err := service.validate(serviceNames)
		if err != nil {
			return fmt.Errorf("invalid service config %s/%s: %w", service.Protocol, service.Name, err)
		}
	}

	for _, ingress := range r.Ingresses {
		err := ingress.validate(serviceNames)
		if err != nil {
			return fmt.Errorf("invalid ingress config %s: %w", ingress.Name, err)
		}
	}

	return nil
}

func parseHcl(path string) (*HclRoot, error) {
	var config HclRoot
	err := hclsimple.DecodeFile(path, nil, &config)
	if err != nil {
		return nil, err
	}

	err = config.validate()
	if err != nil {
		return nil, err
	}

	return &config, nil
}
