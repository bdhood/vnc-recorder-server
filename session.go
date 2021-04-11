package main

import (
	"sync"
	"fmt"
	"net"

	guuid "github.com/google/uuid"
)

var session_map SessionMap = SessionMap { enabled: make(map[guuid.UUID]bool), conn: make(map[guuid.UUID]net.Conn) }

type SessionMap struct {
	mutex sync.Mutex
	enabled map[guuid.UUID]bool
	conn map[guuid.UUID]net.Conn
}

func session_start(sessionId guuid.UUID, conn net.Conn) {
	session_map.mutex.Lock()
	session_map.enabled[sessionId] = true
	session_map.conn[sessionId] = conn
	session_log(sessionId, "starting...")
	session_map.mutex.Unlock()
}

func session_stop(sessionId guuid.UUID, conn net.Conn) {
	session_map.mutex.Lock()
	session_map.enabled[sessionId] = false
	session_map.conn[sessionId] = conn
	session_log(sessionId, "stopping...")
	session_map.mutex.Unlock()
}

func session_remove(sessionId guuid.UUID) {
	session_map.mutex.Lock()
	delete(session_map.enabled, sessionId)
	session_map.conn[sessionId].Close();
	delete(session_map.conn, sessionId)
	session_map.mutex.Unlock()
}

func session_log(sessionId guuid.UUID, text string) {
	trimmedSession := sessionId.String()[len(sessionId.String())-8:]
	fmt.Println(trimmedSession + " | " + text)
}

func session_get_enabled(sessionId guuid.UUID) bool {
	return session_map.enabled[sessionId]
}

func session_get_conn(sessionId guuid.UUID) net.Conn {
	return session_map.conn[sessionId]
}
