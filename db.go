package main

import (
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

const (
	fileType    = "FILEPLAYER" // Используется текстовыми форматами
	magicNumber = 0x125D       // Используется двоичными форматами
	fileVersion = 100          // Используется всеми форматами
	dateFormat  = "2017-03-09" // Эта дата всегда должна использоваться

	maxGoroutines      = 100
	maxSizeOfSmallFile = 1024 * 32
)

type DB struct {
	path  string
	Users map[string]*User

	///////
	forward chan []byte
	join    chan *client
	leave   chan *client
	clients map[*client]bool
}

func newDB(file string) *DB {
	path, err := filepath.Abs(file)
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	db := &DB{
		path:  path,
		Users: make(map[string]*User),

		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
	}

	if err := db.Load(); err != nil {
		log.Fatalln(err)
		return nil
	}
	save := UpdateMainPlayList(db)
	if len(save) > 0 {
		db.Save()
	}

	return db
}

func (db *DB) run() {
	for {
		select {
		case client := <-db.join:
			// joining
			db.clients[client] = true
		case client := <-db.leave:
			// leaving
			delete(db.clients, client)
			close(client.send)
		case msg := <-db.forward:
			// forward message to all clients
			for client := range db.clients {
				select {
				case client.send <- msg:
				// send the message
				default:
					// failed to send
					delete(db.clients, client)
					close(client.send)
				}
			}
		}
	}
}

func (db *DB) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP DB:", err)
		return
	}
	playerName := "User"
	params, _ := url.ParseQuery(req.URL.RawQuery)
	if len(params["name"]) > 0 {
		playerName = params["name"][0]
	}
	usr, ok := db.Users[playerName]

	if ok {
		client := &client{
			user:   usr,
			socket: socket,
			send:   make(chan []byte, messageBufferSize),
			db:     db,
		}
		db.join <- client
		defer func() { db.leave <- client }()
		go client.write()
		client.read()
	}
}

func (db *DB) Save() error {
	var file, err = os.OpenFile(db.path, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	if err = encoder.Encode(magicNumber); err != nil {
		return err
	}
	if err = encoder.Encode(fileVersion); err != nil {
		return err
	}
	return encoder.Encode(db.Users)
}

func (db *DB) Load() error {
	var err error

	// detect if file exists
	_, err = os.Stat(db.path)
	var reader *os.File

	// create file if not exists
	if os.IsNotExist(err) {
		db.Users[defaultUser] = NewUser()
		return db.Save()
	} else if err != nil {
		return err
	} else {
		reader, err = os.Open(db.path)
		if err != nil {
			return err
		}
	}
	defer reader.Close()

	//READ DATA
	decoder := gob.NewDecoder(reader)
	var magic int
	if err := decoder.Decode(&magic); err != nil {
		return err
	}
	if magic != magicNumber {
		return errors.New("cannot read non-file player gob file")
	}
	var version int
	if err := decoder.Decode(&version); err != nil {
		return err
	}
	if version > fileVersion {
		return fmt.Errorf("version %d is too new to read", version)
	}
	if err := decoder.Decode(&db.Users); err != nil {
		return err
	}

	return err
}
