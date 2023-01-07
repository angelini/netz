package config

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"golang.org/x/exp/slices"
)

type HclGlobalConfig struct {
	LogDirectory      string `hcl:"log_directory"`
	ExternalPort      int    `hcl:"external_port,optional"`
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

type HclRoot struct {
	Global   *HclGlobalConfig    `hcl:"global,block"`
	Services []*HclServiceConfig `hcl:"service,block"`
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
