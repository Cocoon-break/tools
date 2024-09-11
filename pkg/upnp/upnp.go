package upnp

import (
	"context"
	"fmt"

	"github.com/huin/goupnp"
	"github.com/huin/goupnp/dcps/internetgateway1"
)

type ConnectionClient interface {
	GetGenericPortMappingEntryCtx(ctx context.Context, NewPortMappingIndex uint16) (NewRemoteHost string, NewExternalPort uint16, NewProtocol string, NewInternalPort uint16, NewInternalClient string, NewEnabled bool, NewPortMappingDescription string, NewLeaseDuration uint32, err error)
	GetSpecificPortMappingEntryCtx(ctx context.Context, NewRemoteHost string, NewExternalPort uint16, NewProtocol string) (NewInternalPort uint16, NewInternalClient string, NewEnabled bool, NewPortMappingDescription string, NewLeaseDuration uint32, err error)
	AddPortMappingCtx(ctx context.Context, NewRemoteHost string, NewExternalPort uint16, NewProtocol string, NewInternalPort uint16, NewInternalClient string, NewEnabled bool, NewPortMappingDescription string, NewLeaseDuration uint32) (err error)
	DeletePortMappingCtx(ctx context.Context, NewRemoteHost string, NewExternalPort uint16, NewProtocol string) error
}
type MappingEntry struct {
	Proto          string `json:"proto"`
	ExternalPort   uint16 `json:"external_port"`
	InternalPort   uint16 `json:"internal_port"`
	InternalClient string `json:"internal_ip"`
}

type Proto string

const (
	TCP Proto = "TCP"
	UDP Proto = "UDP"
)

type UPNPWrapper struct {
	clients []ConnectionClient
}

func NewUpnpWrapper(tag string) (*UPNPWrapper, error) {
	clients := []ConnectionClient{}
	switch tag {
	case internetgateway1.URN_WANIPConnection_1:
		wanIpClients, _, err := internetgateway1.NewWANIPConnection1Clients()
		if err != nil {
			return nil, err
		}
		for _, c := range wanIpClients {
			clients = append(clients, c)
		}
	case internetgateway1.URN_WANPPPConnection_1:
		wanPPPClients, _, err := internetgateway1.NewWANPPPConnection1Clients()
		if err != nil {
			return nil, err
		}
		for _, c := range wanPPPClients {
			clients = append(clients, c)
		}
	default:
		return nil, fmt.Errorf("tag:%s not support", tag)
	}
	if len(clients) == 0 {
		return nil, fmt.Errorf("tag:%s no clients", tag)
	}
	return &UPNPWrapper{clients: clients}, nil
}

func (u *UPNPWrapper) GetClientsCount() int {
	return len(u.clients)
}

func (u *UPNPWrapper) GetGenericPortMappingEntry() []MappingEntry {
	return u.GetGenericPortMappingEntryCtx(context.Background())
}

// 获取所有的upnp端口映射
func (u *UPNPWrapper) GetGenericPortMappingEntryCtx(ctx context.Context) []MappingEntry {
	mappingList := make([]MappingEntry, 0, 8)
	for _, client := range u.clients {
		for i := 0; true; i++ {
			_, externalPort, protocol, internalPort, internalClient, _, _, _, err := client.GetGenericPortMappingEntryCtx(ctx, uint16(i))
			if err != nil {
				break
			}
			mappingList = append(mappingList, MappingEntry{
				Proto:          protocol,
				ExternalPort:   externalPort,
				InternalPort:   internalPort,
				InternalClient: internalClient,
			})
		}
	}
	return mappingList
}

func (u *UPNPWrapper) GetSpecificPortMappingEntry(proto Proto, externalPorts []uint16) []MappingEntry {
	return u.GetSpecificPortMappingEntryCtx(context.Background(), proto, externalPorts)
}

// 获取指定协议和指定外网端口列表的upnp端口映射
func (u *UPNPWrapper) GetSpecificPortMappingEntryCtx(ctx context.Context, proto Proto, externalPorts []uint16) []MappingEntry {
	mappingList := make([]MappingEntry, 0, 8)
	for _, client := range u.clients {
		for _, p := range externalPorts {
			internalPort, internalClient, _, _, _, err := client.GetSpecificPortMappingEntryCtx(ctx, "", uint16(p), string(proto))
			if err != nil {
				break
			}
			mappingList = append(mappingList, MappingEntry{
				Proto:          string(proto),
				ExternalPort:   p,
				InternalPort:   internalPort,
				InternalClient: internalClient,
			})
		}
	}
	return mappingList
}

func (u *UPNPWrapper) AddPortMapping(proto Proto, externalPort uint16, internalPort uint16, internalClient string, description string) []error {
	return u.AddPortMappingCtx(context.Background(), proto, externalPort, internalPort, internalClient, description)
}

// 添加upnp端口映射
func (u *UPNPWrapper) AddPortMappingCtx(ctx context.Context, proto Proto, externalPort uint16, internalPort uint16, internalClient string, description string) []error {
	if description == "" {
		description = "wsp-tools"
	}
	errs := make([]error, 0, len(u.clients))
	for _, client := range u.clients {
		err := client.AddPortMappingCtx(ctx, "", externalPort, string(proto), internalPort, internalClient, true, description, 0)
		if err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errs
}

func (u *UPNPWrapper) DeletePortMapping(proto Proto, externalPort uint16) []error {
	return u.DeletePortMappingCtx(context.Background(), proto, externalPort)
}

// 删除端口映射
func (u *UPNPWrapper) DeletePortMappingCtx(ctx context.Context, proto Proto, externalPort uint16) []error {
	errs := make([]error, 0, len(u.clients))
	for _, client := range u.clients {
		err := client.DeletePortMappingCtx(ctx, "", externalPort, string(proto))
		if err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errs
}

// 找到upnp支持的服务类型，目前仅支持urn:schemas-upnp-org:service:WANIPConnection:1"和"urn:schemas-upnp-org:service:WANPPPConnection:1"
// 如果返回为空字符串，则表示不支持upnp
func FindUPNPTargetCtx(ctx context.Context) string {
	tags := []string{internetgateway1.URN_WANIPConnection_1, internetgateway1.URN_WANPPPConnection_1}
	matchTag := ""
	for _, tag := range tags {
		devices, err := goupnp.DiscoverDevicesCtx(ctx, tag)
		if err != nil {
			continue
		}
		if len(devices) > 0 {
			matchTag = tag
			break
		}
	}
	return matchTag
}

// 判断设备是否支持upnp
func IsUpnpAvailable() bool {
	return FindUPNPTargetCtx(context.Background()) != ""
}
