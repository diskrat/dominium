# Dominium

Blockchain para registro descentralizado de ativos e propriedades.

## O Problema: Fraudes e Inconsistência em Registros

Em muitos sistemas tradicionais, o processo de registro de posse e transferência de patrimônio sofre com:

- **Gasto Duplo (Vendas Múltiplas):** Vender o mesmo ativo para duas pessoas diferentes antes que a transferência seja oficializada.
- **Adulteração:** Registros em papel ou em bancos de dados centralizados podem ser facilmente alterados ou forjados.
- **Falta de Transparência:** Dificuldade em rastrear com precisão o histórico de posse de um ativo ao longo do tempo.

## Nossa Solução (A Ideia)

Um **ledger imutável de ativos** operado por uma rede de múltiplos nós. No _Dominium_, a transferência de posse de qualquer bem (que possua um ID único) é tratada como uma transação nativa da rede.

- **Imutabilidade via Criptografia:** Blocos encadeados por hash tornam impossível adulterar o histórico de um ativo sem refazer todo o esforço de _Proof of Work_ da rede.
- **Validação Distribuída:** Nós mineradores validam as transferências, garantindo que o mesmo ativo não seja transferido duas vezes (prevenção do gasto duplo).
- **Prova de Trabalho (PoW):** Utilizada para assegurar que a rede descentralizada chegue a um consenso seguro sobre quem é o titular verídico de cada ativo registrado.

## Funcionalidades e Requisitos do Sistema

Este projeto implementa uma blockchain distribuída baseada em Proof of Work (PoW) com os seguintes requisitos técnicos:

### 1. Estrutura da Blockchain

- Blocos encadeados por hash, contendo:
    - Lista de transações.
    - Hash do bloco anterior.
    - Nonce para o Proof of Work.

### 2. Proof of Work e Consenso

- O hash do bloco deve iniciar com _N_ zeros (com dificuldade ajustável).
- Adoção da cadeia com maior trabalho acumulado em caso de múltiplos forks.
- Suporte nativo à ocorrência de forks (ex: mineração simultânea).

### 3. Rede e Mineradores

- Execução com múltiplos nós (≥ 3).
- Comunicação e disseminação de blocos/transações na rede.
- Os nós deverão consumir a mempool para montar blocos, rodar o Proof of Work e validar a recepção de novos blocos.

### 4. Mempool e API

- Mempool mantendo as transações pendentes.
- API dedicada para a submissão de novas transações.
- Script para criação de transações aleatórias usando seed, de forma a garantir a reprodutibilidade.

### 5. Experimentação e Visualização

- Visualizar o crescimento da blockchain em tempo real, exibindo forks, os vários blocos gerados e as relações entre eles.
- Executar e simular um ataque de gasto duplo (double spending) verificando seu impacto na rede.

---

## Guia de Instalação e Execução

Existem duas formas de rodar a rede Dominium: a **Automática** (recomendada para testes rápidos) e a **Manual** (ideal para acompanhar os logs detalhados de cada nó).

---

## 1. Modo Automático (Docker Compose)

Este comando sobe toda a infraestrutura (Kafka + 3 Nós da Blockchain) de uma única vez, já configurando as portas e identidades.

Na pasta raiz do projeto, execute:

```bash
docker-compose up --build
```

Este comando irá compilar o código Go dentro dos containers e iniciar a rede P2P instantaneamente.

---

## 2. Modo Manual (Passo a Passo)

### A. Configuração do Docker (O Mensageiro Kafka)

O Kafka atua como o protocolo P2P da rede, permitindo que os nós Alpha, Beta e Gamma se comuniquem.

**1. Para baixar e criar o container (primeira vez):**

```bash
docker run -d --name kafka-dominium -p 9092:9092 apache/kafka
```

**2. Se o container já existir e estiver parado:**

```bash
docker start kafka-dominium
```

**3. Verificação:** Digite `docker ps`. O status do `apache/kafka` deve ser `Up`.

