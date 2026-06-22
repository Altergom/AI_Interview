package wsx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const defaultHandshakeTimeout = 10 * time.Second

// DialBearer dials a Gorilla WebSocket endpoint with Authorization: Bearer header.
// It is intended for upstream WS services that require API key based authentication.
func DialBearer(ctx context.Context, url, apiKey string) (*websocket.Conn, error) {
	hdr := http.Header{"Authorization": []string{"Bearer " + apiKey}}
	dialer := websocket.Dialer{HandshakeTimeout: defaultHandshakeTimeout}
	conn, _, err := dialer.DialContext(ctx, url, hdr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// WriteJSON marshals v into a text frame and writes it to conn.
// It keeps error prefix consistent for log/trace correlation across modules.
func WriteJSON(conn *websocket.Conn, v any, prefix string) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("[%s] marshal msg: %w", prefix, err)
	}
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("[%s] write msg: %w", prefix, err)
	}
	return nil
}

