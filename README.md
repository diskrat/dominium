# Dominium
Blockchain para registro descentralizado de ativos e propriedades.

## O Problema: Fraudes e Inconsistência em Registros
Em muitos sistemas tradicionais, o processo de registro de posse e transferência de patrimônio sofre com:
- **Gasto Duplo (Vendas Múltiplas):** Vender o mesmo ativo para duas pessoas diferentes antes que a transferência seja oficializada.
- **Adulteração:** Registros em papel ou em bancos de dados centralizados podem ser facilmente alterados ou forjados.
- **Falta de Transparência:** Dificuldade em rastrear com precisão o histórico de posse de um ativo ao longo do tempo.

## Nossa Solução (A Ideia)
Um **ledger imutável de ativos** operado por uma rede de múltiplos nós. No *Dominium*, a transferência de posse de qualquer bem (que possua um ID único) é tratada como uma transação nativa da rede.

- **Imutabilidade via Criptografia:** Blocos encadeados por hash tornam impossível adulterar o histórico de um ativo sem refazer todo o esforço de *Proof of Work* da rede.
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
- O hash do bloco deve iniciar com *N* zeros (com dificuldade ajustável).
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
