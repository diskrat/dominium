package mocks

import (
	"encoding/json"
	"fmt"
	"sync"
)

// MockMessage representa uma mensagem mockada
type MockMessage struct {
	Type    string      `json:"type"`
	From    string      `json:"from"`
	Payload interface{} `json:"payload"`
}

// MockRabbitMQ simula o comportamento do RabbitMQ para testes
type MockRabbitMQ struct {
	peerID      string
	messages    []MockMessage
	handlers    map[string]func(*MockMessage) error
	subscribers map[string][]chan MockMessage
	mu          sync.RWMutex
	published   []MockMessage // Histórico de mensagens publicadas
	isConnected bool
	shouldFail  bool // Para simular falhas
}

// NewMockRabbitMQ cria uma nova instância do mock
func NewMockRabbitMQ(peerID string) *MockRabbitMQ {
	return &MockRabbitMQ{
		peerID:      peerID,
		messages:    make([]MockMessage, 0),
		handlers:    make(map[string]func(*MockMessage) error),
		subscribers: make(map[string][]chan MockMessage),
		published:   make([]MockMessage, 0),
		isConnected: false, // Começa desconectado
		shouldFail:  false,
	}
}

// Connect simula a conexão com RabbitMQ
func (m *MockRabbitMQ) Connect() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldFail {
		return fmt.Errorf("mock connection failed")
	}

	m.isConnected = true
	return nil
}

// Disconnect simula a desconexão
func (m *MockRabbitMQ) Disconnect() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.isConnected = false
	return nil
}

// Publish simula o envio de mensagem
func (m *MockRabbitMQ) Publish(messageType string, payload interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isConnected {
		return fmt.Errorf("not connected")
	}

	if m.shouldFail {
		return fmt.Errorf("mock publish failed")
	}

	msg := MockMessage{
		Type:    messageType,
		From:    m.peerID,
		Payload: payload,
	}

	m.published = append(m.published, msg)
	m.messages = append(m.messages, msg)

	// Notifica subscribers
	if channels, ok := m.subscribers[messageType]; ok {
		for _, ch := range channels {
			select {
			case ch <- msg:
			default:
				// Canal cheio, ignora
			}
		}
	}

	return nil
}

// Subscribe simula a subscrição a mensagens
func (m *MockRabbitMQ) Subscribe(messageType string, handler func(*MockMessage) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isConnected {
		return fmt.Errorf("not connected")
	}

	m.handlers[messageType] = handler
	return nil
}

// RegisterHandler registra um handler para um tipo de mensagem
func (m *MockRabbitMQ) RegisterHandler(messageType string, handler func(*MockMessage) error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.handlers[messageType] = handler
}

// GetPublishedMessages retorna todas as mensagens publicadas
func (m *MockRabbitMQ) GetPublishedMessages() []MockMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return append([]MockMessage{}, m.published...)
}

// GetPublishedMessagesByType retorna mensagens de um tipo específico
func (m *MockRabbitMQ) GetPublishedMessagesByType(messageType string) []MockMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	filtered := make([]MockMessage, 0)
	for _, msg := range m.published {
		if msg.Type == messageType {
			filtered = append(filtered, msg)
		}
	}

	return filtered
}

// ClearPublishedMessages limpa o histórico de mensagens
func (m *MockRabbitMQ) ClearPublishedMessages() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.published = make([]MockMessage, 0)
	m.messages = make([]MockMessage, 0)
}

// SetShouldFail configura se o mock deve simular falhas
func (m *MockRabbitMQ) SetShouldFail(shouldFail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.shouldFail = shouldFail
}

// IsConnected retorna se está conectado
func (m *MockRabbitMQ) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.isConnected
}

// SimulateIncomingMessage simula o recebimento de uma mensagem
func (m *MockRabbitMQ) SimulateIncomingMessage(messageType string, payload interface{}) error {
	m.mu.RLock()
	handler, exists := m.handlers[messageType]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no handler registered for message type: %s", messageType)
	}

	msg := &MockMessage{
		Type:    messageType,
		From:    "external",
		Payload: payload,
	}

	return handler(msg)
}

// ToJSON converte mensagem para JSON
func (m *MockMessage) ToJSON() (string, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
