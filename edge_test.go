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

func TestClonePreservesMapKeysPointingToValueFields(t *testing.T) {
	t.Parallel()
	type node struct {
		Value int
	}
	type holder struct {
		Node node
	}

	t.Run("key points to value field", func(t *testing.T) {
		t.Parallel()
		originalHolder := &holder{Node: node{Value: 7}}
		original := map[*node]*holder{
			&originalHolder.Node: originalHolder,
		}

		cloned := Clone(original)

		require.Len(t, cloned, 1)
		var clonedKey *node
		var clonedHolder *holder
		for key, value := range cloned {
			clonedKey = key
			clonedHolder = value
		}

		require.NotNil(t, clonedKey)
		require.NotNil(t, clonedHolder)
		assert.False(t, clonedHolder == originalHolder)
		assert.True(t, clonedKey == &clonedHolder.Node, "map key should point at the cloned value field")

		clonedKey.Value = 11
		assert.Equal(t, 11, clonedHolder.Node.Value)
		assert.Equal(t, 7, originalHolder.Node.Value)
	})

	t.Run("struct key points to value field", func(t *testing.T) {
		t.Parallel()
		type key struct {
			Ref *node
		}

		originalHolder := &holder{Node: node{Value: 7}}
		original := map[key]*holder{
			{Ref: &originalHolder.Node}: originalHolder,
		}

		cloned := Clone(original)

		require.Len(t, cloned, 1)
		var clonedKey key
		var clonedHolder *holder
		for key, value := range cloned {
			clonedKey = key
			clonedHolder = value
		}

		require.NotNil(t, clonedKey.Ref)
		require.NotNil(t, clonedHolder)
		assert.True(t, clonedKey.Ref == &clonedHolder.Node, "struct map key should point at the cloned value field")

		clonedKey.Ref.Value = 11
		assert.Equal(t, 11, clonedHolder.Node.Value)
		assert.Equal(t, 7, originalHolder.Node.Value)
	})

	t.Run("interface key points to value field", func(t *testing.T) {
		t.Parallel()
		originalHolder := &holder{Node: node{Value: 7}}
		original := map[any]*holder{
			&originalHolder.Node: originalHolder,
		}

		cloned := Clone(original)

		require.Len(t, cloned, 1)
		var clonedKey *node
		var clonedHolder *holder
		for key, value := range cloned {
			clonedKey = key.(*node)
			clonedHolder = value
		}

		require.NotNil(t, clonedKey)
		require.NotNil(t, clonedHolder)
		assert.True(t, clonedKey == &clonedHolder.Node, "interface map key should point at the cloned value field")

		clonedKey.Value = 11
		assert.Equal(t, 11, clonedHolder.Node.Value)
		assert.Equal(t, 7, originalHolder.Node.Value)
	})
}
