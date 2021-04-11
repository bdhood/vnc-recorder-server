# VNC Recorder Server

> Fork of https://github.com/saily/vnc-recorder

> this is wip, don't use in production!

Record [VNC] screens to mp4 video using [ffmpeg]. Thanks to
[amitbet for providing his vnc2video](https://github.com/amitbet/vnc2video)
library which made this wrapper possible.

## Use

   docker run -it widerin/vnc-recorder --help

```
   NAME:
      vnc-recorder-server - A tcp server that can connect to multiple vnc servers and record the sessions

   USAGE:
      vnc-recorder-server [global options] command [command options] [arguments...]

   VERSION:
      0.1.0

   COMMANDS:
      help, h  Shows a list of commands or help for one command

   GLOBAL OPTIONS:
      --ffmpeg value  Which ffmpeg executable to use (default: "ffmpeg") [%FFMPEG_BIN%]
      --host value    Interface to listen on (default: "0.0.0.0") [%VRS_HOST%]
      --port value    Port to listen on (default: 25192) [%VRS_PORT%]
      --help, -h      show help (default: false)
      --version, -v   print the version (default: false)
```

## Build

```
docker-compose up
```