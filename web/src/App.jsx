import React, { Suspense, lazy, useState, useEffect } from "react";
import {
    AlertTriangle,
    Blocks,
    LayoutDashboard,
    Loader2,
} from "lucide-react";
import { Toaster, toast } from "sonner";
import { Button } from "./components/ui/button";
import { cn } from "./lib/utils";
import {
    connectNetworkWebSocket,
    executeChaosMint,
    generateWallet,
    triggerDoubleSpend,
    updateDifficulty,
} from "./services/dominiumApi";

const DashboardView = lazy(() => import("./components/DashboardView"));
const BlockExplorerView = lazy(() => import("./components/BlockExplorerView"));
const SimulatorView = lazy(() => import("./components/SimulatorView"));
const AttacksView = lazy(() => import("./components/AttacksView"));

const App = () => {
    const [currentView, setCurrentView] = useState("dashboard");
    const [nodes, setNodes] = useState([]);
    const [newDifficulty, setNewDifficulty] = useState(24);
    const [isUpdatingDiff, setIsUpdatingDiff] = useState(false);

    const [allBlocks, setAllBlocks] = useState([]);
    const [canonicalHashes, setCanonicalHashes] = useState(new Set());
    const [accounts, setAccounts] = useState([]);
    const [mempool, setMempool] = useState([]);
    const [chaosMintCount, setChaosMintCount] = useState(10);
    const [transactionStats, setTransactionStats] = useState({
        canonicalTxCount: 0,
        allBlocksTxCount: 0,
        orphanTxCount: 0,
        discardedTxCount: 0,
        mempoolTxCount: 0,
    });
    const [isMinting, setIsMinting] = useState(false);
    const [attackLogs, setAttackLogs] = useState([]);
    const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

    const handleDifficultyUpdate = async () => {
        setIsUpdatingDiff(true);
        const toastId = toast.loading("Enviando comando para a rede...");
        try {
            await updateDifficulty(newDifficulty);
            toast.success(
                `Dificuldade atualizada para ${newDifficulty} bits. Os proximos blocos usarao esta regra.`,
                { id: toastId },
            );
        } catch (err) {
            toast.error("Erro ao comunicar com o Gateway.", { id: toastId });
        }
        setIsUpdatingDiff(false);
    };

    const handleChaosMintClick = async (count = chaosMintCount) => {
        setIsMinting(true);
        const toastId = toast.loading("Disparando transacoes...");
        const successCount = await executeChaosMint(count);
        if (successCount > 0) {
            toast.success(`${successCount} transacoes enviadas.`, {
                id: toastId,
            });
        } else {
            toast.error("Falha ao enviar transacoes.", { id: toastId });
        }
        setIsMinting(false);
    };

    const handleGenerateWalletClick = async () => {
        const toastId = toast.loading(
            "Gerando chaves criptograficas na Curva Eliptica (P-256)...",
        );

        try {
            const data = await generateWallet();
            toast.success("Carteira ECDSA (P-256) gerada.", {
                id: toastId,
                description:
                    "As chaves foram geradas com sucesso. Copie no painel abaixo.",
            });

            toast(
                <div className="space-y-2 text-xs">
                    <div>
                        <p className="font-semibold text-sky-400">Chave Publica</p>
                        <p className="break-all font-mono text-muted-foreground">
                            {data.public_key}
                        </p>
                    </div>
                    <div>
                        <p className="font-semibold text-red-400">Chave Privada</p>
                        <p className="break-all font-mono text-muted-foreground">
                            {data.private_key}
                        </p>
                    </div>
                </div>,
            );
        } catch (err) {
            toast.error(err.message || "Falha ao gerar carteira.", { id: toastId });
        }
    };

    useEffect(() => {
        const ws = connectNetworkWebSocket();
        ws.onopen = () => console.log("WebSocket connected");

        ws.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                setNodes(data.nodes || []);

                const blocks = data.all_blocks || data.blocks || [];
                const canonical = data.canonical_chain || [];

                setAllBlocks(blocks);

                const cHashes = new Set();
                canonical.forEach((b) => cHashes.add(b.hash || b.Hash));
                setCanonicalHashes(cHashes);

                setAccounts(data.accounts || []);
                setMempool(data.mempool || []);
                setTransactionStats({
                    canonicalTxCount: data.canonical_tx_count || 0,
                    allBlocksTxCount: data.all_blocks_tx_count || 0,
                    orphanTxCount: data.orphan_tx_count || 0,
                    discardedTxCount: data.discarded_tx_count || 0,
                    mempoolTxCount: data.mempool_tx_count || 0,
                });

                // Simulação simples de interceptação de Logs de Ataque
                if (data.mempool && data.mempool.length > 0) {
                    const attackTxs = data.mempool.filter(
                        (tx) => tx.id && tx.id.startsWith("ATTACK_"),
                    );
                    if (attackTxs.length > 0) {
                        setAttackLogs((prev) => {
                            const newLogs = [...prev];
                            attackTxs.forEach((tx) => {
                                if (!newLogs.find((l) => l.txId === tx.id)) {
                                    newLogs.unshift({
                                        time: new Date().toLocaleTimeString(),
                                        txId: tx.id,
                                        nft: tx.nft_id || tx.NftID,
                                        status: "DETECTADO: Conflito na Mempool",
                                    });
                                }
                            });
                            return newLogs.slice(0, 10); // Mantém os últimos 10
                        });
                    }
                }
            } catch (error) {
                console.error("Error parsing WebSocket:", error);
            }
        };

        ws.onerror = (error) => console.error("WebSocket error:", error);
        ws.onclose = () => console.log("WebSocket disconnected");
        return () => ws.close();
    }, []);

    const handleDoubleSpendClick = async () => {
        const toastId = toast.loading("Iniciando ataque de gasto duplo...");
        try {
            const data = await triggerDoubleSpend();
            toast("Ataque enviado.", {
                id: toastId,
                description: `NFT: ${data.nft}. Aguardando validacao dos nos.`,
            });

            setAttackLogs((prev) => [
                {
                    time: new Date().toLocaleTimeString(),
                    txId: `A:${data.tx_a.substring(0, 8)} | B:${data.tx_b.substring(0, 8)}`,
                    nft: data.nft,
                    status: "ATAQUE DISPARADO: Gasto Duplo",
                },
                ...prev,
            ]);
        } catch (err) {
            toast.error(`Falha no ataque: ${err.message}`, { id: toastId });
        }
    };

    const renderContent = () => {
        const loadingFallback = (
            <div className="flex items-center justify-center py-16">
                <Loader2 className="h-8 w-8 animate-spin text-primary" />
            </div>
        );

        switch (currentView) {
            case "explorer":
                return (
                    <Suspense fallback={loadingFallback}>
                        <BlockExplorerView
                            nodes={nodes}
                            allBlocks={allBlocks}
                            canonicalHashes={canonicalHashes}
                        />
                    </Suspense>
                );
            default:
                return (
                    <div className="space-y-4">
                        <Suspense fallback={loadingFallback}>
                            <DashboardView
                                nodes={nodes}
                                mempool={mempool}
                                accounts={accounts}
                                transactionStats={transactionStats}
                            />
                        </Suspense>
                        <div className="grid items-start gap-4 xl:grid-cols-[minmax(0,1.15fr)_minmax(0,0.85fr)]">
                            <Suspense fallback={loadingFallback}>
                                <SimulatorView
                                    newDifficulty={newDifficulty}
                                    setNewDifficulty={setNewDifficulty}
                                    handleDifficultyUpdate={handleDifficultyUpdate}
                                    isUpdatingDiff={isUpdatingDiff}
                                    handleChaosMintClick={handleChaosMintClick}
                                    isMinting={isMinting}
                                    handleGenerateWalletClick={handleGenerateWalletClick}
                                    chaosMintCount={chaosMintCount}
                                    setChaosMintCount={setChaosMintCount}
                                />
                            </Suspense>
                            <Suspense fallback={loadingFallback}>
                                <AttacksView
                                    handleDoubleSpendClick={handleDoubleSpendClick}
                                    attackLogs={attackLogs}
                                />
                            </Suspense>
                        </div>
                    </div>
                );
        }
    };

    const menuItems = [
        {
            key: "dashboard",
            label: "Painel Geral",
            icon: LayoutDashboard,
        },
        {
            key: "explorer",
            label: "Explorador de Blocos",
            icon: Blocks,
        },
    ];

    return (
        <div className="flex min-h-screen bg-background text-foreground">
            <aside className="hidden w-72 flex-shrink-0 border-r border-border bg-card/70 p-4 lg:block">
                <div className="mb-4 flex items-center gap-2 rounded-lg border border-border bg-background/50 p-3">
                    <AlertTriangle className="h-5 w-5 text-primary" />
                    <p className="text-sm font-semibold">Dominium Control</p>
                </div>
                <nav className="space-y-2">
                    {menuItems.map((item) => {
                        const Icon = item.icon;
                        const active = currentView === item.key;
                        return (
                            <button
                                key={item.key}
                                onClick={() => setCurrentView(item.key)}
                                className={cn(
                                    "flex w-full items-center gap-3 rounded-md px-3 py-2 text-sm transition",
                                    active
                                        ? "bg-primary text-primary-foreground"
                                        : "text-muted-foreground hover:bg-secondary hover:text-foreground",
                                )}
                            >
                                <Icon className="h-4 w-4" />
                                {item.label}
                            </button>
                        );
                    })}
                </nav>
            </aside>

            <main className="min-w-0 flex-1">
                <header className="sticky top-0 z-20 flex items-center justify-between border-b border-border bg-background/90 px-4 py-3 backdrop-blur lg:px-6">
                    <h1 className="text-lg font-semibold tracking-tight lg:text-2xl">
                        Dominium Network Visualizer
                    </h1>
                    <Button
                        variant="outline"
                        className="lg:hidden"
                        onClick={() => setMobileMenuOpen((v) => !v)}
                    >
                        Menu
                    </Button>
                </header>

                {mobileMenuOpen && (
                    <div className="border-b border-border bg-card p-3 lg:hidden">
                        <nav className="grid grid-cols-1 gap-2 sm:grid-cols-2">
                            {menuItems.map((item) => {
                                const Icon = item.icon;
                                return (
                                    <Button
                                        key={item.key}
                                        size="sm"
                                        variant={
                                            currentView === item.key
                                                ? "default"
                                                : "secondary"
                                        }
                                        onClick={() => {
                                            setCurrentView(item.key);
                                            setMobileMenuOpen(false);
                                        }}
                                        className="w-full"
                                    >
                                        <Icon className="h-4 w-4" />
                                        <span className="ml-1">{item.label}</span>
                                    </Button>
                                );
                            })}
                        </nav>
                    </div>
                )}

                <section className="mx-auto w-full max-w-[1560px] p-3 lg:p-6">
                    {renderContent()}
                </section>
            </main>
            <Toaster
                theme="dark"
                richColors
                closeButton
                position="top-right"
                toastOptions={{
                    className: "border-border bg-card text-card-foreground",
                }}
            />
        </div>
    );
};

export default App;
