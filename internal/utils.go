package internal

import (
	"io/fs"
	"net"
	"os"
	"sort"

	"github.com/hetznercloud/hcloud-go/hcloud"
)

func SlicesConcat[T any](ins ...[]T) []T {
	result := []T{}
	for _, in := range ins {
		result = append(result, in...)
	}
	return result
}

func SlicesMap[I any, O any](in []I, fn func(elem I) O) []O {
	result := []O{}
	for _, elem := range in {
		result = append(result, fn(elem))
	}
	return result
}

func SlicesFlatMap[I any, O any](in []I, fn func(elem I) []O) []O {
	result := []O{}
	for _, elem := range in {
		result = append(result, fn(elem)...)
	}
	return result
}

func SlicesCompare[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func ValuePointer[T any](value T) *T {
	return &value
}

func ValuePointerCompare[T comparable](a, b *T) bool {
	if a == nil && b == nil {
		return true
	} else if a != nil && b != nil {
		return *a == *b
	} else {
		return false
	}
}

func WriteFileIfChanged(name string, content []byte, perm fs.FileMode) (bool, error) {
	bytes, err := os.ReadFile(name)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}

	if os.IsNotExist(err) || !SlicesCompare(content, bytes) {
		if err := os.WriteFile(name, content, perm); err != nil {
			return true, err
		}
		return true, nil
	} else {
		return false, nil
	}
}

func HcloudFirewallRulesCompare(a, b []hcloud.FirewallRule) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		ar := a[i]
		br := b[i]

		if ar.Direction != br.Direction {
			return false
		}
		if !compareIPNetSliceUnordered(ar.SourceIPs, br.SourceIPs) {
			return false
		}
		if !compareIPNetSliceUnordered(ar.DestinationIPs, br.DestinationIPs) {
			return false
		}
		if ar.Protocol != br.Protocol {
			return false
		}
		if !ValuePointerCompare(ar.Port, br.Port) {
			return false
		}
		if !ValuePointerCompare(ar.Description, br.Description) {
			return false
		}
	}

	return true
}

func compareIPNetSliceUnordered(a, b []net.IPNet) bool {
	as := SlicesMap(a, func(x net.IPNet) string { return x.String() })
	bs := SlicesMap(b, func(x net.IPNet) string { return x.String() })
	sort.Strings(as)
	sort.Strings(bs)
	return SlicesCompare(as, bs)
}
