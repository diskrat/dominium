const API_BASE_URL =
    (typeof import.meta !== "undefined" && import.meta.env?.VITE_API_URL) ||
    (typeof process !== "undefined" && process.env?.REACT_APP_API_URL) ||
    "http://localhost:8085";

const WS_BASE_URL =
    (typeof import.meta !== "undefined" && import.meta.env?.VITE_WS_URL) ||
    (typeof process !== "undefined" && process.env?.REACT_APP_WS_URL) ||
    "ws://localhost:8080";

const request = async (path, options = {}) => {
    const response = await fetch(`${API_BASE_URL}${path}`, options);
    if (!response.ok) {
        const text = await response.text();
        throw new Error(text || `Falha na chamada ${path}`);
    }
    return response;
};

export const connectNetworkWebSocket = () => new WebSocket(`${WS_BASE_URL}/ws`);

export const updateDifficulty = async (difficulty) => {
    await request("/network/difficulty", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ difficulty }),
    });
};

export const generateWallet = async () => {
    const response = await request("/wallet/generate");
    return response.json();
};

export const triggerDoubleSpend = async () => {
    const response = await fetch(`${API_BASE_URL}/attacks/double-spend`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
    });

    const text = await response.text();
    let body;
    try {
        body = JSON.parse(text);
    } catch (error) {
        body = text;
    }

    return {
        ok: response.ok,
        status: response.status,
        statusText: response.statusText,
        body,
    };
};

export const executeChaosMint = async (count = 50, onSuccess) => {
    let successCount = 0;

    for (let i = 0; i < count; i++) {
        try {
            const resp = await request("/transactions", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({
                    type: "mint",
                    recipient: `carteira_teste_${Math.floor(Math.random() * 1000)}`,
                    nft_id: `nft_caos_${Date.now()}_${i}`,
                }),
            });

            // parse returned transaction id (if any) and notify caller
            let data = null;
            try {
                data = await resp.json();
            } catch (e) {
                data = null;
            }

            successCount++;
            if (typeof onSuccess === "function") {
                const txId = data && (data.tx_id || data.TxID) ? (data.tx_id || data.TxID) : null;
                try {
                    onSuccess(txId);
                } catch (e) {
                    console.error("onSuccess callback error:", e);
                }
            }
        } catch (error) {
            console.error("Falha ao comunicar com a API Gateway:", error);
        }
    }

    return successCount;
};
