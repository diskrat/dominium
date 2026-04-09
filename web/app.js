class BlockchainViewer {
    constructor() {
        this.nodeURL = this.resolveInitialNodeURL();
        
        this.blocks = [];
        this.forks = [];
        this.pollingInterval = 2000;
        this.currentDifficulty = 4;
        this.stateEntries = 0;
        
        this.syncSelectorWithNodeURL();
        
        this.init();
    }

    normalizeBaseURL(rawURL) {
        const url = String(rawURL || '').trim().replace(/\/+$/, '');

        if (!url) {
            return window.location.origin;
        }

        if (url.startsWith('http://') || url.startsWith('https://')) {
            return url;
        }

        if (url.startsWith('/')) {
            return `${window.location.origin}${url}`;
        }

        return `http://${url}`;
    }

    resolveInitialNodeURL() {
        const selectorValue = document.getElementById('node-url').value;
        return this.normalizeBaseURL(selectorValue || 'http://localhost:8080');
    }

    syncSelectorWithNodeURL() {
        const selector = document.getElementById('node-url');
        const exactOption = Array.from(selector.options).find(option => option.value === this.nodeURL);

        if (exactOption) {
            selector.value = this.nodeURL;
            return;
        }

        const sameOriginOption = Array.from(selector.options).find(option => {
            return this.normalizeBaseURL(option.value) === this.nodeURL;
        });

        if (sameOriginOption) {
            selector.value = sameOriginOption.value;
        }
    }

    buildURL(path) {
        if (!path || path === '/') {
            return this.nodeURL;
        }

        return `${this.nodeURL}${path.startsWith('/') ? path : `/${path}`}`;
    }

    async fetchJSONWithFallback(paths) {
        let lastError = null;

        for (const path of paths) {
            try {
                const response = await fetch(this.buildURL(path));
                if (!response.ok) {
                    continue;
                }

                const data = await response.json();
                return data;
            } catch (error) {
                lastError = error;
            }
        }

        if (lastError) {
            throw lastError;
        }

        throw new Error('No compatible route available');
    }

    getBlockTransactions(block) {
        if (Array.isArray(block.transactions)) {
            return block.transactions;
        }

        if (block.transaction) {
            return [block.transaction];
        }

        return [];
    }

    formatTimestamp(timestamp) {
        if (typeof timestamp === 'number' && Number.isFinite(timestamp) && timestamp > 0) {
            return new Date(timestamp * 1000).toLocaleString();
        }
        return 'N/A';
    }

    escapeHTML(value) {
        return String(value ?? '-')
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#39;');
    }

    init() {
        this.bindEvents();
        this.startPolling();
        this.fetchData();
    }

    bindEvents() {
        document.getElementById('node-url').addEventListener('change', (e) => {
            this.nodeURL = this.normalizeBaseURL(e.target.value);
            this.fetchData();
        });

        document.getElementById('refresh-btn').addEventListener('click', () => {
            this.fetchData();
        });

        const simulateTxBtn = document.getElementById('simulate-tx-btn');
        if (simulateTxBtn) {
            simulateTxBtn.addEventListener('click', async () => {
                await this.triggerSimulation('/simulate-tx');
            });
        }

        const simulateAttackBtn = document.getElementById('simulate-attack-btn');
        if (simulateAttackBtn) {
            simulateAttackBtn.addEventListener('click', async () => {
                await this.triggerSimulation('/simulate-attack');
            });
        }

        const difficultySlider = document.getElementById('difficulty-slider');
        if (difficultySlider) {
            difficultySlider.addEventListener('change', async (e) => {
                const nextValue = Number(e.target.value);
                await this.setDifficulty(nextValue);
            });
        }

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
        await this.fetchBlocks();
        await Promise.all([this.fetchMempool(), this.fetchState(), this.fetchDifficulty(), this.checkHealth()]);
    }

    async checkHealth() {
        try {
            const data = await this.fetchJSONWithFallback(['/health']);
            
            const statusEl = document.getElementById('node-status');
            statusEl.textContent = data.status === 'healthy' ? '🟢 Conectado' : '🔴 Desconectado';
            statusEl.className = `status ${data.status === 'healthy' ? 'connected' : 'disconnected'}`;
            
            document.getElementById('block-count').textContent = `Blocos: ${data.blocks ?? this.blocks.length}`;
        } catch (error) {
            const statusEl = document.getElementById('node-status');
            const connected = this.blocks.length > 0 || this.stateEntries > 0;
            statusEl.textContent = connected ? '🟢 Conectado' : '🔴 Desconectado';
            statusEl.className = `status ${connected ? 'connected' : 'disconnected'}`;
            document.getElementById('block-count').textContent = `Blocos: ${this.blocks.length}`;
        }
    }

    async fetchState() {
        try {
            const data = await this.fetchJSONWithFallback(['/state']);
            this.stateEntries = (data && typeof data === 'object') ? Object.keys(data).length : 0;
        } catch (error) {
            this.stateEntries = 0;
        }
    }

    async fetchDifficulty() {
        try {
            const response = await fetch(this.buildURL('/difficulty'));
            if (!response.ok) {
                return;
            }

            const data = await response.json();
            if (typeof data?.difficulty === 'number') {
                this.currentDifficulty = data.difficulty;
                document.getElementById('difficulty').textContent = this.currentDifficulty;
                const slider = document.getElementById('difficulty-slider');
                const value = document.getElementById('difficulty-value');
                if (slider) {
                    slider.value = String(this.currentDifficulty);
                }
                if (value) {
                    value.textContent = String(this.currentDifficulty);
                }
            }
        } catch (error) {
            // /difficulty GET may not be available on older nodes
        }
    }

    async setDifficulty(value) {
        try {
            const response = await fetch(this.buildURL(`/difficulty?value=${encodeURIComponent(String(value))}`), {
                method: 'POST'
            });

            if (!response.ok) {
                return;
            }

            this.currentDifficulty = value;
            document.getElementById('difficulty').textContent = this.currentDifficulty;
            const valueEl = document.getElementById('difficulty-value');
            if (valueEl) {
                valueEl.textContent = String(this.currentDifficulty);
            }
            this.renderStats();
        } catch (error) {
            console.error('Error setting difficulty:', error);
        }
    }

    async triggerSimulation(path) {
        try {
            await fetch(this.buildURL(path), { method: 'POST' });
            setTimeout(() => {
                this.fetchData();
            }, 300);
        } catch (error) {
            console.error(`Error triggering ${path}:`, error);
        }
    }

    async fetchBlocks() {
        try {
            const data = await this.fetchJSONWithFallback(['/blockchain', '/blocks', '/']);
            if (Array.isArray(data)) {
                this.blocks = data;
            } else if (Array.isArray(data?.blocks)) {
                this.blocks = data.blocks;
            } else {
                this.blocks = [];
            }

            this.renderChain();
            this.renderStats();
        } catch (error) {
            console.error('Error fetching blocks:', error);
            this.blocks = [];
            this.renderChain();
        }
    }

    async fetchMempool() {
        try {
            const data = await this.fetchJSONWithFallback(['/mempool']);
            const count = Array.isArray(data) ? data.length : (data?.count || 0);
            document.getElementById('mempool-count').textContent = `Mempool: ${count}`;
        } catch (error) {
            console.error('Error fetching mempool:', error);
            document.getElementById('mempool-count').textContent = 'Mempool: 0';
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
            const blockIndex = block.index ?? i;
            const isFork = this.isForkBlock(block);
            const previousHash = block.hash_of_previous || block.header?.hash_of_previous || block.header?.previous_block_hash || block.previous_hash || '-';
            const nonce = block.nonce ?? block.header?.nonce ?? '-';
            const timestamp = block.timestamp ?? block.header?.timestamp;
            const timestampText = this.formatTimestamp(timestamp);
            const txCount = this.getBlockTransactions(block).length;
            
            html += `
                <div class="block-card ${isFork ? 'fork' : ''}" onclick="viewer.showBlockDetails(${Number(blockIndex)})">
                    <div class="block-header">
                        <span class="block-index">Bloco #${blockIndex}</span>
                        ${isFork ? '<span style="color: #ff9800;">🔀 Fork</span>' : ''}
                    </div>
                    <div class="block-hash">Hash: ${this.escapeHTML(block.hash || '-')}</div>
                    <div class="block-info">
                        <span>Hash Anterior: <strong>${this.escapeHTML(String(previousHash).substring(0, 16))}...</strong></span>
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

        const difficulty = this.currentDifficulty;
        const workPerBlock = Math.pow(16, difficulty);
        const totalWork = this.blocks.length * workPerBlock;
        
        document.getElementById('difficulty').textContent = difficulty;
        document.getElementById('total-work').textContent = totalWork.toLocaleString();

        const lastBlock = this.blocks[this.blocks.length - 1];
        if (lastBlock) {
            const timestamp = lastBlock.timestamp ?? lastBlock.header?.timestamp;
            document.getElementById('last-block-time').textContent = this.formatTimestamp(timestamp);
        }
    }

    showBlockDetails(index) {
        const block = this.blocks.find(b => Number(b.index) === Number(index));
        if (!block) return;

        const content = document.getElementById('block-details-content');
        
        const previousHash = block.hash_of_previous || block.header?.hash_of_previous || block.header?.previous_block_hash || block.previous_hash || '-';
        const merkleRoot = block.header?.merkle_root || '-';
        const nonce = block.nonce ?? block.header?.nonce ?? '-';
        const timestamp = block.timestamp ?? block.header?.timestamp;
        const version = block.header?.version ?? '-';
        const transactions = this.getBlockTransactions(block);

        let txContent = '';
        if (transactions.length > 0) {
            const txItems = transactions.map((tx) => {
                const txId = tx.txid || tx.TXid || tx.id_transacao || '-';
                const sender = tx.sender || tx.SenderID || tx.remetente || '-';
                const receiver = tx.receiver || tx.ReceiverID || tx.destinatario || '-';
                const asset = tx.asset || tx.AssetID || tx.dados_propriedade?.matricula || '-';
                const action = tx.action || tx.Action || tx.tipo_operacao || 'N/A';
                const actionClass = String(action).toUpperCase();
                const fee = tx.fee ?? tx.Fee ?? '-';

                return `
                    <div class="tx-item">
                        <span class="tx-type ${this.escapeHTML(actionClass)}">${this.escapeHTML(actionClass)}</span>
                        <p><strong>ID:</strong> ${this.escapeHTML(txId)}</p>
                        <p><strong>De:</strong> ${this.escapeHTML(String(sender).substring(0, 20))}...</p>
                        <p><strong>Para:</strong> ${this.escapeHTML(String(receiver).substring(0, 20))}...</p>
                        <p><strong>Ativo:</strong> ${this.escapeHTML(asset)}</p>
                        <p><strong>Taxa:</strong> ${this.escapeHTML(fee)}</p>
                    </div>
                `;
            }).join('');

            txContent = `
                <div class="tx-list">
                    <h3 style="color: #00d4ff; margin-bottom: 10px;">Transações (${transactions.length})</h3>
                    ${txItems}
                </div>
            `;
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
                <span class="detail-value">${this.formatTimestamp(timestamp)}</span>
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
