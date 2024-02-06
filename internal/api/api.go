// api.go
package api

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/italobbarros/rinha-backend-2024/internal/models"
	_ "github.com/lib/pq"
)

func NewApi(db *sql.DB) *Api {
	clientes := &Clientes{
		Map: map[int]map[string]int64{
			1: {
				"limite": 100000,
				"saldo":  0,
			},
			2: {
				"limite": 80000,
				"saldo":  0,
			},
			3: {
				"limite": 1000000,
				"saldo":  0,
			},
			4: {
				"limite": 10000000,
				"saldo":  0,
			},
			5: {
				"limite": 500000,
				"saldo":  0,
			},
		},
	}

	return &Api{
		Clientes: clientes,
		db:       db,
	}
}

func (a *Api) cadastrarTransacao(c *gin.Context) {
	clienteIDStr := c.Param("id")
	clienteID, err := strconv.Atoi(clienteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do cliente não é um numero"})
		return
	}

	if clienteID > 5 || clienteID <= 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "ID do cliente não é um numero"})
		return
	}
	var transacao models.PostTransacaoRequest
	if err := c.BindJSON(&transacao); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := a.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao iniciar a transação"})
		return
	}
	defer func() {
		if err := recover(); err != nil {
			tx.Rollback()
		}
	}()

	_, err = tx.Exec(`
		INSERT INTO historico_transacoes (id_cliente, valor, tipo, descricao, data_transacao)
		VALUES ($1, $2, $3, $4, $5)
	`, clienteID, transacao.Valor, transacao.Tipo, transacao.Descricao, time.Now())
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao inserir a transação"})
		return
	}

	var novoSaldo float64
	err = tx.QueryRow(`
		UPDATE clientes
		SET saldo = saldo + $1
		WHERE id = $2
		RETURNING saldo
	`, transacao.Valor, clienteID).Scan(&novoSaldo)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar o saldo"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao commitar a transação"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transação cadastrada com sucesso", "novo_saldo": novoSaldo})
}

func (a *Api) Run() {
	router := gin.Default()
	router.POST("/clientes/:id/transacoes", a.cadastrarTransacao)
	router.Run(":8080")
}
