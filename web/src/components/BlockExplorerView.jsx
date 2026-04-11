import React, { useMemo, useRef, useState } from "react";
import {
    Background,
    Controls,
    ReactFlow,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";
import { Blocks } from "lucide-react";
import { Badge } from "./ui/badge";
import { Button } from "./ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "./ui/tabs";

const ZERO_HASH_RE = /^0+$/;
const MIN_ZOOM = 0.2;
const MAX_ZOOM = 1.8;
const DEFAULT_ZOOM = 1.3;
const DEFAULT_VIEWPORT = { x: 0, y: 0, zoom: DEFAULT_ZOOM };
const NODE_WIDTH = 230;
const NODE_HEIGHT = 132;

const normalizeBlock = (block, index) => {
    const hash = block.hash || block.Hash || `unknown-${index}`;
    const parentHash =
        block.hash_of_previous || block.ParentHash || block.previousHash || "";
    const height =
        typeof block.height === "number"
            ? block.height
            : typeof block.Height === "number"
                ? block.Height
                : index;
    const minerId = block.miner_id || block.MinerID || "desconhecido";
    const txCount =
        block.tx_count !== undefined ? block.tx_count : block.TxCount || 0;

    return {
        hash,
        parentHash,
        height,
        minerId,
        txCount,
    };
};

const isZeroHash = (hash) => !hash || ZERO_HASH_RE.test(hash);

const statusTheme = {
    consolidated: {
        label: "Consolidado",
        className: "border-emerald-500 bg-emerald-500/10",
        badge: "success",
        color: "#10b981",
    },
    fork: {
        label: "Fork",
        className: "border-amber-500 bg-amber-500/10",
        badge: "secondary",
        color: "#f59e0b",
    },
    orphan: {
        label: "Orfao",
        className: "border-rose-500 bg-rose-500/10",
        badge: "destructive",
        color: "#f43f5e",
    },
};

const buildLocalCanonicalHashSet = (rawBlocks) => {
    const blocks = rawBlocks.map((block, index) => normalizeBlock(block, index));
    if (blocks.length === 0) {
        return new Set();
    }

    const byHash = new Map(blocks.map((block) => [block.hash, block]));
    const bestTip = [...blocks].sort((a, b) => {
        if (b.height !== a.height) {
            return b.height - a.height;
        }
        return b.timestamp - a.timestamp;
    })[0];

    const localCanonical = new Set();
    let current = bestTip;

    while (current) {
        localCanonical.add(current.hash);
        if (!current.parentHash || isZeroHash(current.parentHash)) {
            break;
        }
        current = byHash.get(current.parentHash);
    }

    return localCanonical;
};

const buildFlowData = (rawBlocks, canonicalHashes) => {
    const blocks = rawBlocks.map((block, index) => normalizeBlock(block, index));
    const blockByHash = new Map(blocks.map((block) => [block.hash, block]));
    const canonicalSet = canonicalHashes instanceof Set ? canonicalHashes : new Set();

    const statusByHash = new Map();
    blocks.forEach((block) => {
        if (canonicalSet.has(block.hash)) {
            statusByHash.set(block.hash, "consolidated");
            return;
        }

        const parentExists = blockByHash.has(block.parentHash);
        if (!isZeroHash(block.parentHash) && !parentExists) {
            statusByHash.set(block.hash, "orphan");
            return;
        }

        statusByHash.set(block.hash, "fork");
    });

    const nodeGapX = 280;
    const nodeGapY = 180;
    const nodes = [];
    const positionByHash = new Map();
    const occupiedByDepth = new Map();

    const getDepth = (block) => Math.max(0, Number(block.height) || 0);
    const occupy = (depth, x) => {
        if (!occupiedByDepth.has(depth)) {
            occupiedByDepth.set(depth, new Set());
        }
        occupiedByDepth.get(depth).add(x);
    };
    const isOccupied = (depth, x) => occupiedByDepth.get(depth)?.has(x);
    const resolveCollision = (depth, x, preferredDirection) => {
        let candidate = x;
        while (isOccupied(depth, candidate)) {
            candidate += preferredDirection * nodeGapX;
        }
        return candidate;
    };

    const canonicalBlocks = blocks
        .filter((block) => statusByHash.get(block.hash) === "consolidated")
        .sort((a, b) => getDepth(a) - getDepth(b));

    canonicalBlocks.forEach((block) => {
        const depth = getDepth(block);
        positionByHash.set(block.hash, { x: 0, y: depth * nodeGapY });
        occupy(depth, 0);
    });

    const nonCanonicalBlocks = blocks
        .filter((block) => statusByHash.get(block.hash) !== "consolidated")
        .sort((a, b) => {
            const byDepth = getDepth(a) - getDepth(b);
            if (byDepth !== 0) {
                return byDepth;
            }
            return String(a.hash).localeCompare(String(b.hash));
        });

    const childIndexByParent = new Map();
    const orphanIndexRef = { value: 0 };

    const nextChildIndex = (parentHash) => {
        const current = childIndexByParent.get(parentHash) || 0;
        childIndexByParent.set(parentHash, current + 1);
        return current;
    };

    const getBranchOffset = (index) => {
        const side = index % 2 === 0 ? -1 : 1;
        const distance = Math.floor(index / 2) + 1;
        return side * distance * nodeGapX;
    };

    nonCanonicalBlocks.forEach((block) => {
        const depth = getDepth(block);
        const parent = blockByHash.get(block.parentHash);
        let baseX = 0;
        let direction = 1;

        if (!parent || !positionByHash.has(parent.hash)) {
            const orphanIndex = orphanIndexRef.value;
            orphanIndexRef.value += 1;
            baseX = getBranchOffset(orphanIndex + 2); // empurra orfaos mais para fora
            direction = baseX < 0 ? -1 : 1;
        } else {
            const parentPos = positionByHash.get(parent.hash);
            const parentStatus = statusByHash.get(parent.hash) || "fork";
            const childIndex = nextChildIndex(parent.hash);

            if (parentStatus === "consolidated") {
                // Fork raiz fica ao lado do bloco consolidado no mesmo nivel.
                baseX = parentPos.x + getBranchOffset(childIndex);
                direction = baseX < parentPos.x ? -1 : 1;
            } else {
                // Cadeia de fork continua para baixo na mesma coluna lateral.
                if (childIndex === 0) {
                    baseX = parentPos.x;
                } else {
                    baseX = parentPos.x + getBranchOffset(childIndex);
                }
                direction = baseX < parentPos.x ? -1 : 1;
            }
        }

        const x = resolveCollision(depth, baseX, direction);
        positionByHash.set(block.hash, { x, y: depth * nodeGapY });
        occupy(depth, x);
    });

    blocks.forEach((block) => {
        const status = statusByHash.get(block.hash) || "fork";
        const theme = statusTheme[status];
        const position = positionByHash.get(block.hash) || { x: 0, y: 0 };
        nodes.push({
            id: block.hash,
            position,
            data: {
                label: (
                    <div className={`w-[230px] rounded-md border p-3 ${theme.className}`}>
                        <div className="mb-2 flex items-center justify-between gap-2">
                            <p className="font-mono text-[11px] text-foreground/90">
                                #{block.height}
                            </p>
                            <Badge variant={theme.badge}>{theme.label}</Badge>
                        </div>
                        <p className="mb-1 font-mono text-[11px] text-primary">
                            {block.hash.slice(0, 14)}...
                        </p>
                        <p className="text-[11px] text-muted-foreground">
                            Minerador: {block.minerId}
                        </p>
                        <p className="text-[11px] text-muted-foreground">
                            TXs: {block.txCount}
                        </p>
                    </div>
                ),
            },
            draggable: false,
            selectable: false,
        });
    });

    const edges = blocks
        .filter((block) => blockByHash.has(block.parentHash))
        .map((block) => {
            const status = statusByHash.get(block.hash) || "fork";
            const theme = statusTheme[status];
            return {
                id: `${block.parentHash}-${block.hash}`,
                source: block.parentHash,
                target: block.hash,
                animated: status === "consolidated",
                style: {
                    stroke: theme.color,
                    strokeWidth: status === "consolidated" ? 2.6 : 1.8,
                },
            };
        });

    const stats = {
        consolidated: blocks.filter((b) => statusByHash.get(b.hash) === "consolidated").length,
        fork: blocks.filter((b) => {
            if (statusByHash.get(b.hash) !== "fork") {
                return false;
            }
            const parent = blockByHash.get(b.parentHash);
            return parent && statusByHash.get(parent.hash) === "consolidated";
        }).length,
        orphan: blocks.filter((b) => statusByHash.get(b.hash) === "orphan").length,
    };

    const canonicalFocus = blocks
        .filter((block) => statusByHash.get(block.hash) === "consolidated")
        .sort((a, b) => b.height - a.height)[0];
    const fallbackFocus = [...blocks].sort((a, b) => b.height - a.height)[0];
    const focusHash = canonicalFocus?.hash || fallbackFocus?.hash;
    const focusPoint = positionByHash.get(focusHash) || { x: 0, y: 0 };

    return { nodes, edges, stats, focusPoint };
};

const FlowPanel = ({
    title,
    blocks,
    canonicalHashes,
    viewport,
    onViewportChange,
    onDefaultZoomRequest,
}) => {
    const containerRef = useRef(null);
    const { nodes, edges, stats, focusPoint } = useMemo(
        () => buildFlowData(blocks, canonicalHashes),
        [blocks, canonicalHashes],
    );

    const handleDefaultZoomClick = () => {
        if (!containerRef.current) {
            return;
        }

        const { clientWidth, clientHeight } = containerRef.current;
        const nextViewport = {
            x: clientWidth / 2 - (focusPoint.x + NODE_WIDTH / 2) * DEFAULT_ZOOM,
            y: clientHeight / 2 - (focusPoint.y + NODE_HEIGHT / 2) * DEFAULT_ZOOM,
            zoom: DEFAULT_ZOOM,
        };
        onDefaultZoomRequest(nextViewport);
    };

    if (blocks.length === 0) {
        return (
            <div className="rounded-md border border-dashed border-border py-16 text-center text-sm text-muted-foreground">
                {title}: aguardando blocos...
            </div>
        );
    }

    return (
        <div className="space-y-3">
            <div className="flex flex-wrap items-center gap-2 text-xs">
                <Badge variant="success">Consolidados: {stats.consolidated}</Badge>
                <Badge variant="secondary">Forks (branches): {stats.fork}</Badge>
                <Badge variant="destructive">Orfaos: {stats.orphan}</Badge>
            </div>
            <div
                ref={containerRef}
                className="relative h-[560px] rounded-md border border-border bg-slate-950/80"
            >
                <Button
                    variant="secondary"
                    size="sm"
                    onClick={handleDefaultZoomClick}
                    className="absolute right-3 top-3 z-10"
                >
                    Zoom padrao: {Math.round(DEFAULT_VIEWPORT.zoom * 100)}%
                </Button>
                <ReactFlow
                    nodes={nodes}
                    edges={edges}
                    colorMode="dark"
                    viewport={viewport}
                    onViewportChange={onViewportChange}
                    proOptions={{ hideAttribution: true }}
                    minZoom={MIN_ZOOM}
                    maxZoom={MAX_ZOOM}
                    nodesDraggable={false}
                    nodesConnectable={false}
                    elementsSelectable={false}
                    className="block-explorer-flow"
                >
                    <Controls showInteractive={false} />
                    <Background gap={18} size={1} />
                </ReactFlow>
            </div>
        </div>
    );
};

const BlockExplorerView = ({ nodes, allBlocks, canonicalHashes }) => {
    const [sharedViewport, setSharedViewport] = useState(DEFAULT_VIEWPORT);

    return (
        <Card>
            <CardHeader>
                <CardTitle className="flex items-center gap-2">
                    <Blocks className="h-4 w-4 text-primary" />
                    Explorador de Blocos
                </CardTitle>
            </CardHeader>
            <CardContent>
                <Tabs defaultValue="global" className="w-full">
                    <TabsList className="mb-3 flex h-auto w-full flex-wrap justify-start gap-2 bg-transparent p-0">
                        <TabsTrigger value="global">Visao Global (Consenso)</TabsTrigger>
                        {nodes.map((node) => (
                            <TabsTrigger key={node.id} value={node.id}>
                                Visao do {node.id}
                            </TabsTrigger>
                        ))}
                    </TabsList>

                    <TabsContent value="global">
                        <FlowPanel
                            title="Visao Global"
                            blocks={allBlocks}
                            canonicalHashes={canonicalHashes}
                            viewport={sharedViewport}
                            onViewportChange={setSharedViewport}
                            onDefaultZoomRequest={setSharedViewport}
                        />
                    </TabsContent>

                    {nodes.map((node) => {
                        const localChain = node.local_blocks || [];
                        const localCanonicalHashes = buildLocalCanonicalHashSet(localChain);
                        return (
                            <TabsContent key={node.id} value={node.id}>
                                <FlowPanel
                                    title={`Visao de ${node.id}`}
                                    blocks={[...localChain].sort((a, b) => a.timestamp - b.timestamp)}
                                    canonicalHashes={localCanonicalHashes}
                                    viewport={sharedViewport}
                                    onViewportChange={setSharedViewport}
                                    onDefaultZoomRequest={setSharedViewport}
                                />
                            </TabsContent>
                        );
                    })}
                </Tabs>
            </CardContent>
        </Card>
    );
};

export default BlockExplorerView;
