package main

import (
	"context"
	"errors"
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
		zlog.Info("upup_generic_mapping", zlog.Any("mapping", gm))
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
			zlog.Info("upup_add_mapping", zlog.Any("external_port", eport), zlog.Any("internal_port", iport), zlog.Any("internal_ip", ipv4))
			return nil
		},
	}
}

func DeleteCmd() cli.Command {
	return cli.Command{
		Name:    "delete",
		Aliases: []string{"d"},
		Usage:   "delete upnp port mapping",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "proto",
				Value: "TCP",
				Usage: "TCP or UDP protocol,default TCP",
			},
			&cli.Uint64Flag{
				Name:  "eport",
				Value: 0,
				Usage: "external port will be delete",
			},
		},
		Action: func(c *cli.Context) error {
			proto := c.String("proto")
			eport := c.Uint64("eport")
			t := upnp.FindUPNPTargetCtx(context.Background())
			if t == "" {
				return errors.New("not found valid upnp target")
			}
			w, err := upnp.NewUpnpWrapper(t)
			if err != nil {
				return err
			}
			if errs := w.DeletePortMappingCtx(context.Background(), upnp.Proto(proto), uint16(eport)); len(errs) > 0 {
				return errs[0]
			}
			return nil
		},
	}
}
