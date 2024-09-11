package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"tools/pkg/upnp"
	"tools/pkg/zlog"

	"github.com/huin/goupnp"
	"github.com/huin/goupnp/dcps/internetgateway1"
)

var (
	_internalIp   = ""
	_internalPort = 0
	_externalPort = 0
	_proto        = ""
	_action       = ""
)

func init() {
	flag.StringVar(&_internalIp, "internal_ip", "", "设备上网卡的内网IP")
	flag.StringVar(&_proto, "proto", "TCP", "协议,TCP/UDP,默认TCP")
	flag.StringVar(&_action, "action", "visit", "操作类型,visit/add/delete,默认visit")
	flag.IntVar(&_internalPort, "internal_port", 0, "upnp要映射的内网端口")
	flag.IntVar(&_externalPort, "external_port", 0, "upnp要映射的外网端口")
}

func main() {
	flag.Parse()
	zlog.Info("upnp_params",
		zlog.Any("internal_ip", _internalIp),
		zlog.Any("internal_port", _internalPort),
		zlog.Any("external_port", _externalPort),
		zlog.Any("proto", _proto),
		zlog.Any("action", _action),
	)
	switch _action {
	case "visit":
		visit()
	case "add":
		add()
	case "delete":
		delete()
	}
}

func visit() {
	t := discoverUPNP()
	w, err := upnp.NewUpnpWrapper(t)
	if err != nil {
		zlog.Error("NewUpnpClients", zlog.Any("err", err))
		return
	}
	zlog.Info("upup_clients", zlog.Any("tag", t), zlog.Any("clients_cnt", w.GetClientsCount()))
	gm := w.GetGenericPortMappingEntryCtx(context.Background())
	zlog.Info("upup_generic_mapping", zlog.Any("mapping", gm))
	ports := []uint16{16818, 16823, 16718, 16723, 8282, 7843}
	w.GetSpecificPortMappingEntryCtx(context.Background(), upnp.Proto(_proto), ports)
	zlog.Info("upup_specific_mapping", zlog.Any("ports", ports), zlog.Any("mapping", gm))
}
func add() {
	t := discoverUPNP()
	w, err := upnp.NewUpnpWrapper(t)
	if err != nil {
		zlog.Error("NewUpnpClients", zlog.Any("err", err))
		return
	}
	zlog.Info("upup_clients", zlog.Any("tag", t), zlog.Any("clients_cnt", w.GetClientsCount()))
	// ctx, fn := context.WithTimeout(context.Background(), 10*time.Second)
	w.AddPortMappingCtx(context.Background(), upnp.Proto(_proto), uint16(_externalPort), uint16(_internalPort), _internalIp, "demo for upnp")
	zlog.Info("upup_add_mapping", zlog.Any("external_port", _externalPort), zlog.Any("internal_port", _internalPort), zlog.Any("internal_ip", _internalIp))

}

func delete() {
	t := discoverUPNP()
	w, err := upnp.NewUpnpWrapper(t)
	if err != nil {
		zlog.Error("NewUpnpClients", zlog.Any("err", err))
		return
	}
	zlog.Info("upup_clients", zlog.Any("tag", t), zlog.Any("clients_cnt", w.GetClientsCount()))
	ctx, fn := context.WithTimeout(context.Background(), 10*time.Second)
	w.DeletePortMappingCtx(ctx, upnp.Proto(_proto), uint16(_externalPort))
	zlog.Info("upup_delete_mapping", zlog.Any("external_port", _externalPort))
	fn()
}

func discoverUPNP() string {
	// URN_WANIPConnection_1
	// URN_WANPPPConnection_1
	tags := []string{internetgateway1.URN_WANIPConnection_1, internetgateway1.URN_WANPPPConnection_1}
	matchTag := ""
	for _, tag := range tags {
		devices, err := goupnp.DiscoverDevices(tag)
		if err != nil {
			fmt.Printf("tag: %s err:%s \n", tag, err.Error())
			continue
		}
		fmt.Printf("tag: %s device_cnt:%d \n", tag, len(devices))
		for _, device := range devices {
			if device.Err != nil {
				log.Println("failed to discover device. skipping.", device.Err, "url", device.Location.String())
				continue
			}
			fmt.Printf("%s\n", device.Location.Host)

			fmt.Printf("    Devices\n")
			device.Root.Device.VisitDevices(func(dev *goupnp.Device) {
				fmt.Printf("        %s - %s - %s\n", dev.FriendlyName, dev.Manufacturer, dev.DeviceType)
			})

			fmt.Printf("    Services:\n")
			device.Root.Device.VisitServices(func(svc *goupnp.Service) {
				fmt.Printf("        %s\n", svc.String())
			})
		}
		if len(devices) > 0 {
			matchTag = tag
			break
		}
	}
	return matchTag
}
