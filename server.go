package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "github.com/urfave/cli/v2"
    "net"
    "strconv"
    "strings"

    guuid "github.com/google/uuid"
)

type Command struct {
    Command string
}

type CommandStart struct {
    Command string
    Addr string
    Port int
    Password string
    Output string
    FrameRate int
    ConstantRateFactor int
}

type CommandStop struct {
    Command string
    Session guuid.UUID
}

func server_init(c *cli.Context) error {
    addr := fmt.Sprintf("%s:%d", c.String("host"), c.Int("port"))
    server_log("starting server on " + addr)
    listener, err := net.Listen("tcp4", addr)
    if err != nil {
        server_log("error: " + err.Error())
        return err
    }
    defer listener.Close()
    server_log("listening...")
    
    for {
        conn, err := listener.Accept()
        if err != nil {
            server_log("error: " + err.Error())
            return err
        }
        go server_handle_connection(conn)
    }
}

func server_handle_connection(conn net.Conn) {
    server_log("client connected " + conn.RemoteAddr().String())    
    for {
        data, err := bufio.NewReader(conn).ReadString('\n')
        if err != nil {
            if strings.Contains(err.Error(), "EOF") || strings.Contains(err.Error(), "use of closed network connection") {
                server_log("client disconnected " + conn.RemoteAddr().String())    
            } else {
                server_log("error: " + err.Error())
            }
            return
        }

        var cmd Command
        json.Unmarshal([]byte(data), &cmd)
        server_process_command(cmd, conn, data)
    }
    conn.Close()
}

func server_process_command(cmd Command, conn net.Conn, data string) {
    switch (cmd.Command) {
        case "start":
            server_command_start(conn, data)
        case "stop":
            server_command_stop(conn, data)
        default:
            server_log("Invalid command: " + cmd.Command)
    }
}

func server_command_start(conn net.Conn, data string) {
    var startCmd CommandStart
    json.Unmarshal([]byte(data), &startCmd)
    sessionId := guuid.New()
    session_start(sessionId, conn)
    session_log(sessionId, "session created host=" + startCmd.Addr + ":" + strconv.Itoa(startCmd.Port) + " pass=" + startCmd.Password + " output=" + startCmd.Output)
    go recorder_proc(conn, startCmd, sessionId)
}

func server_command_stop(conn net.Conn, data string) {
    var stopCmd CommandStop
    json.Unmarshal([]byte(data), &stopCmd)
    session_stop(stopCmd.Session, conn)
}

func server_send(sessionId guuid.UUID, status string) {
    conn := session_get_conn(sessionId);
    conn.Write([]byte("{\"status\": \"" + status + "\",\"session\":\"" + sessionId.String() + "\"}\n"))
}

func server_log(message string) {
    fmt.Println("server   | " + message)
}