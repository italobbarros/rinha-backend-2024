package models

import "time"

type PostTransacaoResponseSuccess struct {
	Limite int   `json:"limite"`
	Saldo  int64 `json:"saldo"`
}

type GetHistTransacao struct {
	Valor         int64     `json:"valor"`
	Tipo          string    `json:"tipo"`
	Descricao     string    `json:"descricao"`
	DataTransacao time.Time `json:"realizada_em"`
}

type GetTransacao struct {
	Limite int64     `json:"limite"`
	Saldo  int64     `json:"saldo"`
	Data   time.Time `json:"data_extrato"`
}

type GetExtratoHistResponseSuccess struct {
	Saldo             GetTransacao
	UltimasTransacoes []GetHistTransacao `json:"ultimas_transacoes"`
}

type PostTransacaoResponseNotFound struct {
	Detail string `json:"detail" example:"ID do cliente não existe"`
}

type PostTransacaoResponseBadRequest struct {
	Detail string `json:"detail" example:"Algum outro erro.."`
}
