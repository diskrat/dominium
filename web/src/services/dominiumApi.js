// src/services/dominiumApi.js
const API_BASE_URL = "http://localhost:8085";
export const executeChaosMint = async (count = 10) => {
    console.log(`Iniciando Chaos Mint com ${count} transações...`);
    let successCount = 0;

    for (let i = 0; i < count; i++) {
        try {
            const response = await fetch(`${API_BASE_URL}/transactions`, {
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

            if (response.ok) {
                successCount++;
            } else {
                console.error(`Erro na transação ${i + 1}`);
            }
        } catch (error) {
            console.error("Falha ao comunicar com a API Gateway:", error);
        }
    }

    return successCount; // Retorna quantas deram certo para a UI saber
};
