#!/bin/bash

echo "🚀 Running Go Deep Clone Performance Benchmarks"
echo "=============================================="
echo ""

# Change to benchmarks directory
cd "$(dirname "$0")"

echo "📊 Environment:"
echo "Platform: $(go env GOOS)/$(go env GOARCH)"
echo "Go Version: $(go version | cut -d' ' -f3)"
echo "CPU: $(sysctl -n machdep.cpu.brand_string 2>/dev/null || lscpu | grep 'Model name' | cut -d: -f2 | xargs)"
echo ""

echo "⏱️  Running benchmarks..."
echo ""

# Run benchmarks with nice formatting
go test -bench=. -benchmem -benchtime=1s | tee benchmark_results.txt

echo ""
echo "✅ Benchmarks completed!"
echo "📝 Results saved to benchmark_results.txt"