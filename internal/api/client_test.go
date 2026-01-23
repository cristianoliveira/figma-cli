package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ClientTestSuite struct {
	suite.Suite
	mockClient *MockClient
}

func (s *ClientTestSuite) SetupTest() {
	s.mockClient = &MockClient{}
}

func (s *ClientTestSuite) TestGetFile() {
	ctx := context.Background()
	expectedFile := &File{Key: "test", Name: "Test File"}
	s.mockClient.On("GetFile", ctx, "test").Return(expectedFile, nil)

	file, err := s.mockClient.GetFile(ctx, "test")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expectedFile, file)
	s.mockClient.AssertExpectations(s.T())
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
