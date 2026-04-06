/**
 * Dominium Ledger - P2P Network Dashboard
 * Manages real-time data synchronization across multiple blockchain nodes.
 */

// ==========================================
// 1. CONFIGURATION (NETWORK NODES)
// ==========================================

const NETWORK_NODES = [
    { url: "http://localhost:8080", id: "8080", name: "Alpha" },
    { url: "http://localhost:8081", id: "8081", name: "Beta" },
    { url: "http://localhost:8082", id: "8082", name: "Gamma" },
];

function getRandomNode() {
    return NETWORK_NODES[Math.floor(Math.random() * NETWORK_NODES.length)];
}

// ==========================================
// 2. API INTERACTION (CONTROLS)
// ==========================================

async function updateDifficulty(val) {
    document.getElementById("diffVal").innerText = val;
    console.log(`Updating network difficulty to ${val} for all nodes...`);

    for (const node of NETWORK_NODES) {
        try {
            await fetch(`${node.url}/difficulty?value=${val}`, {
                method: "POST",
            });
            console.log(`[${node.name}] Difficulty updated`);
        } catch (e) {
            console.error(`Error updating difficulty on ${node.name}:`, e);
        }
    }
}

async function simulateTx() {
    const node = getRandomNode();
    try {
        console.log(`Sending transaction to ${node.name}...`);
        await fetch(`${node.url}/simulate-tx`, { method: "POST" });
    } catch (e) {
        console.error(`Failed to send tx to ${node.name}, node might be down.`);
    }
}

async function simulateManyTxs(count) {
    console.log(`Generating ${count} transactions across the network...`);
    for (let i = 0; i < count; i++) {
        const node = getRandomNode();
        fetch(`${node.url}/simulate-tx`, { method: "POST" });
    }
}

async function simulateAttack() {
    alert(
        "Double Spend Attack initiated! Monitor all 3 nodes to see the defense in action.",
    );
    const node = getRandomNode();
    try {
        await fetch(`${node.url}/simulate-attack`, { method: "POST" });
    } catch (e) {
        console.error("Error simulating attack:", e);
    }
}

// ==========================================
// 3. DATA FETCHING (NETWORK MONITORING)
// ==========================================

async function fetchNodeMempool(node) {
    try {
        const response = await fetch(`${node.url}/mempool`);
        const txs = await response.json();
        const container = document.getElementById(`mempool-${node.id}`);

        if (!txs || txs.length === 0) {
            container.innerHTML =
                '<div style="color:#666; font-style:italic; text-align:center;">Empty...</div>';
            return;
        }

        container.innerHTML = txs
            .map(
                (tx) => `
            <div class="mempool-tx">
                <div><strong>${tx.AssetID}</strong></div>
                <div class="tx-details">Action: ${tx.Action}</div>
            </div>
        `,
            )
            .join("");
    } catch (e) {
        document.getElementById(`mempool-${node.id}`).innerHTML =
            '<div style="color:#ff5252; text-align:center;">Offline</div>';
    }
}

async function fetchNodeState(node) {
    try {
        const response = await fetch(`${node.url}/state`);
        const state = await response.json();
        const container = document.getElementById(`state-${node.id}`);

        if (!state || Object.keys(state).length === 0) {
            container.innerHTML =
                '<div style="color:#666; font-style:italic; text-align:center;">No records.</div>';
            return;
        }

        container.innerHTML = Object.entries(state)
            .map(
                ([asset, owner]) => `
            <div class="state-item">
                <div style="color:#00e676;"><strong>${asset}</strong></div>
                <div class="owner-details">${owner.substring(0, 8)}...</div>
            </div>
        `,
            )
            .join("");
    } catch (e) {
        document.getElementById(`state-${node.id}`).innerHTML =
            '<div style="color:#ff5252; text-align:center;">Offline</div>';
    }
}

async function fetchNodeBlockchain(node) {
    try {
        const response = await fetch(`${node.url}/blockchain`);
        const blocks = await response.json();
        const container = document.getElementById(`chain-${node.id}`);

        if (!blocks) return;

        // INVERSÃO: O array reverso faz o bloco mais novo renderizar primeiro (na esquerda)
        const reversedBlocks = [...blocks].reverse();

        container.innerHTML = "";
        reversedBlocks.forEach((block, index) => {
            const blockDiv = document.createElement("div");
            blockDiv.className = "block"; // Utiliza a classe do flex-shrink no CSS

            const isGenesis = block.index === 0;

            const txHtml =
                block.transactions && block.transactions.length > 0
                    ? block.transactions
                          .map(
                              (tx) =>
                                  `<div>📄 ${tx.AssetID} <span style="color:#aaa; font-size:0.8em">(${tx.Action})</span></div>`,
                          )
                          .join("")
                    : `<div style="color: #666; font-style: italic;">${isGenesis ? "Genesis Block" : "No Txs"}</div>`;

            // Tratamento do nome do minerador (garantindo que se vier vazio, mostre algo)
            const minerName = block.miner
                ? block.miner
                : isGenesis
                  ? "Genesis System"
                  : "Unknown";

            // O ÚNICO INNERHTML NECESSÁRIO
            blockDiv.innerHTML = `
                <div style="display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid #444; padding-bottom: 8px; margin-bottom: 8px;">
                    <h4 style="margin:0; color: ${isGenesis ? "#fbc02d" : "#00e676"}; font-size: 1.2em;">Block #${block.index}</h4>
                    <span style="font-size: 0.8em; background: #333; padding: 2px 6px; border-radius: 4px; color: #fff;">${minerName}</span>
                </div>
                
                <div class="block-info">
                    <p><strong>Hash:</strong></p>
                    <p class="full-hash">${block.hash}</p>
                    
                    <p><strong>Prev Hash:</strong></p>
                    <p class="full-hash prev">${block.hash_of_previous || "0000000000000000000000000000000000000000000000000000000000000000"}</p>
                </div>
                
                <div style="display: flex; justify-content: space-between; margin: 10px 0; font-size: 0.85em; color: #aaa;">
                    <span><strong>Nonce:</strong> ${block.nonce}</span>
                    <span><strong>Txs:</strong> ${block.transaction_counter}</span>
                </div>

                <div class="tx-list">${txHtml}</div>
            `;

            container.appendChild(blockDiv);

            // Seta ligando ao bloco anterior
            if (index < reversedBlocks.length - 1) {
                const arrow = document.createElement("div");
                arrow.className = "arrow";
                arrow.innerHTML = "◀";
                container.appendChild(arrow);
            }
        });
    } catch (e) {
        document.getElementById(`chain-${node.id}`).innerHTML = "";
    }
}

// ==========================================
// 4. DASHBOARD LIFECYCLE
// ==========================================

function updateDashboard() {
    NETWORK_NODES.forEach((node) => {
        fetchNodeMempool(node);
        fetchNodeState(node);
        fetchNodeBlockchain(node);
    });
}

// Polling update
setInterval(updateDashboard, 2000);
window.onload = updateDashboard;
