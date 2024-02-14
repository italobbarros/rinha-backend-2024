package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/italobbarros/rinha-backend-2024/internal/models"
)

var (
	insertNewTransactionStmt *sql.Stmt
	updateCreditStmt         *sql.Stmt
	updateDebitStmt          *sql.Stmt
	selectExtractStmt        *sql.Stmt
)

func Init(db *sql.DB) {
	// Inicializar prepared statements uma vez ao carregar o pacote
	insertNewTransactionStmt, _ = db.Prepare(`
        INSERT INTO historico_transacoes (id_cliente, valor, tipo, descricao, data_transacao)
        VALUES ($1, $2, $3, $4, $5)
    `)

	updateCreditStmt, _ = db.Prepare(`
	UPDATE clientes
        SET saldo = saldo + $1
        WHERE id = $2 Returning saldo, limite;
    `)

	updateDebitStmt, _ = db.Prepare(`
        UPDATE clientes
        SET saldo = saldo - $1
        WHERE id = $2 Returning saldo, limite;
    `)

	selectExtractStmt, _ = db.Prepare(`
		SELECT c.saldo, c.limite, h.valor, h.tipo, h.descricao, h.data_transacao
		FROM historico_transacoes h
		LEFT JOIN clientes c ON c.id = h.id_cliente
		WHERE c.id = $1
		ORDER BY h.id DESC
		LIMIT 10 for update;
	`)

}

func DatabaseIsConnected(db *sql.DB) bool {
	err := db.Ping()
	return err == nil
}
func Close() {
	insertNewTransactionStmt.Close()
	updateCreditStmt.Close()
	updateDebitStmt.Close()
}

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

func GetClientesById(db *sql.DB, clientId int) (models.GetTransacao, error) {
	var cliente models.GetTransacao

	tx, err := db.Begin()
	if err != nil {
		return cliente, err
	}
	// Consulta SQL com seleção específica de colunas
	query := `SELECT saldo,limite FROM clientes WHERE id = $1;`

	// Executa a consulta SQL
	rows, err := tx.Query(query, clientId)
	if err != nil {
		tx.Rollback()
		return cliente, err
	}
	defer rows.Close()

	// Processa os resultados da consulta
	for rows.Next() {
		if err := rows.Scan(&cliente.Saldo, &cliente.Limite); err != nil {
			tx.Rollback()
			return cliente, err
		}
	}
	// Commita a transação
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return cliente, err
	}
	return cliente, nil
}

func UpdateTransationClient(tx *sql.Tx, clientId int, transacao *models.PostTransacaoRequest, newSaldo int64, clientDb *models.Cliente) error {
	fmt.Println("UpdateTransationClient clientId:", clientId)

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

func UpdateCreditTransationClient(db *sql.DB, clientId int, transacao *models.PostTransacaoRequest) (models.PostTransacaoResponseSuccess, error) {
	var result models.PostTransacaoResponseSuccess

	// Inicie a transação com o contexto
	tx, err := db.Begin()
	if err != nil {
		return result, err
	}

	_, err = tx.Stmt(insertNewTransactionStmt).Exec(clientId, transacao.Valor, transacao.Tipo, transacao.Descricao, time.Now())
	if err != nil {
		tx.Rollback()
		return result, err
	}

	row := tx.Stmt(updateCreditStmt).QueryRow(transacao.Valor, clientId)

	// Escanear os valores retornados
	if err := row.Scan(&result.Saldo, &result.Limite); err != nil {
		tx.Rollback()
		return result, err
	}

	if err := tx.Commit(); err != nil {
		return result, err
	}

	return result, nil
}

func UpdateDebitTransationClient(db *sql.DB, clientId int, transacao *models.PostTransacaoRequest) (models.PostTransacaoResponseSuccess, error) {
	var result models.PostTransacaoResponseSuccess

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel() // Certifique-se de chamar cancel para liberar recursos do contexto

	// Inicie a transação com o contexto
	tx, err := db.Begin()
	if err != nil {
		return result, err
	}

	// Use goroutine para executar a transação e aguardar o contexto ser cancelado
	done := make(chan struct{})
	var success bool = false

	go func() {
		defer close(done) // close the channel when the goroutine exits

		row := tx.StmtContext(ctx, updateDebitStmt).QueryRowContext(ctx, transacao.Valor, clientId)

		// Escaneie os valores retornados
		if err := row.Scan(&result.Saldo, &result.Limite); err != nil {
			fmt.Println("Error scanning row:", err)
			return
		}

		if result.Saldo < result.Limite*-1 {
			return
		}
		_, err := tx.StmtContext(ctx, insertNewTransactionStmt).ExecContext(ctx, clientId, transacao.Valor, transacao.Tipo, transacao.Descricao, time.Now())

		if err != nil {
			fmt.Println("Error executing INSERT:", err)
			return
		}

		success = true
	}()

	// Aguarde até que a transação seja concluída ou o contexto seja cancelado
	select {
	case <-done:
		if !success {
			tx.Rollback()
			return result, errors.New("transaction failed")
		}
		// A transação foi concluída com sucesso
	case <-ctx.Done():
		// O contexto foi cancelado devido ao timeout
		tx.Rollback()
		return result, errors.New("timeout exceeded")
	}

	if err := tx.Commit(); err != nil {
		return result, err
	}
	// Commit ou rollback dependendo do resultado

	return result, nil
}

func GetValueClient(db *sql.DB, clientId int) (*sql.Tx, models.Cliente, error) {
	var cliente models.Cliente
	fmt.Println("GetValueClient clientId:", clientId)
	tx, err := db.Begin()
	if err != nil {
		return tx, cliente, err
	}

	// Execute a consulta dentro da transação com FOR UPDATE
	rows, err := tx.Query("SELECT id, saldo, limite FROM clientes WHERE id = $1 FOR UPDATE;", clientId)
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
	fmt.Println("GetValueAndHist clientId:", clienteID)
	var response models.GetExtratoHistResponseSuccess

	// Inicia a transação
	tx, err := db.Begin()
	if err != nil {
		return response, err
	}
	defer tx.Rollback()

	// Executa a consulta SQL
	rows, err := tx.Stmt(selectExtractStmt).Query(clienteID)
	if err != nil {
		return response, err
	}
	defer rows.Close()

	// Tamanho estimado para o slice
	response.UltimasTransacoes = make([]models.GetHistTransacao, 0, 10)

	// Processa os resultados da consulta
	for rows.Next() {
		var t models.GetHistTransacao
		if err := rows.Scan(&response.Saldo.Saldo, &response.Saldo.Limite, &t.Valor, &t.Tipo, &t.Descricao, &t.DataTransacao); err != nil {
			return response, err
		}
		response.UltimasTransacoes = append(response.UltimasTransacoes, t)
	}
	// Commita a transação
	if err := tx.Commit(); err != nil {
		return response, err
	}

	if response.Saldo.Limite == 0 {
		response.Saldo, err = GetClientesById(db, clienteID)
		if err != nil {
			return response, err
		}
		response.UltimasTransacoes = make([]models.GetHistTransacao, 0)
	}

	return response, nil
}
