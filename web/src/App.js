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
} from "antd";
import {
    DashboardOutlined,
    ExperimentOutlined,
    ThunderboltOutlined,
    NodeIndexOutlined,
    ApiOutlined,
} from "@ant-design/icons";
import "./App.css";

const { Header, Content, Sider } = Layout;
const { Title, Text } = Typography;

const App = () => {
    const [currentView, setCurrentView] = useState("dashboard");
    const [socket, setSocket] = useState(null);
    const [nodes, setNodes] = useState([]);
    const [blocks, setBlocks] = useState([]);
    const [accounts, setAccounts] = useState([]);
    const [mempool, setMempool] = useState([]);

    useEffect(() => {
        // Connect to WebSocket
        const ws = new WebSocket("ws://localhost:8080/ws");

        ws.onopen = () => {
            console.log("WebSocket connected");
        };

        ws.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                setNodes(data.nodes || []);
                setBlocks(data.blocks || []);
                setAccounts(data.accounts || []);
                setMempool(data.mempool || []);
            } catch (error) {
                console.error("Error parsing WebSocket message:", error);
            }
        };

        ws.onerror = (error) => {
            console.error("WebSocket error:", error);
        };

        ws.onclose = () => {
            console.log("WebSocket disconnected");
        };

        setSocket(ws);

        return () => {
            ws.close();
        };
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

    const renderDashboard = () => (
        <div>
            <Title level={2}>Painel de Observabilidade - Rede Dominium</Title>

            <Row gutter={[16, 16]}>
                {/* Node Stats */}
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
                                            Dificuldade: {node.difficulty}
                                        </Text>
                                        <br />
                                        <Text>Altura: {node.blockHeight}</Text>
                                    </Card>
                                </Col>
                            ))}
                        </Row>
                    </Card>
                </Col>

                {/* Mempool Status */}
                <Col span={12}>
                    <Card title="Mempool" bordered={false}>
                        <List
                            size="small"
                            dataSource={mempool.slice(0, 10)}
                            renderItem={(tx) => (
                                <List.Item>
                                    <Text code>
                                        {tx.id
                                            ? tx.id.substring(0, 16)
                                            : "unknown"}
                                        ...
                                    </Text>
                                    <Tag color="blue">
                                        {tx.type === 1 ? "Mint" : "Transfer"}
                                    </Tag>
                                    <Text>
                                        {tx.timestamp
                                            ? new Date(
                                                  tx.timestamp,
                                              ).toLocaleTimeString()
                                            : "unknown"}
                                    </Text>
                                </List.Item>
                            )}
                        />
                        {mempool.length > 10 && (
                            <Text type="secondary">
                                ... e mais {mempool.length - 10} transações
                            </Text>
                        )}
                    </Card>
                </Col>

                {/* Account State */}
                <Col span={12}>
                    <Card title="Estado das Contas" bordered={false}>
                        <Table
                            size="small"
                            columns={[
                                {
                                    title: "Chave Pública",
                                    dataIndex: "publicKey",
                                    key: "publicKey",
                                    ellipsis: true,
                                },
                                {
                                    title: "NFTs",
                                    dataIndex: "nfts",
                                    key: "nfts",
                                    render: (nfts) => (nfts ? nfts.length : 0),
                                },
                            ]}
                            dataSource={accounts.slice(0, 10)}
                            pagination={false}
                        />
                    </Card>
                </Col>

                {/* Blockchain Canvas Placeholder */}
                <Col span={24}>
                    <Card title="Blockchain Canvas" bordered={false}>
                        <div
                            style={{
                                height: "300px",
                                backgroundColor: "#f5f5f5",
                                display: "flex",
                                alignItems: "center",
                                justifyContent: "center",
                            }}
                        >
                            <Text>
                                Canvas interativo da blockchain será
                                implementado aqui
                            </Text>
                        </div>
                    </Card>
                </Col>
            </Row>
        </div>
    );

    const renderSimulator = () => (
        <div>
            <Title level={2}>Simulador de Transações</Title>
            <Space direction="vertical" size="large">
                <Card title="Controles de Simulação">
                    <Space>
                        <Button type="primary" icon={<ThunderboltOutlined />}>
                            Chaos Mint (10 transações)
                        </Button>
                        <Button icon={<ApiOutlined />}>
                            Gerar Carteira Aleatória
                        </Button>
                    </Space>
                </Card>

                <Card title="Fluxo de Transações">
                    <div
                        style={{
                            height: "200px",
                            backgroundColor: "#f5f5f5",
                            display: "flex",
                            alignItems: "center",
                            justifyContent: "center",
                        }}
                    >
                        <Text>
                            Visualização do fluxo API → Kafka → Mempool será
                            implementada aqui
                        </Text>
                    </div>
                </Card>
            </Space>
        </div>
    );

    const renderAttacks = () => (
        <div>
            <Title level={2}>Simulador de Ataques - Double Spend</Title>
            <Space direction="vertical" size="large">
                <Card title="Ataques Disponíveis">
                    <Space>
                        <Button danger icon={<ThunderboltOutlined />}>
                            Race Attack
                        </Button>
                        <Button danger>Simular Fork</Button>
                        <Button>Pausar Mineração (Debug)</Button>
                    </Space>
                </Card>

                <Card title="Monitor de Ataques">
                    <div
                        style={{
                            height: "200px",
                            backgroundColor: "#f5f5f5",
                            display: "flex",
                            alignItems: "center",
                            justifyContent: "center",
                        }}
                    >
                        <Text>
                            Monitor de rejeições e resolução de conflitos será
                            implementado aqui
                        </Text>
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
