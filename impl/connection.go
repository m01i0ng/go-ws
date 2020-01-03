package impl

import (
  "errors"
  "sync"

  "github.com/gorilla/websocket"
  "github.com/kataras/golog"
)

type Connection struct {
  conn      *websocket.Conn
  inChan    chan []byte
  outChan   chan []byte
  closeChan chan []byte

  mutex    sync.Mutex
  isClosed bool
}

func NewConnection(conn *websocket.Conn) *Connection {
  c := &Connection{
    conn:      conn,
    inChan:    make(chan []byte, 1000),
    outChan:   make(chan []byte, 1000),
    closeChan: make(chan []byte, 1),
  }

  go c.readLoop()
  go c.writeLoop()

  return c
}

func (c *Connection) Read() (data []byte, err error) {
  select {
  case data = <-c.inChan:
  case <-c.closeChan:
    err = errors.New("connection is closed")
  }

  return
}

func (c *Connection) Write(data []byte) (err error) {
  select {
  case c.outChan <- data:
  case <-c.closeChan:
    err = errors.New("connection is closed")
  }

  return
}

func (c *Connection) Close() (err error) {
  err = c.conn.Close()

  c.mutex.Lock()
  if !c.isClosed {
    close(c.closeChan)
    c.isClosed = true
  }
  c.mutex.Unlock()

  return
}

func (c *Connection) readLoop() {
  for {
    _, bytes, err := c.conn.ReadMessage()
    if err != nil {
      golog.Errorf("read error: %s", err)
      _ = c.Close()
      break
    }

    select {
    case c.inChan <- bytes:
    case <-c.closeChan:
      _ = c.Close()
    }
  }
}

func (c *Connection) writeLoop() {
  for {
    select {
    case data := <-c.outChan:
      err := c.conn.WriteMessage(websocket.TextMessage, data)
      if err != nil {
        golog.Errorf("write error: %s", err)
        _ = c.Close()
        break
      }
    case <-c.closeChan:
      _ = c.Close()
      break
    }
  }
}
