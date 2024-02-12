package api

import (
	"database/sql"
	"sync"
)

type Sync struct {
	mutex     sync.Mutex
	semaphore map[int]chan struct{}
}

type Api struct {
	Clientes *Clientes
	db       *sql.DB
	sync     Sync
}

type ClientSync struct {
	Mutex   sync.Mutex
	Channel chan bool
}
type Clientes struct {
	Sync  map[int]chan struct{}
	Map   map[int]map[string]int64
	Mutex sync.Mutex
}
