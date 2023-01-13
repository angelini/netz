package builder

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/angelini/netz/pkg/config"
	"github.com/angelini/netz/pkg/proxy"
	bootstrap "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

func writeEnvoyConfig(path string, config *bootstrap.Bootstrap) error {
	marshalOptions := protojson.MarshalOptions{
		UseProtoNames:   true,
		Indent:          "  ",
		EmitUnpopulated: false,
	}

	bootstrapBytes, err := marshalOptions.Marshal(config)
	if err != nil {
		return fmt.Errorf("cannot marshal bootstrap to JSON: %w", err)
	}

	err = os.WriteFile(filepath.Join(path, "envoy.config.json"), bootstrapBytes, 0664)
	if err != nil {
		return fmt.Errorf("cannot write envoy config to %s: %w", filepath.Join(path, "envoy.config.json"), err)
	}

	return nil
}

func BuildLocal(log *zap.Logger, root *config.Root, name string, distDir string) error {
	dir := filepath.Join(distDir, fmt.Sprintf("%s-local-proxy", name))
	log.Info("building local service proxy files", zap.String("name", name), zap.String("dir", dir))

	err := os.MkdirAll(dir, 0775)
	if err != nil {
		return fmt.Errorf("cannot create output dir: %s", dir)
	}

	service, found := root.Services[name]
	if !found {
		return fmt.Errorf("cannot find service config: %s", name)
	}

	clusters := []*cluster.Cluster{proxy.BuildClusterConfig(
		"local",
		config.Host{Address: "127.0.0.1", Port: service.LocalPort},
		proxy.ClusterOptions{
			ConnectionTimeout: root.Global.ConnectionTimeout,
			KeepaliveInterval: root.Global.KeepaliveInterval,
		},
	)}

	listeners := []*listener.Listener{proxy.BuildListenerConfig(
		"local",
		service.Protocol,
		config.Receive,
		root.Global.IngressPort,
		[]string{"*"},
		root.Global.LogDirectory,
		-1,
	)}

	for _, otherService := range root.Services {
		if service.ServiceMap[otherService.Name] {
			clusters = append(clusters, proxy.BuildClusterConfig(
				otherService.Name,
				otherService.Host,
				proxy.ClusterOptions{
					ConnectionTimeout: root.Global.ConnectionTimeout,
					KeepaliveInterval: root.Global.KeepaliveInterval,
				},
			))

			listeners = append(listeners, proxy.BuildListenerConfig(
				otherService.Name,
				otherService.Protocol,
				config.Transmit,
				otherService.LocalPort,
				[]string{"localhost"},
				root.Global.LogDirectory,
				0,
			))
		}
	}

	envoyConfig := proxy.BuildBootstrapConfig(proxy.BootstrapOptions{
		Clusters:  clusters,
		Listeners: listeners,
	})

	err = writeEnvoyConfig(dir, envoyConfig)
	if err != nil {
		return err
	}

	err = writeDockerfile(name, dir)
	if err != nil {
		return err
	}

	return nil
}

func BuildAllLocal(log *zap.Logger, root *config.Root, distDir string) error {
	for serviceName := range root.Services {
		err := BuildLocal(log, root, serviceName, distDir)
		if err != nil {
			return fmt.Errorf("failed to build service %s: %w", serviceName, err)
		}
	}
	return nil
}
