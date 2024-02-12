package api

import "fmt"

// AddClient adiciona um novo cliente à estrutura
func (c *Clientes) Add(id int, limite int64, saldo int64) {
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
func (c *Clientes) Update(id int, limite int64, saldo int64) {
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

func (c *Clientes) Get(id int) (map[string]int64, error) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	if cliente, exists := c.Map[id]; exists {
		return cliente, nil
	}

	return nil, fmt.Errorf("cliente não encontrado")
}

func (c *Clientes) ObterCanal(id int) chan struct{} {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	canal, ok := c.Sync[id]
	if !ok {
		canal = make(chan struct{})
		close(canal)
		c.Sync[id] = canal
	}

	return canal
}

func (c *Clientes) LiberarCanal(id int) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	delete(c.Sync, id)
}
