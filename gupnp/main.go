package main

import (
	"context"
	"errors"
	"math"
	"os"

	"tools/pkg/upnp"
	"tools/pkg/zlog"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "upnp"
	app.Usage = "upnp port mapping control"
	app.Version = "0.0.1"
	app.Commands = []cli.Command{
		AddCmd(),
		DeleteCmd(),
	}
	app.Action = func(c *cli.Context) error {
		t := upnp.FindUPNPTargetCtx(context.Background())
		if t == "" {
			return errors.New("not found upnp")
		}
		w, err := upnp.NewUpnpWrapper(t)
		if err != nil {
			return err
		}
		gm := w.GetGenericPortMappingEntryCtx(context.Background())
		for _, m := range gm {
			zlog.Info("upup_generic_mapping", zlog.Any("mapping", m))
		}
		return nil
	}
	if err := app.Run(os.Args); err != nil {
		zlog.Fatal("upnp_run", zlog.Any("err", err))
	}
}

func AddCmd() cli.Command {
	return cli.Command{
		Name:    "add",
		Aliases: []string{"a"},
		Usage:   "add upnp port mapping",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "proto",
				Value: "TCP",
				Usage: "TCP or UDP protocol,default TCP",
			},
			&cli.StringFlag{
				Name:  "interface",
				Value: "eth0",
				Usage: "the network interface name",
			},
			&cli.Uint64Flag{
				Name:  "eport",
				Value: 0,
				Usage: "external port for upnp mapping",
			},
			&cli.Uint64Flag{
				Name:  "iport",
				Value: 0,
				Usage: "internal port for upnp mapping",
			},
		},
		Action: func(c *cli.Context) error {
			proto := c.String("proto")
			i := c.String("interface")
			eport := c.Uint64("eport")
			iport := c.Uint64("iport")
			if !isValidPort(eport) || !isValidPort(iport) {
				return errors.New("invalid port")
			}
			ipv4, err := GetIpV4ByName(i)
			if err != nil {
				return err
			}
			t := upnp.FindUPNPTargetCtx(context.Background())
			if t == "" {
				return errors.New("not found valid upnp target")
			}
			w, err := upnp.NewUpnpWrapper(t)
			if err != nil {
				return err
			}
			if errs := w.AddPortMappingCtx(context.Background(), upnp.Proto(proto), uint16(eport), uint16(iport), ipv4, "wsp for upnp"); len(errs) > 0 {
				return errs[0]
			}
			zlog.Info("upup_add_mapping_success", zlog.Any("proto", proto), zlog.Any("external_port", eport), zlog.Any("internal_port", iport), zlog.Any("internal_ip", ipv4))
			return nil
		},
	}
}

func DeleteCmd() cli.Command {
	return cli.Command{
		Name:    "delete",
		Aliases: []string{"d"},
		Usage:   "delete upnp port mapping by batch",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "proto",
				Value: "TCP",
				Usage: "TCP or UDP protocol,default TCP",
			},
			&cli.Int64SliceFlag{
				Name:  "eport",
				Usage: "--eport 1111 --eport 2222 external port [1111,2222] will be delete",
			},
			&cli.Int64SliceFlag{
				Name:  "iport",
				Usage: "--eport 1111 --eport 2222 internal port [1111,2222] will be delete",
			},
		},
		Action: func(c *cli.Context) error {
			proto := c.String("proto")
			eports := c.Int64Slice("eport")
			iports := c.Int64Slice("iport")
			for _, e := range eports {
				if !isValidPort(uint64(e)) {
					return errors.New("invalid external port")
				}
			}
			for _, i := range iports {
				if !isValidPort(uint64(i)) {
					return errors.New("invalid internal port")
				}
			}
			t := upnp.FindUPNPTargetCtx(context.Background())
			if t == "" {
				return errors.New("not found valid upnp target")
			}
			w, err := upnp.NewUpnpWrapper(t)
			if err != nil {
				return err
			}
			gm := w.GetGenericPortMappingEntryCtx(context.Background())
			shouldDeleteMap := make(map[string]upnp.MappingEntry, len(gm))
			for _, m := range gm {
				if m.Proto != proto {
					continue
				}
				for _, e := range eports {
					if m.ExternalPort == uint16(e) {
						shouldDeleteMap[m.Uuid] = m
					}
				}
				for _, e := range iports {
					if m.InternalPort == uint16(e) {
						shouldDeleteMap[m.Uuid] = m
					}
				}
			}
			for _, m := range shouldDeleteMap {
				if errs := w.DeletePortMappingCtx(context.Background(), upnp.Proto(m.Proto), m.ExternalPort); len(errs) > 0 {
					zlog.Error("upnp_delete_mapping failed", zlog.Any("mapping", m), zlog.Any("err", errs[0]))
					continue
				}
				zlog.Info("upnp_delete_mapping success", zlog.Any("mapping", m))
			}
			return nil
		},
	}
}

func isValidPort(port uint64) bool {
	return port > 0 && port <= math.MaxUint16
}
