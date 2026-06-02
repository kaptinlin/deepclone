package deepclone

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClonePreservesPointerToStructField(t *testing.T) {
	t.Parallel()
	type node struct {
		Value int
	}

	t.Run("field before pointer", func(t *testing.T) {
		t.Parallel()
		type holder struct {
			Node node
			Ref  *node
		}

		original := &holder{Node: node{Value: 7}}
		original.Ref = &original.Node

		cloned := Clone(original)

		require.NotNil(t, cloned)
		require.NotNil(t, cloned.Ref)
		assert.False(t, original == cloned)
		assert.True(t, cloned.Ref == &cloned.Node, "pointer to field should point at the cloned field")

		cloned.Ref.Value = 11
		assert.Equal(t, 11, cloned.Node.Value)
		assert.Equal(t, 7, original.Node.Value)
	})

	t.Run("pointer before field", func(t *testing.T) {
		t.Parallel()
		type holder struct {
			Ref  *node
			Node node
		}

		original := &holder{Node: node{Value: 7}}
		original.Ref = &original.Node

		cloned := Clone(original)

		require.NotNil(t, cloned)
		require.NotNil(t, cloned.Ref)
		assert.False(t, original == cloned)
		assert.True(t, cloned.Ref == &cloned.Node, "pointer to field should point at the cloned field")

		cloned.Ref.Value = 11
		assert.Equal(t, 11, cloned.Node.Value)
		assert.Equal(t, 7, original.Node.Value)
	})

	t.Run("pointer before nested field", func(t *testing.T) {
		t.Parallel()
		type wrapper struct {
			Node node
		}
		type holder struct {
			Ref     *node
			Wrapper wrapper
		}

		original := &holder{Wrapper: wrapper{Node: node{Value: 7}}}
		original.Ref = &original.Wrapper.Node

		cloned := Clone(original)

		require.NotNil(t, cloned)
		require.NotNil(t, cloned.Ref)
		assert.False(t, original == cloned)
		assert.True(t, cloned.Ref == &cloned.Wrapper.Node, "pointer to nested field should point at the cloned field")

		cloned.Ref.Value = 11
		assert.Equal(t, 11, cloned.Wrapper.Node.Value)
		assert.Equal(t, 7, original.Wrapper.Node.Value)
	})
}
