package runware

import (
	"fmt"
	"testing"
	
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MockRunware struct {
	APIKeyFunc        func() string
	ConnectedFunc     func() bool
	CloseFunc         func() error
	SendFunc          func([]byte) error
	ListenFunc        func() chan []byte
	ReconnectedFunc   func() chan struct{}
	ReconnectedCalled bool
	Conn              *websocket.Conn
}

func (m *MockRunware) APIKey() string {
	if m.APIKeyFunc != nil {
		return m.APIKeyFunc()
	}
	return ""
}

func (m *MockRunware) Connected() bool {
	if m.ConnectedFunc != nil {
		return m.ConnectedFunc()
	}
	return false
}

func (m *MockRunware) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func (m *MockRunware) Send(b []byte) error {
	if m.SendFunc != nil {
		return m.SendFunc(b)
	}
	return nil
}

func (m *MockRunware) Listen() chan []byte {
	if m.ListenFunc != nil {
		return m.ListenFunc()
	}
	return nil
}

func (m *MockRunware) Reconnected() chan struct{} {
	if m.ReconnectedFunc != nil {
		m.ReconnectedCalled = true
		return m.ReconnectedFunc()
	}
	return nil
}

type SDKTestSuite struct {
	suite.Suite
	service SDK
}

func (s *SDKTestSuite) SetupTest() {
	mClient := &MockRunware{
		APIKeyFunc: func() string {
			return "test-api-key"
		},
	}
	
	s.service = SDK{
		Client: mClient,
	}
	
}

func (s *SDKTestSuite) Test_SDKInit() {
	s.Run("SDK Is set correctly", func() {
		assert.Equal(s.T(), "test-api-key", s.service.Client.APIKey())
		
	})
	
	s.Run("SDK API key missing", func() {
		mClient := &MockRunware{
			APIKeyFunc: func() string {
				return ""
			},
		}
		sdk := SDK{
			Client: mClient,
		}
		assert.Equal(s.T(), "", sdk.Client.APIKey())
	})
}

func (s *SDKTestSuite) Test_ErrorMap() {
	testCases := []struct {
		name             string
		msg              map[string]interface{}
		expectedErr      error
		expectedHasError bool
	}{
		{
			name:             "Error True with Known Error ID",
			msg:              map[string]interface{}{"error": true, "errorId": float64(19), "errorMessage": "Invalid API key"},
			expectedErr:      fmt.Errorf("%w:[19:Invalid API key]", ErrInvalidApiKey),
			expectedHasError: true,
		},
		{
			name:             "Error True with Unknown Error ID",
			msg:              map[string]interface{}{"error": true, "errorId": float64(999), "errorMessage": "Unknown error"},
			expectedErr:      fmt.Errorf("%w:[999:Unknown error]", ErrWsUnknownError),
			expectedHasError: true,
		},
		{
			name:             "Error False",
			msg:              map[string]interface{}{"error": false},
			expectedErr:      nil,
			expectedHasError: false,
		},
		{
			name:             "No Error Key",
			msg:              map[string]interface{}{"errorId": 19, "errorMessage": "Invalid API key"},
			expectedErr:      nil,
			expectedHasError: false,
		},
		{
			name:             "Error Key Not Boolean",
			msg:              map[string]interface{}{"error": "not a bool", "errorId": float64(19), "errorMessage": "Invalid API key"},
			expectedErr:      nil,
			expectedHasError: false,
		},
		{
			name:             "Error ID Not Float64",
			msg:              map[string]interface{}{"error": true, "errorId": "not a float64", "errorMessage": "Invalid API key"},
			expectedErr:      fmt.Errorf("%w:[not a float64:Invalid API key]", ErrWsUnknownError),
			expectedHasError: true,
		},
		{
			name:             "Empty Map",
			msg:              map[string]interface{}{},
			expectedErr:      nil,
			expectedHasError: false,
		},
	}
	
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			err, hasError := s.service.OnError(tc.msg)
			assert.Equal(s.T(), tc.expectedErr, err)
			assert.Equal(s.T(), tc.expectedHasError, hasError)
		})
	}
}

func (s *SDKTestSuite) TearDownTest() {
	fmt.Println("Tear down")
}

func Test_SDK(t *testing.T) {
	suite.Run(t, new(SDKTestSuite))
}
