package main

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"time"

	vnc "github.com/amitbet/vnc2video"
	guuid "github.com/google/uuid"
)

func recorder_proc(conn net.Conn, cmd CommandStart, sessionId guuid.UUID) {
	address := fmt.Sprintf("%s:%d", cmd.Addr, cmd.Port)

	dialer, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		session_log(sessionId, "connection to VNC host failed.")
		session_log(sessionId, err.Error())
		server_send(sessionId, "failed")
		session_remove(sessionId)
		return
	}
	defer dialer.Close()

	// Negotiate connection with the server.
	cchServer := make(chan vnc.ServerMessage)
	cchClient := make(chan vnc.ClientMessage)
	errorCh := make(chan error)

	var secHandlers []vnc.SecurityHandler
	if cmd.Password == "" {
		secHandlers = []vnc.SecurityHandler{
			&vnc.ClientAuthNone{},
		}
	} else {
		secHandlers = []vnc.SecurityHandler{
			&vnc.ClientAuthVNC{Password: []byte(cmd.Password)},
		}
	}

	ccflags := &vnc.ClientConfig{
		SecurityHandlers: secHandlers,
		DrawCursor:       true,
		PixelFormat:      vnc.PixelFormat32bit,
		ClientMessageCh:  cchClient,
		ServerMessageCh:  cchServer,
		Messages:         vnc.DefaultServerMessages,
		Encodings: []vnc.Encoding{
			&vnc.RawEncoding{},
			&vnc.TightEncoding{},
			&vnc.HextileEncoding{},
			&vnc.ZRLEEncoding{},
			&vnc.CopyRectEncoding{},
			&vnc.CursorPseudoEncoding{},
			&vnc.CursorPosPseudoEncoding{},
			&vnc.ZLibEncoding{},
			&vnc.RREEncoding{},
		},
		ErrorCh: errorCh,
	}

	vncConnection, err := vnc.Connect(context.Background(), dialer, ccflags)
	defer vncConnection.Close()
	if err != nil {
		session_log(sessionId, "connection negotiation to VNC host failed")
		session_log(sessionId, err.Error())
		server_send(sessionId, "failed")
		session_remove(sessionId)
		return
	}

	session_log(sessionId, "connected to " + address)

	screenImage := vncConnection.Canvas

	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		session_log(sessionId, "ffmpeg binary not found")
		session_log(sessionId, err.Error())
		server_send(sessionId, "failed")
		session_remove(sessionId)
		return
	}

	vcodec := &X264ImageCustomEncoder{
		FFMpegBinPath:      ffmpegPath,
		Framerate:          cmd.FrameRate,
		ConstantRateFactor: cmd.ConstantRateFactor,
	}

	//goland:noinspection GoUnhandledErrorResult
	go vcodec.Run(sessionId, cmd.Output)

	for _, enc := range ccflags.Encodings {
		myRenderer, ok := enc.(vnc.Renderer)

		if ok {
			myRenderer.SetTargetImage(screenImage)
		}
	}

	vncConnection.SetEncodings([]vnc.EncodingType{
		vnc.EncCursorPseudo,
		vnc.EncPointerPosPseudo,
		vnc.EncCopyRect,
		vnc.EncTight,
		vnc.EncZRLE,
		vnc.EncHextile,
		vnc.EncZlib,
		vnc.EncRRE,
	})

	go func() {
		for {
			timeStart := time.Now()

			vcodec.Encode(sessionId, screenImage.Image)

			timeTarget := timeStart.Add((1000 / time.Duration(vcodec.Framerate)) * time.Millisecond)
			timeLeft := timeTarget.Sub(time.Now())
			if timeLeft > 0 {
				time.Sleep(timeLeft)
			}
		}
	}()

	session_log(sessionId, "recording started")
	server_send(sessionId, "started")
	for {
		select {
		case err := <-errorCh:
			session_log(sessionId, "ERROR:")
			session_log(sessionId, err.Error())
			return
		case msg := <-cchClient:
			session_log(sessionId, msg.String())
		case msg := <-cchServer:
			if msg.Type() == vnc.FramebufferUpdateMsgType {
				reqMsg := vnc.FramebufferUpdateRequest{Inc: 1, X: 0, Y: 0, Width: vncConnection.Width(), Height: vncConnection.Height()}
				reqMsg.Write(vncConnection)
			}
		}
		if session_get_enabled(sessionId) == false {
			vcodec.Close(sessionId)
			time.Sleep(time.Second * 2)
			session_log(sessionId, "recording stopped")
			server_send(sessionId, "stopped")
			session_remove(sessionId)
			return
		}
	}
	return
}