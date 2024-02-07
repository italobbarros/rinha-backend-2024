package api

import (
	"database/sql"
	"sync"
)

type Api struct {
	Clientes *Clientes
	db       *sql.DB
}

type Clientes struct {
	MapInsert map[int]chan struct{}
	Map       map[int]map[string]int64
	Mutex     sync.Mutex
}
