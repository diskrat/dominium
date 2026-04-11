import React from "react";
import { Cpu, Database } from "lucide-react";
import { Badge } from "./ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";

const DashboardView = ({ nodes, mempool, accounts }) => {
    return (
        <div className="space-y-4">
            <h2 className="text-xl font-semibold tracking-tight lg:text-2xl">
                Painel de Observabilidade - Rede Dominium
            </h2>

            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <Cpu className="h-4 w-4 text-primary" />
                        Status dos Nos
                    </CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="grid grid-cols-[repeat(auto-fit,minmax(210px,1fr))] gap-3">
                        {nodes.map((node) => (
                            <div key={node.id} className="rounded-md border border-border bg-secondary/30 p-3">
                                <p className="font-semibold">{node.id}</p>
                                <div className="mt-2 flex items-center gap-2">
                                    <Badge variant={node.status === "mining" ? "success" : "secondary"}>
                                        {node.status === "mining" ? "Minerando" : "Ocioso"}
                                    </Badge>
                                </div>
                                <p className="mt-2 text-sm text-muted-foreground">
                                    Dificuldade: {node.difficulty} bits
                                </p>
                                <p className="text-sm text-muted-foreground">Altura: {node.blockHeight}</p>
                            </div>
                        ))}
                    </div>
                </CardContent>
            </Card>

            <div className="grid grid-cols-[repeat(auto-fit,minmax(340px,1fr))] gap-4">
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2 text-base">
                            <Database className="h-4 w-4 text-primary" />
                            Mempool - Fila de Espera
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="max-h-64 space-y-2 overflow-y-auto pr-1">
                            {mempool.length === 0 && (
                                <p className="text-sm text-muted-foreground">Mempool vazia</p>
                            )}
                            {mempool.map((tx) => {
                                const id = tx.id || tx.ID || "unknown";
                                const type = tx.type !== undefined ? tx.type : tx.Type;
                                const ts = tx.timestamp || tx.Timestamp;
                                const date = ts ? new Date(Number(ts) / 1000000) : null;
                                const timeLabel =
                                    date && !Number.isNaN(date) ? date.toLocaleTimeString() : "pendente";

                                return (
                                    <div
                                        key={`${id}-${timeLabel}`}
                                        className="flex items-center justify-between rounded-md border border-border px-3 py-2"
                                    >
                                        <p className="font-mono text-xs">{id.substring(0, 18)}...</p>
                                        <div className="flex items-center gap-2">
                                            <Badge variant={type === 2 ? "success" : "default"}>
                                                {type === 2 ? "MINT" : "TRANSFER"}
                                            </Badge>
                                            <span className="text-xs text-muted-foreground">{timeLabel}</span>
                                        </div>
                                    </div>
                                );
                            })}
                        </div>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle className="text-base">Estado das Contas</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="max-h-64 space-y-2 overflow-y-auto pr-1">
                            {accounts.length === 0 && (
                                <p className="text-sm text-muted-foreground">Nenhuma conta recebida.</p>
                            )}
                            {accounts.map((account) => (
                                <div key={account.publicKey} className="rounded-md border border-border p-3">
                                    <p className="break-all font-mono text-xs text-foreground/90">
                                        {account.publicKey}
                                    </p>
                                    <div className="mt-2 flex flex-wrap gap-1">
                                        <Badge variant="secondary">Total: {account.nfts?.length || 0}</Badge>
                                        {(account.nfts || []).map((id) => (
                                            <Badge key={id} variant="outline" className="font-mono text-[10px]">
                                                {id}
                                            </Badge>
                                        ))}
                                    </div>
                                </div>
                            ))}
                        </div>
                    </CardContent>
                </Card>
            </div>

        </div>
    );
};

export default DashboardView;
