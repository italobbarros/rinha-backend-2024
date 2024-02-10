package models

import "time"

type Cliente struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Limite int64  `json:"limite"`
	Saldo  int64  `json:"saldo"`
}

// Transacao representa a estrutura da tabela historico_transacoes
type HistTransacao struct {
	ID            int       `json:"id_transacao"`
	IDCliente     int       `json:"id_cliente"`
	Valor         int64     `json:"valor"`
	Tipo          string    `json:"tipo"`
	Descricao     string    `json:"descricao"`
	DataTransacao time.Time `json:"realizada_em"`
}
