import React from "react";
import { AlertTriangle } from "lucide-react";
import { Badge } from "./ui/badge";
import { Button } from "./ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";

const AttacksView = ({ handleDoubleSpendClick, attackLogs }) => {
    return (
        <div className="space-y-4">
            <h2 className="text-xl font-semibold tracking-tight lg:text-2xl">
                Simulador de Ataques - Double Spend
            </h2>

            <Card>
                <CardHeader>
                    <CardTitle>Ataques Disponiveis</CardTitle>
                </CardHeader>
                <CardContent>
                    <Button variant="destructive" onClick={handleDoubleSpendClick}>
                        <AlertTriangle className="h-4 w-4" />
                        Executar Double Spend
                    </Button>
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>Monitor de Ataques</CardTitle>
                </CardHeader>
                <CardContent>
                    {attackLogs.length === 0 ? (
                        <div className="rounded-md border border-dashed border-border py-10 text-center text-sm text-muted-foreground">
                            Nenhum ataque detectado na rede recentemente.
                        </div>
                    ) : (
                        <div className="overflow-x-auto rounded-md border border-border">
                            <table className="w-full text-left text-sm">
                                <thead className="bg-secondary/60 text-muted-foreground">
                                    <tr>
                                        <th className="px-3 py-2">Horario</th>
                                        <th className="px-3 py-2">Status</th>
                                        <th className="px-3 py-2">ID do Conflito</th>
                                        <th className="px-3 py-2">Alvo (NFT)</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {attackLogs.map((log) => (
                                        <tr key={log.txId} className="border-t border-border">
                                            <td className="px-3 py-2">{log.time}</td>
                                            <td className="px-3 py-2">
                                                <Badge variant="destructive">{log.status}</Badge>
                                            </td>
                                            <td className="px-3 py-2 font-mono text-xs">{log.txId}</td>
                                            <td className="px-3 py-2">{log.nft}</td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    )}
                </CardContent>
            </Card>
        </div>
    );
};

export default AttacksView;
