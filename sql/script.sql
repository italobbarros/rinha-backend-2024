
-- Criação da tabela de clientes
CREATE TABLE clientes (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255),
    limite INT,
    saldo DECIMAL(20, 2)
);


-- Criação da tabela de histórico de transações
CREATE TABLE historico_transacoes (
    id_transacao SERIAL PRIMARY KEY,
    id_cliente INT REFERENCES clientes(id),
    valor DECIMAL(10, 2),
    tipo CHAR(1), -- 'c' para crédito, 'd' para débito
    descricao VARCHAR(10),
    data_transacao TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


-- Inserção de dados iniciais na tabela de clientes
INSERT INTO clientes (name, limite, saldo) VALUES
    ('Italo', 100000, 0),
    ('Carla', 80000, 0),
    ('Monica', 1000000, 0),
    ('Haroldo', 10000000, 0),
    ('Mariana', 500000, 0);
