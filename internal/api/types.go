package api

import (
	"database/sql"
	"sync"
)

type Api struct {
	Clientes *Clientes
	db       *sql.DB
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
