package model

type CheckCIDRReq struct {
	BaseRequest
	IPv6            bool   `json:"ipv6" example:"false"`                                                 // ipv6开关
	PodCIDR         string `json:"podCidr" example:"10.244.0.0/16" binding:"required,cidrv4"`            // 容器组网段
	ServiceCIDR     string `json:"serviceCidr" example:"10.96.0.0/16" binding:"required,cidrv4"`         // 服务网段
	PodCIDRIPv6     string `json:"podCidrIpv6" example:"fc00::/48" binding:"required_if=IPv6 true"`      // ipv6容器组网段, ipv6为true时有效
	ServiceCIDRIPv6 string `json:"serviceCidrIpv6" example:"fd00::/108" binding:"required_if=IPv6 true"` // ipv6服务网
}

type CheckCIDRResp struct {
	BaseResponse
}
