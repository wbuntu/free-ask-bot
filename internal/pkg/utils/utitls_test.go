package utils

import (
	"net"
	"testing"
)

func TestGetRandString(t *testing.T) {
	itemList := []int{
		100,
		1000,
		10000,
	}
	for _, count := range itemList {
		m := make(map[string]struct{}, count)
		for i := 0; i < count; i++ {
			key := GetRandString(10)
			if _, ok := m[key]; ok {
				t.Errorf("count: %d size: %d: duplicated key occurred: %s", count, len(m), key)
				break
			}
			m[key] = struct{}{}
		}
		t.Logf("count: %d passed", count)
	}
}

func TestCheckCIDROverlap(t *testing.T) {
	type cidrInfo struct {
		a  string
		b  string
		ok bool
	}
	itemList := []cidrInfo{
		{"10.0.0.0/8", "172.16.0.0/12", true},
		{"172.16.0.0/12", "192.16.0.0/16", true},
		{"10.0.0.0/8", "10.0.0.0/24", false},
		{"172.16.0.0/16", "172.16.0.0/24", false},
		{"fd00::/48", "fc00::/48", true},
		{"fd00::/108", "fc00::/108", true},
		{"fd00::/48", "fd00::/108", false},
	}
	for _, item := range itemList {
		ok := CheckCIDROverlap(item.a, item.b) == nil
		if ok != item.ok {
			t.Errorf("check cidr overlap: %s %s failed", item.a, item.b)
			break
		}
		t.Logf("check cidr overlap: %s %s passed", item.a, item.b)
	}
}

func TestCheckCIDRContains(t *testing.T) {
	type cidrPair struct {
		a  string
		b  string
		ok bool
	}
	itemList := []cidrPair{
		{"10.0.0.0/8", "172.16.0.0/12", false},
		{"172.16.0.0/12", "192.16.0.0/16", false},
		{"10.0.0.0/8", "10.0.0.0/24", true},
		{"172.16.0.0/16", "172.16.0.0/24", true},
		{"fd00::/48", "fc00::/48", false},
		{"fd00::/108", "fc00::/108", false},
		{"fd00::/48", "fd00::/108", true},
	}
	for _, item := range itemList {
		ok := CheckCIDRContains(item.a, item.b) == nil
		if ok != item.ok {
			t.Errorf("check cidr contains: %s %s failed", item.a, item.b)
			break
		}
		t.Logf("check cidr contains: %s %s passed", item.a, item.b)
	}
}

func TestCheckK8SPodCIDR(t *testing.T) {
	type cidrInfo struct {
		v  string
		ok bool
	}
	itemListV4 := []cidrInfo{
		{"10.0.0.0/8", true},
		{"172.16.0.0/12", true},
		{"192.168.0.0/16", true},
		{"10.0.0.0/21", false},
		{"172.16.0.0.0/16", false},
	}
	for _, item := range itemListV4 {
		ok := CheckK8SPodCIDR(item.v, net.IPv4len) == nil
		if ok != item.ok {
			t.Errorf("check k8s pod cidr: %s failed", item.v)
			break
		}
		t.Logf("check k8s pod cidr: %s passed", item.v)
	}
	itemListV6 := []cidrInfo{
		{"fd00::/48", true},
		{"fc00::/48", true},
		{"fd00::/116", true},
		{"fc00::/117", false},
		{"fa00::/108", false},
		{"fb00::/108", false},
	}
	for _, item := range itemListV6 {
		ok := CheckK8SPodCIDR(item.v, net.IPv6len) == nil
		if ok != item.ok {
			t.Errorf("check k8s pod cidr: %s failed", item.v)
			break
		}
		t.Logf("check k8s pod cidr: %s passed", item.v)
	}
}

func TestCheckK8SServiceCIDR(t *testing.T) {
	type cidrInfo struct {
		v  string
		ok bool
	}
	itemListV4 := []cidrInfo{
		{"10.0.0.0/16", true},
		{"172.16.0.0/16", true},
		{"192.168.0.0/16", true},
		{"10.0.0.0/15", false},
		{"172.16.0.0/25", false},
		{"192.168.0.0.0/25", false},
	}
	for _, item := range itemListV4 {
		ok := CheckK8SServiceCIDR(item.v, net.IPv4len) == nil
		if ok != item.ok {
			t.Errorf("check k8s service cidr: %s failed", item.v)
			break
		}
		t.Logf("check k8s service cidr: %s passed", item.v)
	}
	itemListV6 := []cidrInfo{
		{"fd00::/108", true},
		{"fc00::/108", true},
		{"fd00::/120", true},
		{"fc00::/121", false},
		{"fa00::/120", false},
		{"fb00::/120", false},
	}
	for _, item := range itemListV6 {
		ok := CheckK8SServiceCIDR(item.v, net.IPv6len) == nil
		if ok != item.ok {
			t.Errorf("check k8s service cidr: %s failed", item.v)
			break
		}
		t.Logf("check k8s pod service: %s passed", item.v)
	}
}

func TestCheckVPCCIDR(t *testing.T) {
	type cidrInfo struct {
		a  string
		b  string
		c  string
		ok bool
	}
	itemList := []cidrInfo{
		{"10.0.0.0/8", "10.0.0.0/16", "10.1.0.0/16", true},
		{"172.16.0.0/12", "172.16.0.0/16", "172.17.0.0/16", true},
		{"192.168.0.0/16", "192.168.0.0/17", "192.168.0.0/18", false},
		{"fd00::/48", "fd00::/108", "fc00::/108", false},
		{"fc00::/48", "fc00::/108", "fd00::/108", false},
	}
	for _, item := range itemList {
		ok := CheckVPCCIDR(item.a, item.b, item.c) == nil
		if ok != item.ok {
			t.Errorf("check vpc cidr: %s %s %s failed", item.a, item.b, item.c)
			break
		}
		t.Logf("check vpc cidr: %s %s %s passed", item.a, item.b, item.c)
	}
}

func TestCheckName(t *testing.T) {
	type nameInfo struct {
		v  string
		ok bool
	}
	itemList := []nameInfo{
		{"a", true},
		{"集", true},
		{"集群-001_v1.1", true},
		{"集群-002_v2.2", true},
		{"abshfjgyetfdskfhlwaJEFJWELFJLEWJF3NFDKJFHAJSFLJSALdfjasljflaskjfdlaffabshfjgyetfdskfhlwaJEFJWELFJLEWJF3NFDKJFHAJSFLJSAsj3jf38fy", true},
		{"abshfjgyetfLJSAsj3jf38fy_", false},
		{"集群集群集群集群集群集群集群集群集群集群集群集群集群集群集群集1", true},
		{"abcdef^^", false},
		{"002*77^", false},
		{"1234abcd", false},
		{"--__vvv", false},
		{".--__vvv", false},
	}
	for _, item := range itemList {
		ok := CheckName(item.v) == nil
		if ok != item.ok {
			t.Errorf("check name: %s failed", item.v)
			break
		}
		t.Logf("check name: %s passed", item.v)
	}
}

func TestCheckPassword(t *testing.T) {
	type passwordInfo struct {
		v  string
		ok bool
	}
	itemList := []passwordInfo{
		{"abc12345", true},
		{",?%12345", true},
		{"ABC12345[]?", true},
		{"abc", false},
		{"abcdefgh", false},
		{"12345678", false},
		{"!@%^-_=+", false},
	}
	for _, item := range itemList {
		ok := CheckPassword(item.v) == nil
		if ok != item.ok {
			t.Errorf("check password: %s failed", item.v)
			break
		}
		t.Logf("check password: %s passed", item.v)
	}
}
