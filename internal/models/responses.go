package models

type PostTransacaoResponseSuccess struct {
	Limite int   `json:"limite"`
	Saldo  int64 `json:"saldo"`
}

type PostTransacaoResponseNotFound struct {
	Detail string `json:"detail" example:"ID do cliente não existe"`
}

type PostTransacaoResponseBadRequest struct {
	Detail string `json:"detail" example:"Algum outro erro.."`
}
