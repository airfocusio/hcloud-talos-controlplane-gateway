package internal

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"text/template"
	"time"

	"github.com/hetznercloud/hcloud-go/hcloud"

	_ "embed"
)

var (
	//go:embed haproxy.cfg.tmpl
	haproxyConfigTemplateString string
)

type ServiceOpts struct {
	HcloudToken  string
	ClusterName  string
	FirewallName string
}

type Service struct {
	logger                Logger
	opts                  ServiceOpts
	ctx                   context.Context
	haproxyConfigTemplate *template.Template
	haproxyConfigFileName string
	haproxyProcess        *exec.Cmd
	hcloudClient          *hcloud.Client
	waitGroup             sync.WaitGroup
}

func NewService(ctx context.Context, logger Logger, opts ServiceOpts) (*Service, error) {
	if opts.HcloudToken == "" {
		return nil, fmt.Errorf("hcloud token missing")
	}
	if opts.ClusterName == "" {
		return nil, fmt.Errorf("cluster name missing")
	}
	if opts.FirewallName == "" {
		return nil, fmt.Errorf("firewall name missing")
	}
	haproxyTemplate, err := template.New("haproxy").Parse(haproxyConfigTemplateString)
	if err != nil {
		return nil, err
	}
	haproxyFile, err := os.CreateTemp("", "haproxy-*.cfg")
	if err != nil {
		return nil, err
	}
	return &Service{
		logger:                logger,
		opts:                  opts,
		ctx:                   ctx,
		hcloudClient:          hcloud.NewClient(hcloud.WithToken(opts.HcloudToken)),
		haproxyConfigTemplate: haproxyTemplate,
		haproxyConfigFileName: haproxyFile.Name(),
	}, nil
}

func (s *Service) Run() error {
	s.logger.Info.Printf("Starting service\n")

	if err := s.ReconcileLoop(5*time.Minute, "haproxy", s.ReconcileHaproxy); err != nil {
		return err
	}
	if err := s.ReconcileLoop(5*time.Minute, "firewall", s.ReconcileFirewall); err != nil {
		return err
	}
	s.RunHaproxy()

	s.waitGroup.Wait()
	s.logger.Info.Printf("Stopped service\n")

	return nil
}

func (s *Service) RunHaproxy() {
	go func() {
		s.waitGroup.Add(1)
		s.logger.Debug.Printf("Starting running haproxy\n")
		defer func() {
			s.logger.Debug.Printf("Stopped running haproxy\n")
			s.waitGroup.Done()
		}()

		for {
			if s.ctx.Err() != nil {
				return
			}

			haproxy := exec.CommandContext(s.ctx, "haproxy", "-f", s.haproxyConfigFileName)
			haproxy.Stdout = os.Stdout
			haproxy.Stderr = os.Stderr
			s.logger.Info.Printf("Starting haproxy")
			s.haproxyProcess = haproxy
			if err := haproxy.Run(); err != nil {
				s.logger.Error.Printf("Running haproxy failed: %v\n", err)
			}
			s.logger.Info.Printf("Stopped haproxy")
			s.haproxyProcess = nil
		}
	}()
}

func (s *Service) ReconcileLoop(delay time.Duration, name string, fn func() error) error {
	s.logger.Debug.Printf("Reconciling %s\n", name)
	if err := fn(); err != nil {
		return err
	}

	go func() {
		s.waitGroup.Add(1)
		s.logger.Debug.Printf("Starting reconcile loop %s\n", name)
		defer func() {
			s.logger.Debug.Printf("Stopped reconcile loop %s\n", name)
			s.waitGroup.Done()
		}()

		for {
			select {
			case <-time.After(delay):
			case <-s.ctx.Done():
				return
			}

			s.logger.Debug.Printf("Reconciling %s\n", name)
			if err := fn(); err != nil {
				s.logger.Warn.Printf("Reconciling %s failed: %v\n", name, err)
			}
		}
	}()

	return nil
}

