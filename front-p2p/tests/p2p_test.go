package tests

import (
	"testing"

	"front-p2p/mocks"
)

func TestMockRabbitMQ_Connect(t *testing.T) {
	mock := mocks.NewMockRabbitMQ("test-peer")

	err := mock.Connect()
	if err != nil {
		t.Errorf("Connect failed: %v", err)
	}

	if !mock.IsConnected() {
		t.Error("Expected to be connected")
	}
}

func TestMockRabbitMQ_Disconnect(t *testing.T) {
	mock := mocks.NewMockRabbitMQ("test-peer")
	mock.Connect()

	err := mock.Disconnect()
	if err != nil {
		t.Errorf("Disconnect failed: %v", err)
	}

	if mock.IsConnected() {
		t.Error("Expected to be disconnected")
	}
}

func TestMockRabbitMQ_Publish(t *testing.T) {
	mock := mocks.NewMockRabbitMQ("test-peer")
	mock.Connect()

	payload := map[string]string{
		"message": "test message",
	}

	err := mock.Publish("test_type", payload)
	if err != nil {
		t.Errorf("Publish failed: %v", err)
	}

	messages := mock.GetPublishedMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	if messages[0].Type != "test_type" {
		t.Errorf("Expected message type 'test_type', got '%s'", messages[0].Type)
	}
}

func TestMockRabbitMQ_PublishWhenDisconnected(t *testing.T) {
	mock := mocks.NewMockRabbitMQ("test-peer")
	// Não conecta - deve falhar

	err := mock.Publish("test_type", "payload")
	if err == nil {
		t.Error("Expected error when publishing while disconnected")
	}
}

func TestMockRabbitMQ_Subscribe(t *testing.T) {
	mock := mocks.NewMockRabbitMQ("test-peer")
	mock.Connect()

	called := false
	handler := func(msg *mocks.MockMessage) error {
		called = true
		return nil
	}

	err := mock.Subscribe("test_type", handler)
	if err != nil {
		t.Errorf("Subscribe failed: %v", err)
	}

	// Simula mensagem recebida
	mock.SimulateIncomingMessage("test_type", "test payload")

	if !called {
		t.Error("Handler was not called")
	}
}

func TestMockRabbitMQ_GetPublishedMessagesByType(t *testing.T) {
	mock := mocks.NewMockRabbitMQ("test-peer")
	mock.Connect()

	mock.Publish("type1", "payload1")
	mock.Publish("type2", "payload2")
	mock.Publish("type1", "payload3")

	type1Messages := mock.GetPublishedMessagesByType("type1")
	if len(type1Messages) != 2 {
		t.Errorf("Expected 2 messages of type1, got %d", len(type1Messages))
	}

	type2Messages := mock.GetPublishedMessagesByType("type2")
	if len(type2Messages) != 1 {
		t.Errorf("Expected 1 message of type2, got %d", len(type2Messages))
	}
}

func TestMockRabbitMQ_FailureSimulation(t *testing.T) {
	mock := mocks.NewMockRabbitMQ("test-peer")
	mock.SetShouldFail(true)

	err := mock.Connect()
	if err == nil {
		t.Error("Expected connect to fail")
	}

	mock.SetShouldFail(false)
	mock.Connect()

	mock.SetShouldFail(true)
	err = mock.Publish("test_type", "payload")
	if err == nil {
		t.Error("Expected publish to fail")
	}
}

func TestMockRabbitMQ_ClearPublishedMessages(t *testing.T) {
	mock := mocks.NewMockRabbitMQ("test-peer")
	mock.Connect()

	mock.Publish("type1", "payload1")
	mock.Publish("type2", "payload2")

	messages := mock.GetPublishedMessages()
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages before clear, got %d", len(messages))
	}

	mock.ClearPublishedMessages()

	messages = mock.GetPublishedMessages()
	if len(messages) != 0 {
		t.Errorf("Expected 0 messages after clear, got %d", len(messages))
	}
}

// Benchmark tests
func BenchmarkMockRabbitMQ_Publish(b *testing.B) {
	mock := mocks.NewMockRabbitMQ("bench-peer")
	mock.Connect()

	payload := map[string]string{"data": "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.Publish("bench_type", payload)
	}
}

func BenchmarkMockRabbitMQ_Subscribe(b *testing.B) {
	mock := mocks.NewMockRabbitMQ("bench-peer")
	mock.Connect()

	handler := func(msg *mocks.MockMessage) error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.Subscribe("bench_type", handler)
	}
}