---

### B. Executando os 3 Nós da Blockchain

Abra três terminais separados na pasta raiz do projeto (`dominium`) e execute os comandos abaixo para criar a rede distribuída:

**Terminal 1 — Node Alpha (Porta 8080)**

```bash
go run cmd/sim/main.go -port=8080 -id=node-alpha -diff=2
```

**Terminal 2 — Node Beta (Porta 8081)**

```bash
go run cmd/sim/main.go -port=8081 -id=node-beta -diff=2
```

**Terminal 3 — Node Gamma (Porta 8082)**

```bash
go run cmd/sim/main.go -port=8082 -id=node-gamma -diff=2
```

---

## 3. Como Testar e Monitorar

Com os nós rodando, abra o seu navegador no endereço:

**http://localhost:8080**

### O que observar no Dashboard:

- **Sincronização em Tempo Real:** Clique em `+ Generate 1 Tx`. Observe que a transação aparece nas Mempools dos 3 nós quase simultaneamente. Isso demonstra a propagação via Kafka.

- **Consenso e Mineração:** Assim que um nó minera o bloco, ele propaga o resultado. Se o Alpha ganhar a corrida, o Beta e Gamma validarão o bloco dele e limparão suas mempools automaticamente.

- **Resiliência a Ataques:** Clique em `Simulate Attack`. O sistema tentará realizar um Double Spend. Você verá nos logs o nó rejeitando a transação fraudulenta enquanto mantém a integridade do Ledger.

- **Dificuldade Dinâmica:** Altere o slider de dificuldade (0 a 32). Note que a rede levará mais tempo para encontrar o Hash conforme o número de zeros aumenta, simulando o comportamento real de redes como o Bitcoin.

---

## 4. Visualizando os Logs em Colunas

### 1. Usando o comando nativo do Docker (Recomendado para Debug)

Em vez de rodar o comando geral que mistura tudo, você pode abrir 3 abas do seu terminal e rodar um comando para cada nó. Assim, você terá três "colunas" físicas na sua tela:

| Aba   | Nó    | Comando                     |
| ----- | ----- | --------------------------- |
| Aba 1 | Alpha | `docker logs -f node-alpha` |
| Aba 2 | Beta  | `docker logs -f node-beta`  |
| Aba 3 | Gamma | `docker logs -f node-gamma` |

---

### 2. Usando o utilitário `multitail` (Colunas reais no Linux/WSL)

Se você estiver no Linux ou usando WSL no Windows, o utilitário `multitail` é perfeito para isso. Ele divide o seu terminal em colunas ou linhas automaticamente.

**Comando:**

```bash
multitail -s 3 -l "docker logs -f node-alpha" -l "docker logs -f node-beta" -l "docker logs -f node-gamma"
```

**Opções utilizadas:**

- `-s 3` — Divide a tela em 3 colunas verticais.
- `-l` — Executa o comando de log para cada painel.

**Visualização esperada no terminal:**

```
┌─────────────────────┬─────────────────────┬─────────────────────┐
│    node-alpha       │     node-beta        │     node-gamma      │
│─────────────────────│─────────────────────│─────────────────────│
│ [NODE] block #42    │ [NODE] block #42     │ [NODE] block #42    │
│ [MEMPOOL]  a1b2c3...     │ [MEMPOOL]  a1b2c3...      │ [MEMPOOL]  a1b2c3...     │
│ [VALIDATION]  hash valid    │ [VALIDATION]  hash valid     │ [VALIDATION]  hash valid    │
│ [TRANSACTION] nonce=18432  │ [CONSENSUS] leader mining │ [CONSENSUS] leader mining│
│ ...                 │ ...                  │ ...                 │
└─────────────────────┴─────────────────────┴─────────────────────┘
```

> **Dica:** Caso o `multitail` não esteja instalado, execute `sudo apt install multitail` no Linux/WSL.
