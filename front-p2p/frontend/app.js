class BlockchainViewer {
    constructor() {
        this.nodeURL = document.getElementById('node-url').value;
        this.blocks = [];
        this.forks = [];
        this.stats = {};
        this.pollingInterval = 2000;
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

        document.getElementById('set-difficulty-btn').addEventListener('click', () => {
            this.updateDifficulty();
        });

        document.getElementById('difficulty-input').addEventListener('keydown', (e) => {
            if (e.key === 'Enter') {
                this.updateDifficulty();
            }
        });

        document.getElementById('close-details').addEventListener('click', () => {
            document.getElementById('block-details').style.display = 'none';
        });
    }

    startPolling() {
        setInterval(() => this.fetchData(), this.pollingInterval);
    }

    async fetchData() {
        await Promise.all([
            this.fetchBlocks(),
            this.fetchForks(),
            this.fetchStats(),
            this.fetchMempool(),
            this.checkHealth()
        ]);
        this.renderForks();
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

    async fetchForks() {
        try {
            const response = await fetch(`${this.nodeURL}/forks`);
            this.forks = await response.json();
        } catch (error) {
            this.forks = [];
        }
    }

    async fetchStats() {
        try {
            const response = await fetch(`${this.nodeURL}/stats`);
            this.stats = await response.json();
        } catch (error) {
            this.stats = {};
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
            const previousHash = block.previous_hash || '-';
            const nonce = block.nonce ?? '-';
            const timestamp = block.timestamp;
            const timestampText = timestamp ? new Date(timestamp * 1000).toLocaleString() : 'N/A';
            const txCount = block.transaction ? 1 : (Array.isArray(block.transactions) ? block.transactions.length : 0);

            html += `
                <div class="block-card" onclick="viewer.showBlockDetailsByHash('${block.hash}')">
                    <div class="block-header">
                        <span class="block-index">Bloco #${block.index}</span>
                        <span style="color: #00d4ff;">⛏️ PoW</span>
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

    renderForks() {
        const container = document.getElementById('forks-container');
        if (!Array.isArray(this.forks) || this.forks.length === 0) {
            container.innerHTML = '<div class="empty">Nenhum fork detectado</div>';
            return;
        }

        const html = this.forks.map((fork, idx) => {
            const first = fork[0];
            const last = fork[fork.length - 1];
            return `
                <div class="fork-card">
                    <h3>Fork #${idx + 1}</h3>
                    <p><strong>Tamanho:</strong> ${fork.length} bloco(s)</p>
                    <p><strong>Início:</strong> #${first?.index ?? '-'}</p>
                    <p><strong>Ponta:</strong> #${last?.index ?? '-'}</p>
                    <p><strong>Hash ponta:</strong> ${(last?.hash || '-').substring(0, 24)}...</p>
                </div>
            `;
        }).join('');

        container.innerHTML = html;
    }

    renderStats() {
        if (!this.blocks || this.blocks.length === 0) {
            return;
        }

        const difficulty = this.stats?.difficulty ?? this.blocks[this.blocks.length - 1]?.difficulty ?? 0;
        const totalWork = this.stats?.total_work ?? 0;
        document.getElementById('difficulty').textContent = difficulty;
        document.getElementById('total-work').textContent = Number(totalWork).toLocaleString();

        const input = document.getElementById('difficulty-input');
        if (input) {
            input.value = difficulty;
        }

        const lastBlock = this.blocks[this.blocks.length - 1];
        if (lastBlock) {
            const timestamp = lastBlock.timestamp;
            document.getElementById('last-block-time').textContent = timestamp
                ? new Date(timestamp * 1000).toLocaleString()
                : '-';
        }
    }

    async updateDifficulty() {
        const input = document.getElementById('difficulty-input');
        const feedback = document.getElementById('difficulty-feedback');
        if (!input || !feedback) {
            return;
        }

        const value = Number(input.value);
        if (!Number.isInteger(value) || value < 0 || value > 8) {
            feedback.textContent = 'Valor inválido (use 0 a 8).';
            feedback.className = 'difficulty-feedback error';
            return;
        }

        feedback.textContent = 'Aplicando...';
        feedback.className = 'difficulty-feedback pending';

        try {
            const response = await fetch(`${this.nodeURL}/difficulty?value=${value}`, { method: 'POST' });
            if (!response.ok) {
                const msg = await response.text();
                throw new Error(msg || 'falha ao atualizar dificuldade');
            }

            const result = await response.json();
            feedback.textContent = `Dificuldade atualizada para ${result.difficulty}.`;
            feedback.className = 'difficulty-feedback success';
            await this.fetchData();
        } catch (error) {
            feedback.textContent = `Erro ao atualizar dificuldade: ${error.message}`;
            feedback.className = 'difficulty-feedback error';
        }
    }

    showBlockDetailsByHash(hash) {
        const block = this.findBlockByHash(hash);
        if (!block) return;

        const content = document.getElementById('block-details-content');
        const previousHash = block.previous_hash || '-';
        const nonce = block.nonce ?? '-';
        const timestamp = block.timestamp;

        let txContent = '';
        if (block.transaction) {
            txContent = `
                <div class="tx-list">
                    <h3 style="color: #00d4ff; margin-bottom: 10px;">Transação</h3>
                    <div class="tx-item">
                        <pre>${JSON.stringify(block.transaction, null, 2)}</pre>
                    </div>
                </div>
            `;
        }

        content.innerHTML = `
            <div class="detail-row"><span class="detail-label">Índice:</span><span class="detail-value">${block.index}</span></div>
            <div class="detail-row"><span class="detail-label">Hash:</span><span class="detail-value">${block.hash}</span></div>
            <div class="detail-row"><span class="detail-label">Hash Anterior:</span><span class="detail-value">${previousHash}</span></div>
            <div class="detail-row"><span class="detail-label">Nonce:</span><span class="detail-value">${nonce}</span></div>
            <div class="detail-row"><span class="detail-label">Dificuldade:</span><span class="detail-value">${block.difficulty ?? '-'}</span></div>
            <div class="detail-row"><span class="detail-label">Trabalho Acumulado:</span><span class="detail-value">${block.cumulative_work ?? '-'}</span></div>
            <div class="detail-row"><span class="detail-label">Minerador:</span><span class="detail-value">${block.miner_id ?? '-'}</span></div>
            <div class="detail-row"><span class="detail-label">Timestamp:</span><span class="detail-value">${timestamp ? new Date(timestamp * 1000).toLocaleString() : 'N/A'}</span></div>
            ${txContent}
        `;

        document.getElementById('block-details').style.display = 'block';
    }

    findBlockByHash(hash) {
        const main = this.blocks.find(b => b.hash === hash);
        if (main) return main;
        for (const chain of this.forks || []) {
            const b = chain.find(x => x.hash === hash);
            if (b) return b;
        }
        return null;
    }
}

const viewer = new BlockchainViewer();
