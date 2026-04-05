class BlockchainViewer {
    constructor() {
        // Detecta ambiente: mock local ou produção
        const isLocalhost = window.location.hostname === 'localhost' || 
                           window.location.hostname === '127.0.0.1';
        
        // Se estiver rodando em localhost:8080, usa mock server em :9000
        // Caso contrário, usa o valor do input
        this.nodeURL = isLocalhost && window.location.port === '8080' 
            ? 'http://localhost:9000'
            : document.getElementById('node-url').value;
        
        this.blocks = [];
        this.forks = [];
        this.pollingInterval = 2000;
        
        // Atualiza o input com a URL em uso
        document.getElementById('node-url').value = this.nodeURL;
        
        this.init();
    }

    init() {
        this.bindEvents();
        this.startPolling();
        this.fetchData();
    }

    bindEvents() {
        document.getElementById('node-url').addEventListener('change', (e) => {
            this.nodeURL = e.target.value;
            this.fetchData();
        });

        document.getElementById('refresh-btn').addEventListener('click', () => {
            this.fetchData();
        });

        document.getElementById('close-details').addEventListener('click', () => {
            document.getElementById('block-details').style.display = 'none';
        });
    }

    startPolling() {
        setInterval(() => {
            this.fetchData();
        }, this.pollingInterval);
    }

    async fetchData() {
        await Promise.all([
            this.fetchBlocks(),
            this.fetchMempool(),
            this.checkHealth()
        ]);
    }

    async checkHealth() {
        try {
            const response = await fetch(`${this.nodeURL}/health`);
            const data = await response.json();
            
            const statusEl = document.getElementById('node-status');
            statusEl.textContent = data.status === 'healthy' ? '🟢 Conectado' : '🔴 Desconectado';
            statusEl.className = `status ${data.status === 'healthy' ? 'connected' : 'disconnected'}`;
            
            document.getElementById('block-count').textContent = `Blocos: ${data.blocks || 0}`;
        } catch (error) {
            const statusEl = document.getElementById('node-status');
            statusEl.textContent = '🔴 Desconectado';
            statusEl.className = 'status disconnected';
        }
    }

    async fetchBlocks() {
        try {
            const response = await fetch(`${this.nodeURL}/blocks`);
            this.blocks = await response.json();
            this.renderChain();
            this.renderStats();
        } catch (error) {
            console.error('Error fetching blocks:', error);
        }
    }

    async fetchMempool() {
        try {
            const response = await fetch(`${this.nodeURL}/mempool`);
            const data = await response.json();
            const count = Array.isArray(data) ? data.length : (data.count || 0);
            document.getElementById('mempool-count').textContent = `Mempool: ${count}`;
        } catch (error) {
            console.error('Error fetching mempool:', error);
        }
    }

    renderChain() {
        const container = document.getElementById('chain-container');
        
        if (!this.blocks || this.blocks.length === 0) {
            container.innerHTML = '<div class="loading">Nenhum bloco encontrado</div>';
            return;
        }

        let html = '';
        
        for (let i = this.blocks.length - 1; i >= 0; i--) {
            const block = this.blocks[i];
            const isFork = this.isForkBlock(block);
            const previousHash = block.header?.previous_block_hash || block.previous_hash || '-';
            const nonce = block.header?.nonce ?? block.nonce ?? '-';
            const timestamp = block.header?.timestamp ?? block.timestamp;
            const timestampText = timestamp
                ? new Date(timestamp * 1000).toLocaleString()
                : 'N/A';
            const txCount = block.transaction
                ? 1
                : (Array.isArray(block.transactions) ? block.transactions.length : 0);
            
            html += `
                <div class="block-card ${isFork ? 'fork' : ''}" onclick="viewer.showBlockDetails(${block.index})">
                    <div class="block-header">
                        <span class="block-index">Bloco #${block.index}</span>
                        ${isFork ? '<span style="color: #ff9800;">🔀 Fork</span>' : ''}
                    </div>
                    <div class="block-hash">Hash: ${block.hash}</div>
                    <div class="block-info">
                        <span>Hash Anterior: <strong>${String(previousHash).substring(0, 16)}...</strong></span>
                        <span>Nonce: <strong>${nonce}</strong></span>
                        <span>Timestamp: <strong>${timestampText}</strong></span>
                        <span>Transações: <strong>${txCount}</strong></span>
                    </div>
                </div>
            `;
            
            if (i > 0) {
                html += '<div class="arrow">⬇️</div>';
            }
        }

        container.innerHTML = html;
    }

    isForkBlock(block) {
        if (!this.forks || this.forks.length === 0) return false;
        
        for (const forkChain of this.forks) {
            for (const forkBlock of forkChain) {
                if (forkBlock.index === block.index && forkBlock.hash === block.hash) {
                    return true;
                }
            }
        }
        return false;
    }

    renderStats() {
        if (!this.blocks || this.blocks.length === 0) {
            return;
        }

        const difficulty = 4;
        const workPerBlock = Math.pow(16, difficulty);
        const totalWork = this.blocks.length * workPerBlock;
        
        document.getElementById('difficulty').textContent = difficulty;
        document.getElementById('total-work').textContent = totalWork.toLocaleString();

        const lastBlock = this.blocks[this.blocks.length - 1];
        if (lastBlock) {
            const timestamp = lastBlock.header?.timestamp ?? lastBlock.timestamp;
            document.getElementById('last-block-time').textContent = timestamp
                ? new Date(timestamp * 1000).toLocaleString()
                : '-';
        }
    }

    showBlockDetails(index) {
        const block = this.blocks.find(b => b.index === index);
        if (!block) return;

        const content = document.getElementById('block-details-content');
        
        const previousHash = block.header?.previous_block_hash || block.previous_hash || '-';
        const merkleRoot = block.header?.merkle_root || '-';
        const nonce = block.header?.nonce ?? block.nonce ?? '-';
        const timestamp = block.header?.timestamp ?? block.timestamp;
        const version = block.header?.version ?? '-';

        let txContent = '';
        if (block.transaction) {
            const tx = block.transaction;
            if (tx.tipo_operacao) {
                txContent = `
                    <div class="tx-list">
                        <h3 style="color: #00d4ff; margin-bottom: 10px;">Transação</h3>
                        <div class="tx-item">
                            <span class="tx-type ${tx.tipo_operacao}">${tx.tipo_operacao}</span>
                            <p><strong>ID:</strong> ${tx.id_transacao}</p>
                            <p><strong>De:</strong> ${tx.remetente?.substring(0, 20) || '-' }...</p>
                            <p><strong>Para:</strong> ${tx.destinatario?.substring(0, 20) || '-' }...</p>
                            <p><strong>Propriedade:</strong> ${tx.dados_propriedade?.matricula || '-'}</p>
                            <p><strong>Descrição:</strong> ${tx.dados_propriedade?.descricao || '-'}</p>
                            <p><strong>Área:</strong> ${tx.dados_propriedade?.area_m2 || '-'} m²</p>
                            <p><strong>Assinatura:</strong> <small>${tx.assinatura_digital?.substring(0, 30) || '-' }...</small></p>
                        </div>
                    </div>
                `;
            } else {
                txContent = `
                    <div class="tx-list">
                        <h3 style="color: #00d4ff; margin-bottom: 10px;">Transação</h3>
                        <div class="tx-item">
                            <pre>${JSON.stringify(tx, null, 2)}</pre>
                        </div>
                    </div>
                `;
            }
        }

        content.innerHTML = `
            <div class="detail-row">
                <span class="detail-label">Índice:</span>
                <span class="detail-value">${block.index}</span>
            </div>
            <div class="detail-row">
                <span class="detail-label">Hash:</span>
                <span class="detail-value">${block.hash}</span>
            </div>
            <div class="detail-row">
                <span class="detail-label">Hash Anterior:</span>
                <span class="detail-value">${previousHash}</span>
            </div>
            <div class="detail-row">
                <span class="detail-label">Merkle Root:</span>
                <span class="detail-value">${merkleRoot}</span>
            </div>
            <div class="detail-row">
                <span class="detail-label">Nonce:</span>
                <span class="detail-value">${nonce}</span>
            </div>
            <div class="detail-row">
                <span class="detail-label">Timestamp:</span>
                <span class="detail-value">${timestamp ? new Date(timestamp * 1000).toLocaleString() : 'N/A'}</span>
            </div>
            <div class="detail-row">
                <span class="detail-label">Versão:</span>
                <span class="detail-value">${version}</span>
            </div>
            ${txContent}
        `;

        document.getElementById('block-details').style.display = 'block';
    }
}

const viewer = new BlockchainViewer();
