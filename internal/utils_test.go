package internal_test

import (
	"fmt"
	"net"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/airfocusio/hcloud-talos-controlplane-gateway/internal"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/stretchr/testify/assert"
)

func TestSlicesCompare(t *testing.T) {
	assert.Equal(t, true, internal.SlicesCompare([]string{}, []string{}))
	assert.Equal(t, true, internal.SlicesCompare([]string{"a"}, []string{"a"}))
	assert.Equal(t, true, internal.SlicesCompare([]string{"a", "b"}, []string{"a", "b"}))
	assert.Equal(t, false, internal.SlicesCompare([]string{}, []string{"a"}))
	assert.Equal(t, false, internal.SlicesCompare([]string{"a"}, []string{}))
	assert.Equal(t, false, internal.SlicesCompare([]string{"a"}, []string{"b"}))
	assert.Equal(t, false, internal.SlicesCompare([]string{"a", "b"}, []string{"a", "c"}))
	assert.Equal(t, false, internal.SlicesCompare([]string{"a", "b"}, []string{"b", "a"}))
}

func TestWriteFileIfChanged(t *testing.T) {
	t.Run("no exists", func(t *testing.T) {
		name := path.Join(os.TempDir(), strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(name)

		if changed, err := internal.WriteFileIfChanged(name, []byte(""), 0644); assert.NoError(t, err) {
			assert.Equal(t, true, changed)
		}
	})

	t.Run("exists unchanged", func(t *testing.T) {
		name := path.Join(os.TempDir(), strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(name)
		os.WriteFile(name, []byte("test"), 0644)

		if changed, err := internal.WriteFileIfChanged(name, []byte("test"), 0644); assert.NoError(t, err) {
			assert.Equal(t, false, changed)
		}
	})

	t.Run("exists changed", func(t *testing.T) {
		name := path.Join(os.TempDir(), strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(name)
		os.WriteFile(name, []byte("test"), 0644)

		if changed, err := internal.WriteFileIfChanged(name, []byte("test2"), 0644); assert.NoError(t, err) {
			assert.Equal(t, true, changed)
		}
	})
}

func TestHcloudFirewallRulesCompare(t *testing.T) {
	assert.Equal(t, true, internal.HcloudFirewallRulesCompare([]hcloud.FirewallRule{}, []hcloud.FirewallRule{}))
	assert.Equal(t, true, internal.HcloudFirewallRulesCompare([]hcloud.FirewallRule{{}}, []hcloud.FirewallRule{{}}))
	assert.Equal(t, false, internal.HcloudFirewallRulesCompare([]hcloud.FirewallRule{{}}, []hcloud.FirewallRule{}))
	assert.Equal(t, false, internal.HcloudFirewallRulesCompare([]hcloud.FirewallRule{}, []hcloud.FirewallRule{{}}))

	t.Run("Direction", func(t *testing.T) {
		assert.Equal(t, true, internal.HcloudFirewallRulesCompare([]hcloud.FirewallRule{{
			Direction: hcloud.FirewallRuleDirectionIn,
		}}, []hcloud.FirewallRule{{
			Direction: hcloud.FirewallRuleDirectionIn,
		}}))
		assert.Equal(t, false, internal.HcloudFirewallRulesCompare([]hcloud.FirewallRule{{
			Direction: hcloud.FirewallRuleDirectionIn,
		}}, []hcloud.FirewallRule{{
			Direction: hcloud.FirewallRuleDirectionOut,
		}}))
	})

	t.Run("SourceIPs", func(t *testing.T) {
		assert.Equal(t, true, internal.HcloudFirewallRulesCompare([]hcloud.FirewallRule{{
			SourceIPs: mustParseCIDRs("10.0.0.0/24"),
		}}, []hcloud.FirewallRule{{
			SourceIPs: mustParseCIDRs("10.0.0.0/24"),
		}}))
		assert.Equal(t, true, internal.HcloudFirewallRulesCompare([]hcloud.FirewallRule{{
			SourceIPs: mustParseCIDRs("10.0.0.0/24", "10.0.1.0/24"),
		}}, []hcloud.FirewallRule{{
			SourceIPs: mustParseCIDRs("10.0.1.0/24", "10.0.0.0/24"),
		}}))
		assert.Equal(t, true, internal.HcloudFirewallRulesCompare([]hcloud.FirewallRule{{
			SourceIPs: mustParseCIDRs("10.0.0.0/24", "10.0.1.0/24"),
		}}, []hcloud.FirewallRule{{
			SourceIPs: mustParseCIDRs("10.0.1.0/24", "10.0.0.0/24"),
		}}))
		assert.Equal(t, false, internal.HcloudFirewallRulesCompare([]hcloud.FirewallRule{{
			SourceIPs: mustParseCIDRs("10.0.0.0/24"),
		}}, []hcloud.FirewallRule{{
			SourceIPs: mustParseCIDRs("10.0.1.0/24"),
		}}))
		assert.Equal(t, false, internal.HcloudFirewallRulesCompare([]hcloud.FirewallRule{{
			SourceIPs: mustParseCIDRs("10.0.0.0/24"),
		}}, []hcloud.FirewallRule{{
			SourceIPs: mustParseCIDRs("10.0.0.0/24", "10.0.1.0/24"),
		}}))
	})

	t.Run("DestinationIPs", func(t *testing.T) {
		assert.Equal(t, true, internal.HcloudFirewallRulesCompare([]hcloud.FirewallRule{{
			DestinationIPs: mustParseCIDRs("10.0.0.0/24"),
		}}, []hcloud.FirewallRule{{
			DestinationIPs: mustParseCIDRs("10.0.0.0/24"),
		}}))
		assert.Equal(t, true, internal.HcloudFirewallRulesCompare([]hcloud.FirewallRule{{
			DestinationIPs: mustParseCIDRs("10.0.0.0/24", "10.0.1.0/24"),
		}}, []hcloud.FirewallRule{{
			DestinationIPs: mustParseCIDRs("10.0.1.0/24", "10.0.0.0/24"),
		}}))
		assert.Equal(t, true, internal.HcloudFirewallRulesCompare([]hcloud.FirewallRule{{
			DestinationIPs: mustParseCIDRs("10.0.0.0/24", "10.0.1.0/24"),
		}}, []hcloud.FirewallRule{{
			DestinationIPs: mustParseCIDRs("10.0.1.0/24", "10.0.0.0/24"),
		}}))
		assert.Equal(t, false, internal.HcloudFirewallRulesCompare([]hcloud.FirewallRule{{
			DestinationIPs: mustParseCIDRs("10.0.0.0/24"),
		}}, []hcloud.FirewallRule{{
			DestinationIPs: mustParseCIDRs("10.0.1.0/24"),
		}}))
		assert.Equal(t, false, internal.HcloudFirewallRulesCompare([]hcloud.FirewallRule{{
			DestinationIPs: mustParseCIDRs("10.0.0.0/24"),
		}}, []hcloud.FirewallRule{{
			DestinationIPs: mustParseCIDRs("10.0.0.0/24", "10.0.1.0/24"),
		}}))
	})

	t.Run("Protocol", func(t *testing.T) {
		assert.Equal(t, true, internal.HcloudFirewallRulesCompare([]hcloud.FirewallRule{{
			Protocol: hcloud.FirewallRuleProtocolTCP,
		}}, []hcloud.FirewallRule{{
			Protocol: hcloud.FirewallRuleProtocolTCP,
		}}))
		assert.Equal(t, false, internal.HcloudFirewallRulesCompare([]hcloud.FirewallRule{{
			Protocol: hcloud.FirewallRuleProtocolTCP,
		}}, []hcloud.FirewallRule{{
			Protocol: hcloud.FirewallRuleProtocolUDP,
		}}))
	})

	t.Run("Port", func(t *testing.T) {
		assert.Equal(t, true, internal.HcloudFirewallRulesCompare([]hcloud.FirewallRule{{
			Port: nil,
		}}, []hcloud.FirewallRule{{
			Port: nil,
		}}))
		assert.Equal(t, true, internal.HcloudFirewallRulesCompare([]hcloud.FirewallRule{{
			Port: internal.ValuePointer("80"),
		}}, []hcloud.FirewallRule{{
			Port: internal.ValuePointer("80"),
		}}))
		assert.Equal(t, false, internal.HcloudFirewallRulesCompare([]hcloud.FirewallRule{{
			Port: nil,
		}}, []hcloud.FirewallRule{{
			Port: internal.ValuePointer("80"),
		}}))
		assert.Equal(t, false, internal.HcloudFirewallRulesCompare([]hcloud.FirewallRule{{
			Port: internal.ValuePointer("80"),
		}}, []hcloud.FirewallRule{{
			Port: nil,
		}}))
		assert.Equal(t, false, internal.HcloudFirewallRulesCompare([]hcloud.FirewallRule{{
			Port: internal.ValuePointer("80"),
		}}, []hcloud.FirewallRule{{
			Port: internal.ValuePointer("81"),
		}}))
	})

	t.Run("Description", func(t *testing.T) {
		assert.Equal(t, true, internal.HcloudFirewallRulesCompare([]hcloud.FirewallRule{{
			Description: nil,
		}}, []hcloud.FirewallRule{{
			Description: nil,
		}}))
		assert.Equal(t, true, internal.HcloudFirewallRulesCompare([]hcloud.FirewallRule{{
			Description: internal.ValuePointer("desc"),
		}}, []hcloud.FirewallRule{{
			Description: internal.ValuePointer("desc"),
		}}))
		assert.Equal(t, false, internal.HcloudFirewallRulesCompare([]hcloud.FirewallRule{{
			Description: nil,
		}}, []hcloud.FirewallRule{{
			Description: internal.ValuePointer("desc"),
		}}))
		assert.Equal(t, false, internal.HcloudFirewallRulesCompare([]hcloud.FirewallRule{{
			Description: internal.ValuePointer("desc"),
		}}, []hcloud.FirewallRule{{
			Description: nil,
		}}))
		assert.Equal(t, false, internal.HcloudFirewallRulesCompare([]hcloud.FirewallRule{{
			Description: internal.ValuePointer("desc"),
		}}, []hcloud.FirewallRule{{
			Description: internal.ValuePointer("desc2"),
		}}))
	})
}

func mustParseCIDR(str string) net.IPNet {
	_, mask, err := net.ParseCIDR(str)
	if err != nil {
		panic(err)
	}
	if mask == nil {
		panic(fmt.Errorf("ip net %s is invalid: %w", str, err))
	}
	return *mask
}

func mustParseCIDRs(str ...string) []net.IPNet {
	return internal.SlicesMap(str, mustParseCIDR)
}
