package main

import (
  "net/http"
  "time"

  "github.com/m01i0ng/go-ws/impl"

  "github.com/gorilla/websocket"
  "github.com/kataras/golog"
)

var upgrader = websocket.Upgrader{
  CheckOrigin: func(r *http.Request) bool {
    return true
  },
}

func main() {
  http.HandleFunc("/ws", wsHandler)
  golog.Info("server starting on :8888")
  err := http.ListenAndServe(":8888", nil)
  if err != nil {
    golog.Errorf("err: %s", err)
  }
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
  conn, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
    golog.Errorf("upgrade error: %s", err)
  }

  golog.Infof("new conn: %s", conn.RemoteAddr().String())

  connection := impl.NewConnection(conn)

  go func() {
    for {
      err := connection.Write([]byte("heart"))
      if err != nil {
        return
      }

      time.Sleep(time.Second)
    }
  }()

  for {
    bytes, err := connection.Read()
    if err != nil {
      _ = connection.Close()
      break
    }

    err = connection.Write(bytes)
    if err != nil {
      _ = connection.Close()
      break
    }
  }
}
