package tunnel

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"time"

	"github.com/ooni/psiphon/oopsi/github.com/Psiphon-Labs/psiphon-tunnel-core/ClientLibrary/clientlib"
)

// psiphonTunnel is a psiphon tunnel
type psiphonTunnel struct {
	// bootstrapTime is the bootstrapTime of the bootstrap
	bootstrapTime time.Duration

	// tunnel is the underlying psiphon tunnel
	tunnel *clientlib.PsiphonTunnel
}

// psiphonMakeWorkingDir creates the working directory
func psiphonMakeWorkingDir(config *Config) (string, error) {
	workdir := filepath.Join(config.TunnelDir, config.Name)
	if err := config.mkdirAll(workdir, 0700); err != nil {
		return "", err
	}
	return workdir, nil
}

// psiphonStart starts the psiphon tunnel.
func psiphonStart(ctx context.Context, config *Config) (Tunnel, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err() // simplifies unit testing this code
	default:
	}
	if config.TunnelDir == "" {
		return nil, ErrEmptyTunnelDir
	}
	configJSON, err := config.Session.FetchPsiphonConfig(ctx)
	if err != nil {
		return nil, err
	}
	workdir, err := psiphonMakeWorkingDir(config)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	tunnel, err := config.startPsiphon(ctx, configJSON, workdir)
	if err != nil {
		return nil, err
	}
	stop := time.Now()
	return &psiphonTunnel{tunnel: tunnel, bootstrapTime: stop.Sub(start)}, nil
}

// TODO(bassosimone): define the NullTunnel rather than relying on
// this magic that a nil psiphonTunnel works.

// Stop is an idempotent method that shuts down the tunnel
func (t *psiphonTunnel) Stop() {
	if t != nil {
		t.tunnel.Stop()
	}
}

// SOCKS5ProxyURL returns the SOCKS5 proxy URL.
func (t *psiphonTunnel) SOCKS5ProxyURL() (proxyURL *url.URL) {
	if t != nil {
		proxyURL = &url.URL{
			Scheme: "socks5",
			Host: net.JoinHostPort(
				"127.0.0.1", fmt.Sprintf("%d", t.tunnel.SOCKSProxyPort)),
		}
	}
	return
}

// BootstrapTime returns the bootstrap time
func (t *psiphonTunnel) BootstrapTime() (duration time.Duration) {
	if t != nil {
		duration = t.bootstrapTime
	}
	return
}