func (s *Service) ReconcileHaproxy() error {
	_, controlplanePrivatePv4s, err := s.RetrieveNodeIPv4s("controlplane")
	if err != nil {
		return fmt.Errorf("retrieving server IPv4s failed: %w", err)
	}

	var haproxyConfig bytes.Buffer
	if err := s.haproxyConfigTemplate.Execute(&haproxyConfig, struct{ IPv4s []string }{
		IPv4s: SlicesMap(controlplanePrivatePv4s, func(ipv4 net.IP) string { return ipv4.String() }),
	}); err != nil {
		return err
	}

	changed, err := WriteFileIfChanged(s.haproxyConfigFileName, haproxyConfig.Bytes(), 0644)
	if err != nil {
		return err
	}

	if changed {
		s.logger.Debug.Printf("Updated haproxy config at %s\n", s.haproxyConfigFileName)
		if s.haproxyProcess != nil {
			s.logger.Debug.Printf("Reloading haproxy\n")
			if err := s.haproxyProcess.Process.Signal(syscall.SIGHUP); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Service) ReconcileFirewall() error {
	nodePublicIPv4s, _, err := s.RetrieveNodeIPv4s("")
	if err != nil {
		return fmt.Errorf("retrieving server IPv4s failed: %w", err)
	}

	firewall, _, err := s.hcloudClient.Firewall.GetByName(s.ctx, s.opts.FirewallName)
	if err != nil {
		return fmt.Errorf("retrieving firewall failed: %w", err)
	}
	if firewall == nil {
		return fmt.Errorf("creating firewall is not supported")
	}

	expectedRules := []hcloud.FirewallRule{
		{
			Direction: hcloud.FirewallRuleDirectionIn,
			Protocol:  hcloud.FirewallRuleProtocolTCP,
			Port:      ValuePointer("6443"),
			SourceIPs: SlicesMap(nodePublicIPv4s, func(ip net.IP) net.IPNet {
				return net.IPNet{
					IP:   ip,
					Mask: net.IPv4Mask(255, 255, 255, 255),
				}
			}),
		},
		{
			Direction: hcloud.FirewallRuleDirectionIn,
			Protocol:  hcloud.FirewallRuleProtocolTCP,
			Port:      ValuePointer("50000"),
			SourceIPs: SlicesMap(nodePublicIPv4s, func(ip net.IP) net.IPNet {
				return net.IPNet{
					IP:   ip,
					Mask: net.IPv4Mask(255, 255, 255, 255),
				}
			}),
		},
	}

	if !HcloudFirewallRulesCompare(expectedRules, firewall.Rules) {
		if _, _, err := s.hcloudClient.Firewall.SetRules(s.ctx, firewall, hcloud.FirewallSetRulesOpts{
			Rules: expectedRules,
		}); err != nil {
			return fmt.Errorf("updating firewall rules failed: %w", err)
		}
		s.logger.Info.Printf("Updated firewall rules")
	}

	return nil
}

func (s *Service) RetrieveNodeIPv4s(role string) ([]net.IP, []net.IP, error) {
	nodeLabelSelector := fmt.Sprintf("hct.airfocus.io/cluster=%s", s.opts.ClusterName)
	if role != "" {
		nodeLabelSelector = fmt.Sprintf("%s,hct.airfocus.io/role=%s", nodeLabelSelector, role)
	}
	nodes, _, err := s.hcloudClient.Server.List(s.ctx, hcloud.ServerListOpts{
		ListOpts: hcloud.ListOpts{
			Page:          1,
			PerPage:       100,
			LabelSelector: nodeLabelSelector,
		},
	})
	if err != nil {
		return nil, nil, err
	}
	if len(nodes) == 0 {
		return nil, nil, fmt.Errorf("no nodes found with labels %s", nodeLabelSelector)
	}

	public := SlicesMap(nodes, func(server *hcloud.Server) net.IP {
		return server.PublicNet.IPv4.IP
	})
	private := SlicesFlatMap(nodes, func(server *hcloud.Server) []net.IP {
		return SlicesMap(server.PrivateNet, func(privateNet hcloud.ServerPrivateNet) net.IP {
			return privateNet.IP
		})
	})

	return public, private, nil
}
