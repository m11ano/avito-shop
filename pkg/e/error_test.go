package e

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewError(t *testing.T) {
	err := NewError(400, "Bad Request", []string{"missing field", "invalid format"})

	assert.Equal(t, 400, err.Code(), "Code should be 400")
	assert.Equal(t, "Bad Request", err.Message(), "Message should be 'Bad Request'")

	details := err.Details()
	assert.Len(t, details, 2, "There should be 2 detail items")
	assert.Equal(t, "missing field", details[0], "First detail should be 'missing field'")
	assert.Equal(t, "invalid format", details[1], "Second detail should be 'invalid format'")

	expectedErrMsg := "Bad Request - details: [missing field invalid format]"
	assert.Equal(t, expectedErrMsg, err.Error(), "Error() output mismatch")
}

func TestIsAppError(t *testing.T) {
	err := NewError(500, "Internal Error", nil)
	assert.True(t, IsAppError(err), "IsAppError should return true for LogicError")

	stdErr := errors.New("just an error")
	assert.False(t, IsAppError(stdErr), "IsAppError should return false for standard error")
}

func TestNewErrorFrom(t *testing.T) {
	parent := NewError(401, "Unauthorized", []string{"token expired"})
	child := NewErrorFrom(parent)

	assert.Equal(t, parent.Code(), child.Code(), "Codes should match")
	assert.Equal(t, parent.Message(), child.Message(), "Messages should match")
	assert.Equal(t, parent.Details(), child.Details(), "Details should match")

	unwrapped := child.Unwrap()
	assert.Len(t, unwrapped, 1, "Unwrap should contain exactly one error")
	assert.Equal(t, parent, unwrapped[0], "Child should wrap parent error")
}

func TestWrapAndUnwrap(t *testing.T) {
	err := NewError(404, "Not Found", nil)
	wrapped := errors.New("low-level error")
	//nolint:errcheck
	err.Wrap(wrapped)

	unwrapped := err.Unwrap()
	assert.Len(t, unwrapped, 1, "Should have exactly 1 wrapped error")
	assert.Equal(t, wrapped, unwrapped[0], "Unwrapped error should match the original wrapped error")
}

func TestIs(t *testing.T) {
	err := NewError(400, "Some Error", nil)
	wrapped1 := errors.New("wrapped1")
	wrapped2 := errors.New("wrapped2")
	//nolint:errcheck
	err.Wrap(wrapped1)
	//nolint:errcheck
	err.Wrap(wrapped2)

	assert.True(t, err.Is(err), "err.Is(err) should be true")
	assert.True(t, err.Is(wrapped1), "err should 'Is' wrapped1")
	assert.True(t, err.Is(wrapped2), "err should 'Is' wrapped2")

	unknown := errors.New("unknown")
	assert.False(t, err.Is(unknown), "err should not 'Is' unknown")
}

func TestLock(t *testing.T) {
	err := NewError(403, "Forbidden", []string{"no access"})
	//nolint:errcheck
	err.Lock()

	//nolint:errcheck
	err.SetMessage("New message").AddDetails([]string{"new detail"}).SetData("testdata")

	assert.Equal(t, "Forbidden", err.Message(), "Message should not change after Lock()")
	assert.Equal(t, []string{"no access"}, err.Details(), "Details should remain unchanged after Lock()")
	assert.Nil(t, err.Data(), "Data should remain nil after Lock()")

	wrapped := errors.New("wrap attempt")
	//nolint:errcheck
	err.Wrap(wrapped)
	assert.Empty(t, err.Unwrap(), "No errors should be wrapped after Lock()")
}

func TestSetMessage(t *testing.T) {
	err := NewError(200, "OK", nil)
	//nolint:errcheck
	err.SetMessage("All Good")
	assert.Equal(t, "All Good", err.Message(), "Message should be updated to 'All Good'")
}

func TestAddDetails(t *testing.T) {
	err := NewError(200, "OK", []string{"initial"})
	//nolint:errcheck
	err.AddDetails([]string{"second"})

	details := err.Details()
	assert.Len(t, details, 2, "Details should have 2 items")
	assert.Equal(t, "initial", details[0], "First detail should be 'initial'")
	assert.Equal(t, "second", details[1], "Second detail should be 'second'")
}

func TestSetData(t *testing.T) {
	err := NewError(200, "OK", nil)
	//nolint:errcheck
	err.SetData(map[string]any{"key": "value"})

	data, ok := err.Data().(map[string]any)
	assert.True(t, ok, "Data should be a map[string]any")
	assert.Equal(t, "value", data["key"], "Data[\"key\"] should be 'value'")
}
