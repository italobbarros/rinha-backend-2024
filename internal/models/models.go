package models

import "time"

type Cliente struct {
	ID     int     `json:"id"`
	Name   string  `json:"name"`
	Limite int     `json:"limite"`
	Saldo  float64 `json:"saldo"`
}

// Transacao representa a estrutura da tabela historico_transacoes
type Transacao struct {
	ID            int       `json:"id_transacao"`
	IDCliente     int       `json:"id_cliente"`
	Valor         float64   `json:"valor"`
	Tipo          string    `json:"tipo"`
	Descricao     string    `json:"descricao"`
	DataTransacao time.Time `json:"data_transacao"`
}
