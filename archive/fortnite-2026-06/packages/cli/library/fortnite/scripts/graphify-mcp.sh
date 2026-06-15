#!/usr/bin/env bash
# graphify-mcp.sh — Build graph and/or start MCP server
# Usage:
#   ./graphify-mcp.sh build [path]     # Build knowledge graph
#   ./graphify-mcp.sh serve [path]     # Start MCP server (requires graph)
#   ./graphify-mcp.sh full [path]      # Build + serve

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
GRAPH_DIR="$ROOT_DIR/graphify-out"

usage() {
    echo "Usage: $0 {build|serve|full} [path]"
    echo "  build  — Build knowledge graph from codebase"
    echo "  serve  — Start MCP server (requires existing graph)"
    echo "  full   — Build graph then start MCP server"
    exit 1
}

cmd="${1:-}"
target="${2:-$ROOT_DIR}"

case "$cmd" in
    build)
        echo "🔨 Building knowledge graph from: $target"
        graphify "$target" --no-viz
        echo "✅ Graph built at: $GRAPH_DIR/graph.json"
        echo "   Stats: $(python3 -c "import json; g=json.load(open('$GRAPH_DIR/graph.json')); print(f'{len(g.get(\"nodes\",[]))} nodes, {len(g.get(\"edges\",[]))} edges')" 2>/dev/null || echo "parse failed")"
        ;;
    serve)
        if [ ! -f "$GRAPH_DIR/graph.json" ]; then
            echo "❌ No graph found. Run '$0 build' first."
            exit 1
        fi
        echo "🚀 Starting graphify MCP server..."
        echo "   Graph: $GRAPH_DIR/graph.json"
        python3 -m graphify.serve "$GRAPH_DIR/graph.json"
        ;;
    full)
        echo "🔨 Building knowledge graph..."
        graphify "$target" --no-viz
        echo "🚀 Starting MCP server..."
        python3 -m graphify.serve "$GRAPH_DIR/graph.json"
        ;;
    *)
        usage
        ;;
esac
