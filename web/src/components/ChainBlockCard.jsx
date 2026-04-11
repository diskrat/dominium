import React from "react";
import { Badge } from "./ui/badge";
import { Card, CardContent } from "./ui/card";

const normalizeBlock = (block) => {
    const hash = block.hash || block.Hash || "---";
    return {
        hash,
        prevHash: block.hash_of_previous || block.ParentHash || "00000000",
        txCount:
            block.tx_count !== undefined ? block.tx_count : block.TxCount || 0,
        difficulty: block.difficulty || block.Difficulty || 0,
        minerId: block.miner_id || block.MinerID || "Desconhecido",
        height: block.height,
    };
};

const ChainBlockCard = ({
    block,
    index,
    arrowColor,
    borderColor,
    boxShadow,
    tagColor,
    tagLabel,
    titlePrefix,
    showCanonicalHeight,
}) => {
    const normalized = normalizeBlock(block);
    const blockNumber =
        showCanonicalHeight && normalized.height !== undefined
            ? normalized.height
            : index;

    return (
        <div className="flex flex-col items-center">
            <Card
                className="relative mx-auto w-full max-w-[260px] border-2"
                style={{ borderColor, boxShadow }}
            >
                <Badge
                    variant={tagColor === "red" ? "destructive" : "default"}
                    className="absolute -top-2 right-3"
                >
                    {tagLabel}
                </Badge>
                <CardContent className="space-y-3 p-4">
                    <h4 className="text-sm font-semibold tracking-wide text-muted-foreground">
                        {titlePrefix}
                        {blockNumber}
                    </h4>
                    <div className="h-px bg-border" />
                    <p className="text-[10px] font-semibold text-primary">HASH COMPLETO</p>
                    <p className="break-all font-mono text-xs text-foreground">{normalized.hash}</p>
                    <div className="grid grid-cols-2 gap-2 text-xs">
                        <div className="rounded-md bg-secondary/50 p-2">
                            <p className="text-[10px] text-muted-foreground">TXs</p>
                            <p className="text-sm font-semibold">{normalized.txCount}</p>
                        </div>
                        <div className="rounded-md bg-secondary/50 p-2">
                            <p className="text-[10px] text-muted-foreground">DIFICULDADE</p>
                            <p className="text-sm font-semibold">{normalized.difficulty}</p>
                        </div>
                    </div>
                    <p className="border-t border-border pt-2 text-[11px] text-muted-foreground">
                        Minerado por <span className="text-primary">{normalized.minerId}</span>
                    </p>
                </CardContent>
            </Card>
            <div className="my-1 text-lg" style={{ color: arrowColor }}>
                ↑
            </div>
            <p className="mx-auto max-w-[260px] break-words text-center font-mono text-xs text-muted-foreground">
                PREV: {normalized.prevHash.substring(0, 36)}...
            </p>
        </div>
    );
};

export default ChainBlockCard;
