package deepclone

// Cloneable allows types to implement custom deep cloning behavior.
// Types implementing this interface will have their Clone method called
// instead of using the default reflection-based cloning.
//
// The Clone method should return a deep copy of the receiver.
// The implementer must ensure all nested data is properly deep copied
// to maintain complete independence from the original. Note that the
// library's built-in circular reference detection does not apply inside
// a custom Clone method; handle cycles manually if needed.
//
// Example:
//
//	type Document struct {
//	    Title    string
//	    Content  []byte
//	    Metadata map[string]any
//	}
//
//	func (d Document) Clone() any {
//	    return Document{
//	        Title:    d.Title,
//	        Content:  deepclone.Clone(d.Content),
//	        Metadata: deepclone.Clone(d.Metadata),
//	    }
//	}
type Cloneable interface {
	Clone() any
}
