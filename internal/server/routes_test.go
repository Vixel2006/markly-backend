package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"markly/internal/handlers"
	"go.mongodb.org/mongo-driver/mongo" // Import the mongo package
)

// MockDBService is a mock implementation of database.Service for testing
type MockDBService struct{}

func (m *MockDBService) Health() map[string]string {
	return map[string]string{"message": "Mock DB is healthy"}
}

func (m *MockDBService) Client() *mongo.Client {
	return nil // Not needed for this specific test
}

func TestHandler(t *testing.T) {
	s := &Server{}
	s.db = &MockDBService{} // Initialize s.db with the mock service
	ch := handlers.NewCommonHandler(s.db)
	server := httptest.NewServer(http.HandlerFunc(ch.HelloWorldHandler))
	defer server.Close()
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("error making request to server. Err: %v", err)
	}
	defer resp.Body.Close()
	// Assertions
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %v", resp.Status)
	}
	expected := "{\"message\":\"Hello World\"}"
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error reading response body. Err: %v", err)
	}
	if expected != string(body) {
		t.Errorf("expected response body to be %v; got %v", expected, string(body))
	}
}
