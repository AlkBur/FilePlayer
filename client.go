package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
)

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize,
}

type client struct {
	user *User
	// socket is the web socket for this client.
	socket *websocket.Conn
	// send is a channel on which messages are sent.
	send chan []byte
	// room is the room this client is chatting in.
	db *DB
}

type Command struct {
	Name string `json:"name"`
	Args string `json:"args"`
}

func NewCommand(b []byte) *Command {
	cmd := new(Command)
	err := json.Unmarshal(b, &cmd)
	if err != nil {
		logDebug("Error command: '%s'", string(b))
	}
	return cmd
}

func (c *client) read() {
	for {
		if _, msg, err := c.socket.ReadMessage(); err == nil {
			cmd := NewCommand(msg)
			switch cmd.Name {
			case "get_settings":
				usr := c.user
				if usr != nil {
					//logDebug("Send settings: %s", usr)
					c.sendCmd("settings", usr)
				}
			case "get_playlist":
				pl, ok := c.user.PlayLists[cmd.Args]
				if ok {
					fullPl := make([]*fileInfoMP3, 0, len(pl.Files))
					for _, id := range pl.Files {
						fl, ok := mainPlayList.Get(id)
						if ok {
							fullPl = append(fullPl, fl)
						}
					}

					data := make(map[string]interface{})
					data["id"] = pl.ID
					data["data"] = fullPl
					//logDebug("Send playlist: %s", data)
					c.sendCmd("playlist", data)
				}
			case "currentFileID":
				logDebug("Get currentFileID = %s", cmd.Args)
				c.user.File = cmd.Args
				c.db.Save()
			default:
				c.db.forward <- msg
			}
		} else {
			break
		}
	}
	c.socket.Close()
}
func (c *client) write() {
	for msg := range c.send {
		if err := c.socket.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
	c.socket.Close()
}

func (c *client) sendCmd(name string, arg interface{}) {
	cmd := Command{
		Name: name,
	}
	b, err := json.Marshal(arg)
	if err == nil {
		cmd.Args = string(b)
		b, err = json.Marshal(cmd)
		if err == nil {
			c.send <- b
		} else {
			logDebug("Error CMD: %s", err)
		}
	} else {
		logDebug("Error CMD: %s", err)
	}
}

