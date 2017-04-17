package main

import (
	"fmt"
	"github.com/AlkBur/ID3tag"
	"github.com/google/uuid"
	"hash/crc32"
	"log"
	"sort"
	"sync"
)

const (
	defaultPlayList = "Default"
	defaultUser     = "User"
)

var mainPlayList listFilesID3

type listFilesID3 struct {
	mux   sync.RWMutex
	files map[string]*fileInfoMP3
}

type PlayList struct {
	Name string
	ID   string
	//Array ID info MP3
	Files []string
}

type User struct {
	Name        string
	PlayLists   map[string]*PlayList
	CurPlayList string
	File        string
}

type fileInfoMP3 struct {
	ID   string
	size int64
	path string
	//TAG
	Title  string
	Album  string
	Artist string
	Year   string
}

func init() {
	mainPlayList.files = make(map[string]*fileInfoMP3)
}

func NewPlayList(name string, cap int) *PlayList {
	return &PlayList{
		Name:  name,
		ID:    uuid.Must(uuid.NewUUID()).String(),
		Files: make([]string, 0, cap),
	}
}

func NewPlayListDefault() *PlayList {
	mainPlayList.mux.RLock()
	pl := newPlayListfromList(defaultPlayList, mainPlayList.files)
	mainPlayList.mux.RUnlock()
	pl.Sort()
	return pl
}

func newPlayListfromList(name string, l map[string]*fileInfoMP3) *PlayList {
	pl := NewPlayList(name, len(l))
	for key := range l {
		pl.Files = append(pl.Files, key)
	}
	return pl
}

func NewUser() *User {
	usr := &User{
		Name:      defaultUser,
		PlayLists: make(map[string]*PlayList),
	}
	pl := NewPlayListDefault()
	usr.PlayLists[pl.ID] = pl
	usr.CurPlayList = pl.ID
	return usr
}

func (pl *PlayList) Len() int {
	return len(pl.Files)
}

func (l *listFilesID3) Len() int {
	return len(l.files)
}

func (pl *PlayList) Swap(i, j int) {
	pl.Files[i], pl.Files[j] = pl.Files[j], pl.Files[i]
}

func (pl *PlayList) Less(i, j int) bool {
	var fi, fj *fileInfoMP3
	var ok bool

	fi, ok = pl.GetFile(i)
	if !ok {
		return false
	}
	fj, ok = pl.GetFile(j)
	if !ok {
		return false
	}
	return fi.Title < fj.Title
}

func (pl *PlayList) GetFile(id int) (*fileInfoMP3, bool) {
	f, ok := mainPlayList.files[pl.Files[id]]
	return f, ok
}

func (l *listFilesID3) Get(id string) (*fileInfoMP3, bool) {
	l.mux.RLock()
	f, ok := l.files[id]
	l.mux.RUnlock()
	return f, ok
}

func (pl *PlayList) AddFile(id string) {
	pl.Files = append(pl.Files, id)
}

func (pl *PlayList) DelFile(id string) {
	for i, item := range pl.Files {
		if item == id {
			pl.Files = append(pl.Files[:i], pl.Files[i+1:]...)
			break
		}
	}
}

func (pl *PlayList) Sort() {
	sort.Sort(pl)
}

func UpdateMainPlayList(db *DB) map[*User]bool {
	change := make(map[*User]bool)
	pathData, err := ID3tag.ReadPath(defaultPathMP3)
	if err != nil {
		log.Fatalln(err)
		return change
	}

	files := make(map[string]*fileInfoMP3)
	format := fmt.Sprintf("%%016X-%%%dX", crc32.Size*2) // == "%016X:%40X"

	h := crc32.NewIEEE()
	for _, id3 := range pathData {
		h.Reset()
		h.Write([]byte(id3.FileName()))

		f := &fileInfoMP3{
			ID:   fmt.Sprintf(format, id3.Size(), h.Sum32()),
			size: id3.Size(),
			path: id3.Path(),
			//TAG
			Title:  id3.Title(),
			Album:  id3.Album(),
			Artist: id3.Artist(),
			Year:   id3.Year(),
		}
		files[f.ID] = f
	}

	mainPlayList.mux.Lock()
	mainPlayList.files = files
	mainPlayList.mux.Unlock()

	if db != nil {
		for _, usr := range db.Users {
			for key, pl := range usr.PlayLists {
				if pl.Name == defaultPlayList && pl.Len() == 0 {
					pl = newPlayListfromList(pl.Name, files)
					pl.Sort()
					pl.ID = key
					usr.PlayLists[key] = pl
				} else {
					arrDel := make([]string, 0)
					for _, id := range pl.Files {
						_, ok := mainPlayList.Get(id)
						if !ok {
							arrDel = append(arrDel, id)
							change[usr] = true
						}
					}
					if len(arrDel) == pl.Len() {
						logDebug("Delete all files: %s", pl)
						logDebug("mainPlayList: %s", mainPlayList)
					}
					for _, id := range arrDel {
						pl.DelFile(id)
					}
					if pl.Name == defaultPlayList && pl.Len() == 0 {

						pl = newPlayListfromList(pl.Name, files)
						pl.Sort()
						pl.ID = key
						usr.PlayLists[key] = pl
					}
				}
			}
		}
	}
	return change
}
