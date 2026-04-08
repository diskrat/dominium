import React, { useState, useEffect } from "react";
import {
    Layout,
    Menu,
    Typography,
    Space,
    Button,
    Card,
    Row,
    Col,
    Table,
    List,
    Tag,
    message,
    Divider,
    Statistic,
    Tabs,
    Modal,
    Slider,
    InputNumber,
} from "antd";
import {
    DashboardOutlined,
    ExperimentOutlined,
    ThunderboltOutlined,
    NodeIndexOutlined,
    ApiOutlined,
    WarningOutlined,
} from "@ant-design/icons";
import "./App.css";
import { executeChaosMint } from "./services/dominiumApi";
import { SettingOutlined } from "@ant-design/icons";

const { Header, Content, Sider } = Layout;
const { Title, Text } = Typography;

const App = () => {
    const [currentView, setCurrentView] = useState("dashboard");
    const [socket, setSocket] = useState(null);
    const [nodes, setNodes] = useState([]);
    const [newDifficulty, setNewDifficulty] = useState(24);
    const [isUpdatingDiff, setIsUpdatingDiff] = useState(false);

    const [allBlocks, setAllBlocks] = useState([]);
    const [canonicalHashes, setCanonicalHashes] = useState(new Set());
    const [accounts, setAccounts] = useState([]);
    const [mempool, setMempool] = useState([]);
    const [isMinting, setIsMinting] = useState(false);
    const [attackLogs, setAttackLogs] = useState([]);

    const handleDifficultyUpdate = async () => {
        setIsUpdatingDiff(true);
        message.loading({
            content: "Enviando comando para a rede...",
            key: "diff",
        });
        try {
            const response = await fetch(
                "http://localhost:8085/network/difficulty",
                {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ difficulty: newDifficulty }),
                },
            );

            if (!response.ok) throw new Error("Falha ao atualizar");

            message.success({
                content: `Dificuldade atualizada para ${newDifficulty} bits! Os próximos blocos usarão esta regra.`,
                key: "diff",
                duration: 4,
            });
        } catch (err) {
            message.error({
                content: "Erro ao comunicar com o Gateway.",
                key: "diff",
            });
        }
        setIsUpdatingDiff(false);
    };

    const handleChaosMintClick = async () => {
        setIsMinting(true);
        message.loading({ content: "Disparando transações...", key: "chaos" });
        const successCount = await executeChaosMint(10);
        if (successCount > 0) {
            message.success({
                content: `${successCount} transações enviadas!`,
                key: "chaos",
                duration: 3,
            });
        } else {
            message.error({
                content: "Falha ao enviar transações.",
                key: "chaos",
                duration: 3,
            });
        }
        setIsMinting(false);
    };
    const generateRandomHex = (size) => {
        return [...Array(size)]
            .map(() => Math.floor(Math.random() * 16).toString(16))
            .join("");
    };

    const handleGenerateWalletClick = async () => {
        const hide = message.loading(
            "Gerando chaves criptográficas na Curva Elíptica (P-256)...",
            0,
        );

        try {
            const response = await fetch(
                "http://localhost:8085/wallet/generate",
            );

            if (!response.ok) {
                throw new Error("Falha ao gerar a carteira no nó Gateway.");
            }

            const data = await response.json();
            hide(); // Esconde o loading

            // Exibe um pop-up bonito na tela com as chaves reais
            Modal.success({
                title: "Carteira ECDSA (P-256) Gerada!",
                content: (
                    <div style={{ marginTop: "20px" }}>
                        <Text type="secondary">
                            Este par de chaves é matematicamente válido na rede
                            Dominium.
                        </Text>
                        <div
                            style={{
                                marginTop: "15px",
                                marginBottom: "15px",
                                padding: "10px",
                                background: "#f5f5f5",
                                borderRadius: "8px",
                            }}
                        >
                            <Text strong style={{ color: "#1890ff" }}>
                                Chave Pública (Seu Endereço):
                            </Text>
                            <br />
                            <Text
                                copyable
                                code
                                style={{
                                    fontSize: "11px",
                                    wordBreak: "break-all",
                                }}
                            >
                                {data.public_key}
                            </Text>
                        </div>
                        <div
                            style={{
                                padding: "10px",
                                background: "#fff1f0",
                                borderRadius: "8px",
                                border: "1px solid #ffa39e",
                            }}
                        >
                            <Text strong style={{ color: "#cf1322" }}>
                                Chave Privada (SECRETA):
                            </Text>
                            <br />
                            <Text
                                copyable
                                code
                                style={{
                                    fontSize: "11px",
                                    wordBreak: "break-all",
                                }}
                            >
                                {data.private_key}
                            </Text>
                        </div>
                    </div>
                ),
                width: 600,
                okText: "Guardei minha chave!",
            });
        } catch (err) {
            hide();
            message.error(err.message);
        }
    };

    useEffect(() => {
        const ws = new WebSocket("ws://localhost:8080/ws");
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
        setSocket(ws);
        return () => ws.close();
    }, []);

    const menuItems = [
        {
            key: "dashboard",
            icon: <DashboardOutlined />,
            label: "Observabilidade",
        },
        { key: "simulator", icon: <ExperimentOutlined />, label: "Simulador" },
        { key: "attacks", icon: <ThunderboltOutlined />, label: "Ataques" },
    ];

    const handleDoubleSpendClick = async () => {
        message.warning({
            content: "Iniciando Ataque de Gasto Duplo...",
            key: "attack",
        });
        try {
            const response = await fetch(
                "http://localhost:8085/attacks/double-spend",
                { method: "POST" },
            );

            if (!response.ok) {
                const errData = await response.text();
                throw new Error(errData);
            }

            const data = await response.json();
            message.error({
                content: `Ataque enviado! NFT: ${data.nft}. Aguardando validação dos nós...`,
                duration: 5,
                key: "attack",
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
            message.error(`Falha no ataque: ${err.message}`);
        }
    };

    const renderDashboard = () => (
        <div>
            <Title level={2}>Painel de Observabilidade - Rede Dominium</Title>
            <Row gutter={[16, 16]}>
                {/* 1. Status dos Nós */}
                <Col span={24}>
                    <Card title="Status dos Nós" bordered={false}>
                        <Row gutter={16}>
                            {nodes.map((node) => (
                                <Col key={node.id} span={6}>
                                    <Card
                                        size="small"
                                        style={{ textAlign: "center" }}
                                    >
                                        <NodeIndexOutlined
                                            style={{
                                                fontSize: "24px",
                                                color:
                                                    node.status === "mining"
                                                        ? "#52c41a"
                                                        : "#d9d9d9",
                                            }}
                                        />
                                        <Title level={4}>{node.id}</Title>
                                        <Tag
                                            color={
                                                node.status === "mining"
                                                    ? "green"
                                                    : "default"
                                            }
                                        >
                                            {node.status === "mining"
                                                ? "Minerando"
                                                : "Ocioso"}
                                        </Tag>
                                        <br />
                                        <Text>
                                            Dificuldade: {node.difficulty} bits
                                        </Text>
                                        <br />
                                        <Text>Altura: {node.blockHeight}</Text>
                                    </Card>
                                </Col>
                            ))}
                        </Row>
                    </Card>
                </Col>

                {/* 2. Mempool */}
                <Col span={12}>
                    <Card title="Mempool - Fila de Espera" bordered={false}>
                        <div style={{ height: "240px", overflowY: "auto" }}>
                            <List
                                size="small"
                                dataSource={mempool}
                                renderItem={(tx) => {
                                    const id = tx.id || tx.ID || "unknown";
                                    const type =
                                        tx.type !== undefined
                                            ? tx.type
                                            : tx.Type;
                                    const ts = tx.timestamp || tx.Timestamp;
                                    const date = ts
                                        ? new Date(Number(ts) / 1000000)
                                        : null;
                                    const timeLabel =
                                        date && !isNaN(date)
                                            ? date.toLocaleTimeString()
                                            : "pendente";

                                    return (
                                        <List.Item>
                                            <Text code>
                                                {id.substring(0, 10)}...
                                            </Text>
                                            <Tag
                                                color={
                                                    type === 2
                                                        ? "green"
                                                        : "blue"
                                                }
                                            >
                                                {type === 2
                                                    ? "MINT"
                                                    : "TRANSFER"}
                                            </Tag>
                                            <Text type="secondary">
                                                {timeLabel}
                                            </Text>
                                        </List.Item>
                                    );
                                }}
                                locale={{ emptyText: "Mempool vazia" }}
                            />
                        </div>
                    </Card>
                </Col>

                {/* 3. Contas */}
                <Col span={12}>
                    <Card title="Estado das Contas" bordered={false}>
                        <Table
                            size="small"
                            rowKey="publicKey"
                            dataSource={accounts}
                            pagination={false}
                            scroll={{ y: 240 }}
                            columns={[
                                {
                                    title: "Chave Pública",
                                    dataIndex: "publicKey",
                                    key: "publicKey",
                                    ellipsis: true,
                                },
                                {
                                    title: "NFTs (IDs)",
                                    dataIndex: "nfts",
                                    key: "nfts",
                                    render: (nfts) => (
                                        <div>
                                            <Tag
                                                color="purple"
                                                style={{ marginBottom: "4px" }}
                                            >
                                                Total: {nfts ? nfts.length : 0}
                                            </Tag>
                                            {/* Se tiver NFTs, lista os IDs abaixo */}
                                            {nfts && nfts.length > 0 && (
                                                <div
                                                    style={{
                                                        display: "flex",
                                                        flexDirection: "column",
                                                        gap: "2px",
                                                    }}
                                                >
                                                    {nfts.map((id) => (
                                                        <Text
                                                            code
                                                            key={id}
                                                            style={{
                                                                fontSize:
                                                                    "10px",
                                                                color: "#888",
                                                            }}
                                                        >
                                                            {id}
                                                        </Text>
                                                    ))}
                                                </div>
                                            )}
                                        </div>
                                    ),
                                },
                            ]}
                        />
                    </Card>
                </Col>

                {/* 4. Explorador de Blocos e Forks */}
                <Col span={24}>
                    <Card
                        title={`Explorador de Blocos (Visão Global e Locais)`}
                        bordered={false}
                    >
                        <Tabs defaultActiveKey="global">
                            {/* ABA 1: Visão Global */}
                            <Tabs.TabPane
                                tab={
                                    <span
                                        style={{
                                            fontWeight: "bold",
                                            color: "#1890ff",
                                        }}
                                    >
                                        Visão Global (Consenso)
                                    </span>
                                }
                                key="global"
                            >
                                <div
                                    style={{
                                        height: "400px",
                                        backgroundColor: "#001529",
                                        overflow: "auto",
                                        padding: "30px",
                                        borderRadius: "8px",
                                        border: "1px solid #1890ff",
                                        display: "grid",
                                        gridAutoFlow: "column",
                                        gridGap: "40px",
                                        alignItems: "start",
                                    }}
                                >
                                    {allBlocks.length === 0 ? (
                                        <Text style={{ color: "#69c0ff" }}>
                                            Aguardando blocos...
                                        </Text>
                                    ) : (
                                        allBlocks.map((block, index) => {
                                            const hash =
                                                block.hash ||
                                                block.Hash ||
                                                "---";
                                            const prevHash =
                                                block.hash_of_previous ||
                                                block.ParentHash ||
                                                "00000000";
                                            const txCount =
                                                block.tx_count !== undefined
                                                    ? block.tx_count
                                                    : block.TxCount || 0;
                                            const difficulty =
                                                block.difficulty ||
                                                block.Difficulty ||
                                                0;
                                            const minerId =
                                                block.miner_id ||
                                                block.MinerID ||
                                                "Desconhecido";
                                            const isCanonical =
                                                canonicalHashes.has(hash);

                                            return (
                                                <div
                                                    key={hash + index}
                                                    style={{
                                                        display: "flex",
                                                        flexDirection: "column",
                                                        alignItems: "center",
                                                    }}
                                                >
                                                    <div
                                                        style={{
                                                            minWidth: "240px",
                                                            padding: "15px",
                                                            backgroundColor:
                                                                "#1f1f1f",
                                                            borderRadius: "8px",
                                                            border: `2px solid ${isCanonical ? "#52c41a" : "#f5222d"}`,
                                                            boxShadow:
                                                                "0 0 15px rgba(24, 144, 255, 0.1)",
                                                            color: "white",
                                                            position:
                                                                "relative",
                                                        }}
                                                    >
                                                        <Tag
                                                            color={
                                                                isCanonical
                                                                    ? "green"
                                                                    : "red"
                                                            }
                                                            style={{
                                                                position:
                                                                    "absolute",
                                                                top: "-10px",
                                                                right: "10px",
                                                            }}
                                                        >
                                                            {isCanonical
                                                                ? "CONFIRMADO"
                                                                : "FORK / ORFÃO"}
                                                        </Tag>
                                                        <Title
                                                            level={5}
                                                            style={{
                                                                color: "#aaa",
                                                                margin: 0,
                                                            }}
                                                        >
                                                            BLOCK #
                                                            {block.height !==
                                                                undefined &&
                                                            isCanonical
                                                                ? block.height
                                                                : index}
                                                        </Title>
                                                        <Divider
                                                            style={{
                                                                background:
                                                                    "#333",
                                                                margin: "10px 0",
                                                            }}
                                                        />
                                                        <Text
                                                            strong
                                                            style={{
                                                                color: "#1890ff",
                                                                fontSize:
                                                                    "10px",
                                                            }}
                                                        >
                                                            HASH COMPLETO:
                                                        </Text>
                                                        <Text
                                                            copyable
                                                            style={{
                                                                color: "#fff",
                                                                fontSize:
                                                                    "14px",
                                                                fontFamily:
                                                                    "monospace",
                                                                display:
                                                                    "block",
                                                                marginBottom:
                                                                    "10px",
                                                            }}
                                                        >
                                                            {hash}
                                                        </Text>
                                                        <Row gutter={8}>
                                                            <Col span={12}>
                                                                <Statistic
                                                                    title={
                                                                        <span
                                                                            style={{
                                                                                color: "#888",
                                                                                fontSize:
                                                                                    "10px",
                                                                            }}
                                                                        >
                                                                            TXs
                                                                        </span>
                                                                    }
                                                                    value={
                                                                        txCount
                                                                    }
                                                                    valueStyle={{
                                                                        color: "#fff",
                                                                        fontSize:
                                                                            "14px",
                                                                    }}
                                                                />
                                                            </Col>
                                                            <Col span={12}>
                                                                <Statistic
                                                                    title={
                                                                        <span
                                                                            style={{
                                                                                color: "#888",
                                                                                fontSize:
                                                                                    "10px",
                                                                            }}
                                                                        >
                                                                            DIFICULDADE
                                                                            (bits)
                                                                        </span>
                                                                    }
                                                                    value={
                                                                        difficulty
                                                                    }
                                                                    valueStyle={{
                                                                        color: "#fff",
                                                                        fontSize:
                                                                            "14px",
                                                                    }}
                                                                />
                                                            </Col>
                                                        </Row>
                                                        <div
                                                            style={{
                                                                marginTop:
                                                                    "12px",
                                                                fontSize:
                                                                    "10px",
                                                                color: "#666",
                                                                borderTop:
                                                                    "1px solid #333",
                                                                paddingTop:
                                                                    "5px",
                                                            }}
                                                        >
                                                            Minerado por:{" "}
                                                            <span
                                                                style={{
                                                                    color: "#1890ff",
                                                                }}
                                                            >
                                                                {minerId}
                                                            </span>
                                                        </div>
                                                    </div>
                                                    <div
                                                        style={{
                                                            color: "#1890ff",
                                                            margin: "5px 0",
                                                            fontSize: "18px",
                                                        }}
                                                    >
                                                        ↑
                                                    </div>
                                                    <Text
                                                        style={{
                                                            color: "#888",
                                                            fontSize: "16px",
                                                            fontFamily:
                                                                "monospace",
                                                            fontWeight: "bold",
                                                            display: "block",
                                                            marginTop: "8px",
                                                        }}
                                                    >
                                                        PREV:{" "}
                                                        {prevHash.substring(
                                                            0,
                                                            36,
                                                        )}
                                                        ...
                                                    </Text>
                                                </div>
                                            );
                                        })
                                    )}
                                </div>
                            </Tabs.TabPane>

                            {/* ABAS SECUNDÁRIAS: Visões Locais */}
                            {nodes.map((node) => {
                                const localChain = node.local_blocks || [];
                                return (
                                    <Tabs.TabPane
                                        tab={`Visão do ${node.id}`}
                                        key={node.id}
                                    >
                                        <div
                                            style={{
                                                height: "400px",
                                                backgroundColor: "#141414",
                                                overflow: "auto",
                                                padding: "30px",
                                                borderRadius: "8px",
                                                border: "1px dashed #555",
                                                display: "grid",
                                                gridAutoFlow: "column",
                                                gridGap: "40px",
                                                alignItems: "start",
                                            }}
                                        >
                                            {localChain.length === 0 ? (
                                                <div
                                                    style={{
                                                        textAlign: "center",
                                                        width: "100%",
                                                        gridColumn: "1 / -1",
                                                        marginTop: "50px",
                                                    }}
                                                >
                                                    <NodeIndexOutlined
                                                        style={{
                                                            fontSize: "48px",
                                                            color: "#555",
                                                            marginBottom:
                                                                "16px",
                                                        }}
                                                    />
                                                    <Title
                                                        level={4}
                                                        style={{
                                                            color: "#888",
                                                        }}
                                                    >
                                                        Aguardando sincronização
                                                        local...
                                                    </Title>
                                                    <Text
                                                        style={{
                                                            color: "#555",
                                                        }}
                                                    >
                                                        Os dados logo aparecerão
                                                        aqui.
                                                    </Text>
                                                </div>
                                            ) : (
                                                localChain
                                                    .sort(
                                                        (a, b) =>
                                                            a.timestamp -
                                                            b.timestamp,
                                                    )
                                                    .map((block, index) => {
                                                        const hash =
                                                            block.hash ||
                                                            block.Hash ||
                                                            "---";
                                                        const prevHash =
                                                            block.hash_of_previous ||
                                                            block.ParentHash ||
                                                            "00000000";
                                                        const txCount =
                                                            block.tx_count !==
                                                            undefined
                                                                ? block.tx_count
                                                                : block.TxCount ||
                                                                  0;
                                                        const difficulty =
                                                            block.difficulty ||
                                                            block.Difficulty ||
                                                            0;
                                                        const minerId =
                                                            block.miner_id ||
                                                            block.MinerID ||
                                                            "Desconhecido";

                                                        return (
                                                            <div
                                                                key={
                                                                    hash + index
                                                                }
                                                                style={{
                                                                    display:
                                                                        "flex",
                                                                    flexDirection:
                                                                        "column",
                                                                    alignItems:
                                                                        "center",
                                                                }}
                                                            >
                                                                <div
                                                                    style={{
                                                                        minWidth:
                                                                            "240px",
                                                                        padding:
                                                                            "15px",
                                                                        backgroundColor:
                                                                            "#1f1f1f",
                                                                        borderRadius:
                                                                            "8px",
                                                                        border: `2px solid #1890ff`,
                                                                        boxShadow:
                                                                            "0 0 15px rgba(24, 144, 255, 0.2)",
                                                                        color: "white",
                                                                        position:
                                                                            "relative",
                                                                    }}
                                                                >
                                                                    <Tag
                                                                        color="blue"
                                                                        style={{
                                                                            position:
                                                                                "absolute",
                                                                            top: "-10px",
                                                                            right: "10px",
                                                                        }}
                                                                    >
                                                                        VISÃO
                                                                        LOCAL
                                                                    </Tag>
                                                                    <Title
                                                                        level={
                                                                            5
                                                                        }
                                                                        style={{
                                                                            color: "#aaa",
                                                                            margin: 0,
                                                                        }}
                                                                    >
                                                                        BLOCK
                                                                        (Local)
                                                                    </Title>
                                                                    <Divider
                                                                        style={{
                                                                            background:
                                                                                "#333",
                                                                            margin: "10px 0",
                                                                        }}
                                                                    />
                                                                    <Text
                                                                        strong
                                                                        style={{
                                                                            color: "#1890ff",
                                                                            fontSize:
                                                                                "10px",
                                                                        }}
                                                                    >
                                                                        HASH
                                                                        COMPLETO:
                                                                    </Text>
                                                                    <Text
                                                                        copyable
                                                                        style={{
                                                                            color: "#fff",
                                                                            fontSize:
                                                                                "14px",
                                                                            fontFamily:
                                                                                "monospace",
                                                                            display:
                                                                                "block",
                                                                            marginBottom:
                                                                                "10px",
                                                                        }}
                                                                    >
                                                                        {hash}
                                                                    </Text>
                                                                    <Row
                                                                        gutter={
                                                                            8
                                                                        }
                                                                    >
                                                                        <Col
                                                                            span={
                                                                                12
                                                                            }
                                                                        >
                                                                            <Statistic
                                                                                title={
                                                                                    <span
                                                                                        style={{
                                                                                            color: "#888",
                                                                                            fontSize:
                                                                                                "10px",
                                                                                        }}
                                                                                    >
                                                                                        TXs
                                                                                    </span>
                                                                                }
                                                                                value={
                                                                                    txCount
                                                                                }
                                                                                valueStyle={{
                                                                                    color: "#fff",
                                                                                    fontSize:
                                                                                        "14px",
                                                                                }}
                                                                            />
                                                                        </Col>
                                                                        <Col
                                                                            span={
                                                                                12
                                                                            }
                                                                        >
                                                                            <Statistic
                                                                                title={
                                                                                    <span
                                                                                        style={{
                                                                                            color: "#888",
                                                                                            fontSize:
                                                                                                "10px",
                                                                                        }}
                                                                                    >
                                                                                        DIFICULDADE
                                                                                        (bits)
                                                                                    </span>
                                                                                }
                                                                                value={
                                                                                    difficulty
                                                                                }
                                                                                valueStyle={{
                                                                                    color: "#fff",
                                                                                    fontSize:
                                                                                        "14px",
                                                                                }}
                                                                            />
                                                                        </Col>
                                                                    </Row>
                                                                    <div
                                                                        style={{
                                                                            marginTop:
                                                                                "12px",
                                                                            fontSize:
                                                                                "10px",
                                                                            color: "#666",
                                                                            borderTop:
                                                                                "1px solid #333",
                                                                            paddingTop:
                                                                                "5px",
                                                                        }}
                                                                    >
                                                                        Minerado
                                                                        por:{" "}
                                                                        <span
                                                                            style={{
                                                                                color: "#1890ff",
                                                                            }}
                                                                        >
                                                                            {
                                                                                minerId
                                                                            }
                                                                        </span>
                                                                    </div>
                                                                </div>
                                                                <div
                                                                    style={{
                                                                        color: "#555",
                                                                        margin: "5px 0",
                                                                        fontSize:
                                                                            "18px",
                                                                    }}
                                                                >
                                                                    ↑
                                                                </div>
                                                                <Text
                                                                    style={{
                                                                        color: "#888",
                                                                        fontSize:
                                                                            "16px",
                                                                        fontFamily:
                                                                            "monospace",
                                                                        fontWeight:
                                                                            "bold",
                                                                        display:
                                                                            "block",
                                                                        marginTop:
                                                                            "8px",
                                                                    }}
                                                                >
                                                                    PREV:{" "}
                                                                    {prevHash.substring(
                                                                        0,
                                                                        36,
                                                                    )}
                                                                    ...
                                                                </Text>
                                                            </div>
                                                        );
                                                    })
                                            )}
                                        </div>
                                    </Tabs.TabPane>
                                );
                            })}
                        </Tabs>
                    </Card>
                </Col>
            </Row>
        </div>
    );

    const renderSimulator = () => (
        <div>
            <Title level={2}>Simulador de Transações e Rede</Title>
            <Space
                direction="vertical"
                size="large"
                style={{ display: "flex" }}
            >
                {/* Parâmetros da Rede */}
                <Card
                    title="Parâmetros de Consenso (Proof of Work)"
                    bordered={false}
                >
                    <div style={{ marginBottom: "10px" }}>
                        <Text strong>Dificuldade de Mineração (Bits): </Text>
                        <Text type="secondary">
                            Define a quantidade de zeros (0) exigida no Hash do
                            bloco.
                        </Text>
                    </div>
                    <Row gutter={16} align="middle">
                        <Col span={12}>
                            <Slider
                                min={8}
                                max={32}
                                onChange={setNewDifficulty}
                                value={
                                    typeof newDifficulty === "number"
                                        ? newDifficulty
                                        : 24
                                }
                            />
                        </Col>
                        <Col span={4}>
                            <InputNumber
                                min={8}
                                max={32}
                                style={{ margin: "0 16px" }}
                                value={newDifficulty}
                                onChange={setNewDifficulty}
                            />
                        </Col>
                        <Col span={8}>
                            <Button
                                type="primary"
                                icon={<SettingOutlined />}
                                onClick={handleDifficultyUpdate}
                                loading={isUpdatingDiff}
                            >
                                Aplicar Nova Dificuldade
                            </Button>
                        </Col>
                    </Row>
                </Card>

                {/* Seus cards antigos continuam aqui embaixo */}
                <Card title="Controles de Simulação">
                    <Space>
                        <Button
                            type="primary"
                            icon={<ThunderboltOutlined />}
                            onClick={handleChaosMintClick}
                            loading={isMinting}
                        >
                            Chaos Mint (10 transações)
                        </Button>
                        <Button
                            icon={<ApiOutlined />}
                            onClick={handleGenerateWalletClick}
                        >
                            Gerar Carteira Aleatória
                        </Button>
                    </Space>
                </Card>
            </Space>
        </div>
    );

    const renderAttacks = () => (
        <div>
            <Title level={2}>Simulador de Ataques - Double Spend</Title>
            <Space
                direction="vertical"
                size="large"
                style={{ display: "flex" }}
            >
                <Card title="Ataques Disponíveis">
                    <Space>
                        <Button
                            danger
                            icon={<WarningOutlined />}
                            onClick={handleDoubleSpendClick}
                        >
                            Executar Double Spend (Gasto Duplo)
                        </Button>
                    </Space>
                </Card>
                <Card title="Monitor de Ataques">
                    <div style={{ padding: "10px" }}>
                        {attackLogs.length === 0 ? (
                            <div
                                style={{
                                    textAlign: "center",
                                    padding: "40px",
                                    color: "#888",
                                }}
                            >
                                Nenhum ataque detectado na rede recentemente.
                            </div>
                        ) : (
                            <Table
                                size="small"
                                dataSource={attackLogs}
                                pagination={false}
                                rowKey="txId"
                                columns={[
                                    {
                                        title: "Horário",
                                        dataIndex: "time",
                                        key: "time",
                                        width: 100,
                                    },
                                    {
                                        title: "Status",
                                        dataIndex: "status",
                                        key: "status",
                                        render: (text) => (
                                            <Tag
                                                color="red"
                                                icon={<WarningOutlined />}
                                            >
                                                {text}
                                            </Tag>
                                        ),
                                    },
                                    {
                                        title: "ID do Conflito",
                                        dataIndex: "txId",
                                        key: "txId",
                                    },
                                    {
                                        title: "Alvo (NFT)",
                                        dataIndex: "nft",
                                        key: "nft",
                                    },
                                ]}
                            />
                        )}
                    </div>
                </Card>
            </Space>
        </div>
    );

    const renderContent = () => {
        switch (currentView) {
            case "simulator":
                return renderSimulator();
            case "attacks":
                return renderAttacks();
            default:
                return renderDashboard();
        }
    };

    return (
        <Layout style={{ minHeight: "100vh" }}>
            <Sider collapsible>
                <div
                    style={{
                        height: 32,
                        margin: 16,
                        background: "rgba(255, 255, 255, 0.2)",
                    }}
                />
                <Menu
                    theme="dark"
                    mode="inline"
                    selectedKeys={[currentView]}
                    items={menuItems}
                    onClick={({ key }) => setCurrentView(key)}
                />
            </Sider>
            <Layout>
                <Header style={{ padding: "0 24px", background: "#fff" }}>
                    <Title level={3} style={{ margin: "16px 0" }}>
                        Dominium Network Visualizer
                    </Title>
                </Header>
                <Content
                    style={{
                        margin: "24px 16px",
                        padding: 24,
                        background: "#fff",
                    }}
                >
                    {renderContent()}
                </Content>
            </Layout>
        </Layout>
    );
};

export default App;
