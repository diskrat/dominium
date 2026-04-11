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
    const response = await request("/attacks/double-spend", { method: "POST" });
    return response.json();
};

export const executeChaosMint = async (count = 10) => {
    let successCount = 0;

    for (let i = 0; i < count; i++) {
        try {
            await request("/transactions", {
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
            successCount++;
        } catch (error) {
            console.error("Falha ao comunicar com a API Gateway:", error);
        }
    }

    return successCount;
};
