package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bonedaddy/mychurnero/client"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "mychurnero"
	app.Usage = "automated churning application"
	app.Description = "mychurnero is an automated churning application designed to reduce the overhead in churning and create a totally automatic solution. this application provides no guarantee in benefits, and should be used with caution, and with great care"
	app.Commands = cli.Commands{
		&cli.Command{
			Name:  "sweep-dust",
			Usage: "sweeps all dust",
			Action: func(c *cli.Context) error {
				cl, err := client.NewClient(c.String("wallet.rpc_address"))
				if err != nil {
					return err
				}
				resp, err := cl.SweepDust(c.String("wallet.name"))
				if err != nil {
					return err
				}
				fmt.Printf("%#v\n", resp)
				return cl.Close()
			},
		},
		&cli.Command{
			Name:  "mining",
			Usage: "mining related commands",
			Subcommands: cli.Commands{
				&cli.Command{
					Name:  "start",
					Usage: "start mining",
					Action: func(c *cli.Context) error {
						cl, err := client.NewClient(c.String("wallet.rpc_address"))
						if err != nil {
							return err
						}
						if err := cl.StartMining(c.String("wallet.name"), c.Uint64("threads")); err != nil {
							return err
						}
						return nil
					},
					Flags: []cli.Flag{
						&cli.Uint64Flag{
							Name:  "threads",
							Usage: "number of threads to use for mining",
							Value: 2,
						},
					},
				},
				&cli.Command{
					Name:  "stop",
					Usage: "stop mining",
					Action: func(c *cli.Context) error {
						cl, err := client.NewClient(c.String("wallet.rpc_address"))
						if err != nil {
							return err
						}
						if err := cl.StopMining(c.String("wallet.name")); err != nil {
							return err
						}
						return cl.Close()
					},
				},
			},
		},
	}
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "wallet.name",
			Aliases: []string{"wn"},
			Usage:   "the wallet to use for churning",
			Value:   "testnetwallet123",
		},
		&cli.StringFlag{
			Name:    "wallet.rpc_address",
			Aliases: []string{"wrpc"},
			Usage:   "the endpoint address of the monero-wallet-rpc server",
			Value:   "http://127.0.0.1:6061/json_rpc",
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
