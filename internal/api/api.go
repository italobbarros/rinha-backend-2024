// api.go
package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
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
	for _, ch := range clientes.MapInsert {
		close(ch)
	}
	return &Api{
		Clientes: clientes,
		db:       db,
	}
}

// cadastrarTransacao cadastra uma nova transação para um cliente específico.
//
// @Summary Cadastra uma transação
// @Description Cadastra uma nova transação associada a um cliente pelo ID.
// @ID cadastrar-transacao
// @Tags Transacoes
// @Produce json
// @Param id path int true "ID do Cliente" Format(int64)
// @Param transacao body models.PostTransacaoRequest true "Detalhes da Transação"
// @Success 200 {object} models.PostTransacaoResponseSuccess
// @Failure 400 {object} models.PostTransacaoResponseBadRequest
// @Failure 404 {object} models.PostTransacaoResponseNotFound
// @Router /clientes/{id}/transacoes [post]
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
	ready := func() bool {
		select {
		case <-chanClient:
			return true
		case <-time.After(2 * time.Second):
			return false
		}
	}()
	if !ready {
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "Tempo na requisicao passou mais do que eu gostaria"})
		return
	}
	a.Clientes.MapInsert[clienteID] = make(chan struct{})
	defer func() {
		newChanClient, _ := a.Clientes.MapInsert[clienteID]
		close(newChanClient)
	}()
	var newSaldo int64
	if transacao.Tipo == "d" {
		newSaldoIsValid := ClientResult["limite"] + ClientResult["saldo"] - transacao.Valor
		if newSaldoIsValid > 0 {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Erro ao iniciar a transação - saldo inconsistente"})
			return
		}
		newSaldo = ClientResult["saldo"] - transacao.Valor
	} else {
		newSaldo = ClientResult["saldo"] + transacao.Valor
	}

	a.Clientes.Update(clienteID, ClientResult["limite"], newSaldo)

	tx, err := a.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao iniciar a transação: %s", err.Error())})
		return
	}

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
	`, newSaldo, clienteID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar o saldo"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao commitar a transação"})
		return
	}
	ClientResult["saldo"] = newSaldo
	c.JSON(http.StatusOK, gin.H{
		"limite": ClientResult["limite"],
		"saldo":  ClientResult["saldo"],
	})
}

func (a *Api) getExtrato(c *gin.Context) {
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

	chanClient, _ := a.Clientes.MapInsert[clienteID]
	ready := func() bool {
		select {
		case <-chanClient:
			return true
		case <-time.After(2 * time.Second):
			return false
		}
	}()
	if !ready {
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "Tempo na requisicao passou mais do que eu gostaria"})
		return
	}
	rows, err := a.db.Query("SELECT valor,tipo,descricao,data_transacao FROM historico_transacoes WHERE id_cliente = $1 ORDER BY id DESC LIMIT 10 FOR UPDATE;", clienteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	defer rows.Close()

	// Itere sobre as linhas e armazene as transações em uma slice
	var histTransacoes []models.GetHistTransacaoSuccess
	for rows.Next() {
		var t models.GetHistTransacaoSuccess
		if err := rows.Scan(&t.Valor, &t.Tipo, &t.Descricao, &t.DataTransacao); err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		histTransacoes = append(histTransacoes, t)
	}
	log.Println(histTransacoes)
	// Verifique se houve algum erro durante o processamento das linhas
	if err := rows.Err(); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"saldo": gin.H{
			"total":        ClientResult["saldo"],
			"data_extrato": time.Now(),
			"limite":       ClientResult["limite"],
		},
		"ultimas_transacoes": histTransacoes})
}

func (a *Api) Run() {
	router := gin.Default()
	router.Use(corsHandler) // Adicionar o middleware CORS

	router.GET("/clientes/:id/extrato", a.getExtrato)
	router.POST("/clientes/:id/transacoes", a.cadastrarTransacao)

	router.Run(os.Getenv("API_SERVER_LISTEN")) //os.Getenv("API_SERVER_LISTEN")
}

func corsHandler(c *gin.Context) {
	// Configurar cabeçalhos CORS
	c.Header("Access-Control-Allow-Origin", "*") // Permitir qualquer origem
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
	c.Header("Content-Type", "application/json")

	if c.Request.Method == http.MethodOptions {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
}
