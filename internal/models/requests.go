package models

// Transacao representa a estrutura da tabela historico_transacoes
type PostTransacaoRequest struct {
	Valor     int64  `json:"valor"`
	Tipo      string `json:"tipo"`
	Descricao string `json:"descricao"`
}
