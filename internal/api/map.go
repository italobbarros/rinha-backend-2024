package api

import "fmt"

// AddClient adiciona um novo cliente à estrutura
func (c *Clientes) AddClient(id int, limite int64, saldo int64) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	if _, exists := c.Map[id]; exists {
		// Cliente já existe, pode tratar isso como necessário
		// Aqui, não fazemos nada, mas você pode lançar um erro, atualizar os valores existentes, etc.
		return
	}

	c.Map[id] = map[string]int64{
		"limite": limite,
		"saldo":  saldo,
	}
}

// UpdateClient atualiza os valores de limite e saldo para um cliente existente
func (c *Clientes) UpdateClient(id int, limite int64, saldo int64) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	if _, exists := c.Map[id]; !exists {
		// Cliente não existe, pode tratar isso como necessário
		// Aqui, não fazemos nada, mas você pode lançar um erro, adicionar o cliente, etc.
		return
	}

	c.Map[id]["limite"] = limite
	c.Map[id]["saldo"] = saldo
}

func (c *Clientes) GetClient(id int) (map[string]int64, error) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	if cliente, exists := c.Map[id]; exists {
		return cliente, nil
	}

	return nil, fmt.Errorf("Cliente não encontrado")
}
