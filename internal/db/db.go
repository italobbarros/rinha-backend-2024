package db

import (
	"database/sql"
	"time"

	"github.com/italobbarros/rinha-backend-2024/internal/models"
)

func GetClientes(db *sql.DB) ([]models.Cliente, error) {
	var clientes []models.Cliente

	rows, err := db.Query("SELECT id,saldo,limite FROM clientes;")
	if err != nil {
		return clientes, err
	}
	defer rows.Close()

	// Itere sobre as linhas e armazene as transações em uma slice
	for rows.Next() {
		var t models.Cliente
		if err := rows.Scan(&t.ID, &t.Saldo, &t.Limite); err != nil {
			return clientes, err
		}
		clientes = append(clientes, t)
	}
	if err := rows.Err(); err != nil {
		return clientes, err
	}
	return clientes, nil
}

func UpdateTransationClient(tx *sql.Tx, clientId int, transacao *models.PostTransacaoRequest, newSaldo int64, clientDb *models.Cliente) error {

	_, err := tx.Exec(`
		INSERT INTO historico_transacoes (id_cliente, valor, tipo, descricao, data_transacao)
		VALUES ($1, $2, $3, $4, $5)
	`, clientId, transacao.Valor, transacao.Tipo, transacao.Descricao, time.Now())
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		UPDATE clientes
		SET saldo = $1
		WHERE id = $2
	`, newSaldo, clientId)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func GetValueClient(db *sql.DB, clientId int) (*sql.Tx, models.Cliente, error) {
	var cliente models.Cliente

	tx, err := db.Begin()
	if err != nil {
		return tx, cliente, err
	}

	// Execute a consulta dentro da transação com FOR UPDATE
	rows, err := tx.Query("SELECT id, saldo, limite FROM clientes WHERE id = $1 FOR UPDATE", clientId)
	if err != nil {
		return tx, cliente, err
	}
	defer rows.Close()

	// Verifique se há pelo menos uma linha
	if rows.Next() {
		// Escaneie os valores na struct Cliente
		if err := rows.Scan(&cliente.ID, &cliente.Saldo, &cliente.Limite); err != nil {
			return tx, cliente, err
		}
	} else {
		// Se não houver linhas, cliente não encontrado
		return tx, cliente, sql.ErrNoRows
	}

	// Verifique se houve algum erro durante o escaneamento
	if err := rows.Err(); err != nil {
		return tx, cliente, err
	}

	return tx, cliente, nil
}

func GetValueAndHist(db *sql.DB, clienteID int) (models.GetExtratoHistResponseSuccess, error) {
	// Inicia a transação
	var response models.GetExtratoHistResponseSuccess
	tx, err := db.Begin()
	if err != nil {
		// Lida com o erro ao iniciar a transação
		return response, err
	}
	defer tx.Rollback()

	// Consulta 1: Atualiza o saldo e limite na tabela clientes
	rows, err := tx.Query("SELECT saldo, limite FROM clientes WHERE id = $1 FOR UPDATE;", clienteID)
	if err != nil {
		// Lida com o erro da consulta
		return response, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&response.Saldo.Saldo, &response.Saldo.Limite); err != nil {
			// Lida com o erro de leitura das colunas
			return response, err
		}
	}

	// Consulta 2: Obtém as últimas transações para o cliente
	rows, err = tx.Query("SELECT valor, tipo, descricao, data_transacao FROM historico_transacoes WHERE id_cliente = $1 ORDER BY id DESC LIMIT 10 FOR UPDATE;", clienteID)
	if err != nil {
		// Lida com o erro da segunda consulta
		return response, err
	}
	defer rows.Close()

	// Processa os resultados da segunda consulta (últimas transações)
	for rows.Next() {
		var t models.GetHistTransacao
		if err := rows.Scan(&t.Valor, &t.Tipo, &t.Descricao, &t.DataTransacao); err != nil {
			// Lida com o erro de leitura das colunas
			return response, err
		}
		response.UltimasTransacoes = append(response.UltimasTransacoes, t)
	}

	// Commita a transação se tudo ocorreu sem erros
	if err := tx.Commit(); err != nil {
		// Lida com o erro ao commitar a transação
		return response, err
	}

	// Resto do código para processar os resultados
	return response, nil
}
