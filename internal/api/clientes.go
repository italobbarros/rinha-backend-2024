package api

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
