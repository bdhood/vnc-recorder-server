package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"os"
	"path"
)

func main() {
	app := &cli.App{
		Name:    path.Base(os.Args[0]),
		Usage:   "A tcp server that can connect to multiple vnc servers and record the sessions",
		Version: "0.1.0",
		Action: server_init,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "ffmpeg",
				Value:   "ffmpeg",
				Usage:   "Which ffmpeg executable to use",
				EnvVars: []string{"FFMPEG_BIN"},
			},
			&cli.StringFlag{
				Name:    "host",
				Value:   "0.0.0.0",
				Usage:   "Interface to listen on",
				EnvVars: []string{"VRS_HOST"},
			},
			&cli.IntFlag{
				Name:    "port",
				Value:   25192,
				Usage:   "Port to listen on",
				EnvVars: []string{"VRS_PORT"},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println("[ server ] server crash")
		fmt.Println(err.Error())
	}
}
