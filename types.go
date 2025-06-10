package deepclone

// Cloneable interface allows types to implement custom deep cloning behavior.
// Types implementing this interface will have their Clone method called
// instead of using the default reflection-based cloning.
//
// The Clone method should return a deep copy of the receiver.
// It's the implementer's responsibility to ensure all nested data
// is properly cloned to maintain complete independence from the original.
//
// Example:
//
//	type Document struct {
//	    Title   string
//	    Content []byte
//	    Metadata map[string]interface{}
//	}
//
//	func (d Document) Clone() any {
//	    return Document{
//	        Title:   d.Title,
//	        Content: deepclone.Clone(d.Content).([]byte),
//	        Metadata: deepclone.Clone(d.Metadata).(map[string]interface{}),
//	    }
//	}
type Cloneable interface {
	Clone() any
}
