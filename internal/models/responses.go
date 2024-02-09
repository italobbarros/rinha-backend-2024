package models

import "time"

type PostTransacaoResponseSuccess struct {
	Limite int   `json:"limite"`
	Saldo  int64 `json:"saldo"`
}

type GetHistTransacaoSuccess struct {
	Valor         float64   `json:"valor"`
	Tipo          string    `json:"tipo"`
	Descricao     string    `json:"descricao"`
	DataTransacao time.Time `json:"realizada_em"`
}
type PostTransacaoResponseNotFound struct {
	Detail string `json:"detail" example:"ID do cliente não existe"`
}

type PostTransacaoResponseBadRequest struct {
	Detail string `json:"detail" example:"Algum outro erro.."`
}
