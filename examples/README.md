# DeepClone Examples

This directory contains various examples demonstrating different use cases of the DeepClone library.

## Available Examples

### 1. Basic Examples (`basic/`)
Demonstrates fundamental cloning operations:
- Primitive types (int, string, bool)
- Collections (slices, maps)
- Pointers and pointer chains
- Complex nested structures
- Custom cloning with Cloneable interface

**Run:**
```bash
cd basic && go run main.go
```

### 2. Circular Reference Examples (`circular/`)
Shows how DeepClone handles circular references safely:
- Simple circular references between nodes
- Self-referencing structures
- Safe cloning without infinite loops

**Run:**
```bash
cd circular && go run main.go
```

### 3. Custom Cloning Examples (`custom/`)
Demonstrates custom cloning behavior via the Cloneable interface:
- Counters with custom increment logic
- Comparison with default deep cloning
- Custom vs automatic cloning behavior

**Run:**
```bash
cd custom && go run main.go
```

## Running All Examples

To run all examples at once:

```bash
# From the examples directory
for dir in */; do
    if [ -f "$dir/main.go" ]; then
        echo "=== Running $dir ==="
        cd "$dir" && go run main.go && cd ..
        echo ""
    fi
done
```

## Key Learning Points

1. **Zero-allocation paths**: Primitive types are cloned without memory allocation
2. **Deep independence**: All cloned data is completely independent from the original
3. **Circular safety**: Circular references are handled without infinite loops
4. **Custom behavior**: Implement Cloneable interface for specialized cloning logic

## Requirements

- Go 1.21 or later
- DeepClone library installed (`go get github.com/kaptinlin/deepclone`)