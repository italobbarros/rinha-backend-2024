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
	postgres "github.com/italobbarros/rinha-backend-2024/internal/db"
	"github.com/italobbarros/rinha-backend-2024/internal/models"
	_ "github.com/lib/pq"
)

func NewApi(db *sql.DB) *Api {
	c, _ := postgres.GetClientes(db)
	log.Println(c)
	clientes := &Clientes{
		Sync: make(map[int]chan struct{}),
		Map: map[int]map[string]int64{
			1: {
				"limite": c[0].Limite,
				"saldo":  c[0].Saldo,
			},
			2: {
				"limite": c[1].Limite,
				"saldo":  c[1].Saldo,
			},
			3: {
				"limite": c[2].Limite,
				"saldo":  c[2].Saldo,
			},
			4: {
				"limite": c[3].Limite,
				"saldo":  c[3].Saldo,
			},
			5: {
				"limite": c[4].Limite,
				"saldo":  c[4].Saldo,
			},
		},
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
	_, err = a.Clientes.Get(clienteID)
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
	ready := func() bool {
		select {
		case <-a.Clientes.ObterCanal(clienteID):
			return true
		case <-time.After(10 * time.Second):
			return false
		}
	}()
	if !ready {
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "Tempo na requisicao passou mais do que eu gostaria"})
		return
	}
	defer func() {
		a.Clientes.LiberarCanal(clienteID)
	}()
	tx, clientDb, err := postgres.GetValueClient(a.db, clienteID)
	if err != nil {
		defer tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro: %s", err.Error())})
		return
	}
	var newSaldo int64
	if transacao.Tipo == "d" {
		newSaldoIsValid := clientDb.Limite + clientDb.Saldo - transacao.Valor
		if newSaldoIsValid < 0 {
			defer tx.Rollback()
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Erro ao iniciar a transação - saldo inconsistente"})
			return
		}
		newSaldo = clientDb.Saldo - transacao.Valor
	} else {
		newSaldo = clientDb.Saldo + transacao.Valor
	}
	err = postgres.UpdateTransationClient(tx, clienteID, &transacao, newSaldo, &clientDb)
	if err != nil {
		defer tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao exec a transação -" + err.Error()})
		return
	}
	//a.Clientes.Update(clienteID, clientDb.Limite, newSaldo)
	clientDb.Saldo = newSaldo
	c.JSON(http.StatusOK, gin.H{
		"limite": clientDb.Limite,
		"saldo":  clientDb.Saldo,
	})
}

func (a *Api) getExtrato(c *gin.Context) {
	clienteIDStr := c.Param("id")
	clienteID, err := strconv.Atoi(clienteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do cliente não é um numero"})
		return
	}
	_, err = a.Clientes.Get(clienteID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ID do cliente não existe"})
		return
	}

	//chama endpoint GetClientSync
	ready := func() bool {
		select {
		case <-a.Clientes.ObterCanal(clienteID):
			return true
		case <-time.After(10 * time.Second):
			return false
		}
	}()
	if !ready {
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "Tempo na requisicao passou mais do que eu gostaria"})
		return
	}
	defer func() {
		a.Clientes.LiberarCanal(clienteID)
	}()
	r, err := postgres.GetValueAndHist(a.db, clienteID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err})
		return
	}
	r.Saldo.Data = time.Now()

	c.JSON(http.StatusOK, r)
}

func (a *Api) GetClientSync(c *gin.Context) {
	clienteIDStr := c.Param("id")
	clienteID, err := strconv.Atoi(clienteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do cliente não é um numero"})
		return
	}
	_, err = a.Clientes.Get(clienteID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ID do cliente não existe"})
		return
	}

	ready := func() bool {
		select {
		case <-a.Clientes.ObterCanal(clienteID):
			return true
		case <-time.After(10 * time.Second):
			return false
		}
	}()
	if !ready {
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "Tempo na requisicao passou mais do que eu gostaria"})
		return
	}
	defer func() {
		a.Clientes.LiberarCanal(clienteID)
	}()

	c.JSON(http.StatusOK, gin.H{"detail": "Deu bom!"})
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
