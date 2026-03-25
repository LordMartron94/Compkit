// Package transform provides a small, explicit dispatcher-based API for transforming
// Syntaxa's LST (lossless syntax tree) into a client-defined structure (commonly an AST).
//
// The core concept is a dispatcher that routes by node kind. Callers register one handler
// per kind, then drive transformation by calling the context transform function on a root
// node. Handlers can recursively transform children by re-entering the context.
//
// This package intentionally exposes a stable public API for the internal dispatch
// mechanism while keeping the implementation encapsulated.
package transform
