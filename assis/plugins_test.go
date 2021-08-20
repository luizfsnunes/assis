package assis_test

import (
	"github.com/luizfsnunes/assis/assis"
	"github.com/luizfsnunes/assis/assis/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPluginDispatcher_DispatchPluginGeneratedFiles(t *testing.T) {
	mockGenerateFiles := &mocks.PluginGeneratedFiles{}
	mockGenerateFiles.
		On("AfterGeneratedFiles", []string{"a", "b", "c"}).
		Return(nil).
		Once()

	registry := assis.NewPluginRegistry(mockGenerateFiles)
	dispatcher := assis.NewPluginDispatcher(registry)

	err := dispatcher.DispatchPluginGeneratedFiles([]string{"a", "b", "c"})
	assert.NoError(t, err)
}
