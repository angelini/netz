package builder

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/angelini/netz/pkg/config"
	"github.com/angelini/netz/pkg/proxy"
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"go.uber.org/zap"
)

func BuildFront(log *zap.Logger, root *config.Root, distDir string) error {
	dir := filepath.Join(distDir, "front-proxy")
	log.Info("building front proxy files", zap.String("dir", dir))

	err := os.MkdirAll(dir, 0775)
	if err != nil {
		return fmt.Errorf("cannot create output dir: %s", dir)
	}

	clusters := []*cluster.Cluster{}

	listeners := []*listener.Listener{}

	for _, service := range root.Services {
		if len(service.Ingresses) > 0 {
			clusters = append(clusters, proxy.BuildClusterConfig(
				service.Name,
				service.Host,
				proxy.ClusterOptions{
					ConnectionTimeout: root.Global.ConnectionTimeout,
					KeepaliveInterval: root.Global.KeepaliveInterval,
				},
			))
		}

		for idx, ingress := range service.Ingresses {
			listeners = append(listeners, proxy.BuildListenerConfig(
				service.Name,
				config.Https,
				config.Receive,
				root.Global.IngressPort,
				ingress.Domains,
				root.Global.LogDirectory,
				idx,
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

	err = writeDockerfile("front-proxy", dir)
	if err != nil {
		return err
	}

	return nil
}
