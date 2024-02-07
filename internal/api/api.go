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
		MapInsert: map[int]chan struct{}{
			1: make(chan struct{}),
			2: make(chan struct{}),
			3: make(chan struct{}),
			4: make(chan struct{}),
			5: make(chan struct{}),
		},
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
	ClientResult, err := a.Clientes.Get(clienteID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ID do cliente não existe"})
		return
	}
	var transacao models.PostTransacaoRequest
	if err := c.BindJSON(&transacao); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if transacao.Tipo != "c" && transacao.Tipo != "d" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tipo da transação diferente de \"c\" e \"d\""})
		return
	}
	length := len(transacao.Descricao)
	if length < 1 || length > 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "descrição possui tamanho menor do que 1 ou maior do que 10"})
		return
	}

	chanClient, _ := a.Clientes.MapInsert[clienteID]
	<-chanClient
	chanClient <- struct{}{}
	var newSaldo int64
	if transacao.Tipo == "d" {
		newSaldo = ClientResult["limite"] - ClientResult["saldo"] - transacao.Valor
		if newSaldo < 0 {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Erro ao iniciar a transação - saldo inconsistente"})
			return
		}
		ClientResult["saldo"] -= newSaldo
	} else {
		newSaldo = ClientResult["saldo"] + transacao.Valor
	}
	a.Clientes.Update(clienteID, ClientResult["limite"], newSaldo)

	tx, err := a.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao iniciar a transação"})
		return
	}
	defer func() {
		if err := recover(); err != nil {
			tx.Rollback()
		}
		close(chanClient)
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

	_, err = tx.Exec(`
		UPDATE clientes
		SET saldo = $1
		WHERE id = $2;
	`, ClientResult["saldo"], clienteID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar o saldo"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao commitar a transação"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"limite": ClientResult["limite"],
		"saldo":  ClientResult["saldo"],
	})
}

func (a *Api) Run() {
	router := gin.Default()
	router.POST("/clientes/:id/transacoes", a.cadastrarTransacao)
	router.Run(":8080")
}
