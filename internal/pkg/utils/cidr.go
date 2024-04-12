package utils

import (
	"net"

	"github.com/pkg/errors"
)

// CheckCIDROverlap 检查两个CIDR是否重叠，支持IPv4与IPv6
func CheckCIDROverlap(cidrAStr string, cidrBStr string) error {
	_, cidrA, err := net.ParseCIDR(cidrAStr)
	if err != nil {
		return err
	}
	_, cidrB, err := net.ParseCIDR(cidrBStr)
	if err != nil {
		return err
	}
	if cidrA.Contains(cidrB.IP) || cidrB.Contains(cidrA.IP) {
		return errors.New("cidr overlap")
	}
	return nil
}

// CheckCIDRContains 检查第一个CIDR是否包含第二个CIDR
func CheckCIDRContains(cidrAStr string, cidrBStr string) error {
	_, cidrA, err := net.ParseCIDR(cidrAStr)
	if err != nil {
		return err
	}
	onesA, _ := cidrA.Mask.Size()
	_, cidrB, err := net.ParseCIDR(cidrBStr)
	if err != nil {
		return err
	}
	onesB, _ := cidrB.Mask.Size()
	if cidrA.Contains(cidrB.IP) && onesA <= onesB {
		return nil
	}
	return errors.New("cidr out of range")
}

type k8sCIDR struct {
	cidr *net.IPNet
	ones int
	bits int
}

var k8sPodCIDRMap = map[int][]*k8sCIDR{}
var k8sServiceCIDRMap = map[int][]*k8sCIDR{}

func init() {
	podCIDRListV4 := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	}
	podCIDRListV6 := []string{
		"fd00::/48",
		"fc00::/48",
	}
	for i := range podCIDRListV4 {
		_, cidr, _ := net.ParseCIDR(podCIDRListV4[i])
		ones, bits := cidr.Mask.Size()
		k8sPodCIDRMap[net.IPv4len] = append(k8sPodCIDRMap[net.IPv4len], &k8sCIDR{
			cidr: cidr,
			ones: ones,
			bits: bits,
		})
	}
	for i := range podCIDRListV6 {
		_, cidr, _ := net.ParseCIDR(podCIDRListV6[i])
		ones, bits := cidr.Mask.Size()
		k8sPodCIDRMap[net.IPv6len] = append(k8sPodCIDRMap[net.IPv6len], &k8sCIDR{
			cidr: cidr,
			ones: ones,
			bits: bits,
		})
	}
	serviceCIDRListV4 := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	}
	serviceCIDRListV6 := []string{
		"fd00::/108",
		"fc00::/108",
	}
	for i := range serviceCIDRListV4 {
		_, cidr, _ := net.ParseCIDR(serviceCIDRListV4[i])
		ones, bits := cidr.Mask.Size()
		k8sServiceCIDRMap[net.IPv4len] = append(k8sServiceCIDRMap[net.IPv4len], &k8sCIDR{
			cidr: cidr,
			ones: ones,
			bits: bits,
		})
	}
	for i := range serviceCIDRListV6 {
		_, cidr, _ := net.ParseCIDR(serviceCIDRListV6[i])
		ones, bits := cidr.Mask.Size()
		k8sServiceCIDRMap[net.IPv6len] = append(k8sServiceCIDRMap[net.IPv6len], &k8sCIDR{
			cidr: cidr,
			ones: ones,
			bits: bits,
		})
	}

}

// CheckK8SPodCIDR 非VPC分配IP场景，检查给定的CIDR是否落在k8s允许的有效的CIDR范围内，IPv4掩码<=20，IPv6掩码<=116
func CheckK8SPodCIDR(cidrStr string, ipAddrLength int) error {
	_, cidr, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return err
	}
	// CIDR: 172.16.0.0/12 ones: 12 bits: 32 包含 CIDR: 172.16.0.0/16 ones: 16 bits: 32
	ones, _ := cidr.Mask.Size()
	if ipAddrLength == net.IPv6len && ones <= 116 || ipAddrLength == net.IPv4len && ones <= 20 {
		for _, v := range k8sPodCIDRMap[ipAddrLength] {
			if v.cidr.Contains(cidr.IP) && v.ones <= ones {
				return nil
			}
		}
	}
	return errors.New("out of valid cidr range")
}

// CheckK8SServiceCIDR 非VPC分配IP场景，检查给定的CIDR是否落在k8s允许的有效的CIDR范围内，IPv4掩码16~24，IPv6掩码108~120
func CheckK8SServiceCIDR(cidrStr string, ipAddrLength int) error {
	_, cidr, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return err
	}
	// CIDR: 172.16.0.0/20 ones: 20 bits: 32 包含 CIDR: 172.16.0.0/24 ones: 24 bits: 32
	ones, _ := cidr.Mask.Size()
	if ipAddrLength == net.IPv6len && ones >= 108 && ones <= 120 || ipAddrLength == net.IPv4len && ones >= 16 && ones <= 24 {
		for _, v := range k8sServiceCIDRMap[ipAddrLength] {
			if v.cidr.Contains(cidr.IP) && v.ones <= ones {
				return nil
			}
		}
	}
	return errors.New("out of valid cidr range")
}

// CheckVPCCIDR VPC分配IP场景，podCIDR与serviceCIDR需要落在vpcCDIR内且不冲突
func CheckVPCCIDR(vpcCIDR string, podCIDR string, serviceCIDR string) error {
	if err := CheckCIDROverlap(podCIDR, serviceCIDR); err != nil {
		return err
	}
	if err := CheckCIDRContains(vpcCIDR, podCIDR); err != nil {
		return err
	}
	if err := CheckCIDRContains(vpcCIDR, serviceCIDR); err != nil {
		return err
	}
	return nil
}
