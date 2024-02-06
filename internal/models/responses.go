package models

type PostTransacaoResponse struct {
	Limite int     `json:"limite"`
	Saldo  float64 `json:"saldo"`
}
