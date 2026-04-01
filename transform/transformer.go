package transform

import (

	"compkit/internal"
	"compkit/validation"
	"syntaxa"
)

/*
LSTNodeTransformHandler is a per-node-kind callback used by the transform dispatcher.

The handler receives:
- ctx: The transform context, used to recursively transform child nodes via dispatch.
- node: The current LST node being transformed.

The handler returns the produced AST node (or other client structure).

Any errors should be reported through the context (validation entries), and the handler
should return the zero value of TAST when it cannot produce a meaningful result.
*/
type LSTNodeTransformHandler[TNodeKind comparable, TAST any] func(
	ctx LSTNodeTransformContext[TNodeKind, TAST],
	node *syntaxa.SyntaxaLSTNode[TNodeKind],
) TAST

/*
LSTNodeTransformDispatcher routes LST nodes to user-provided handlers based on node kind.

This is a thin public wrapper over Compkit's internal dispatcher mechanism. The wrapper:
- Keeps the internal representation private (stable API boundary).
- Presents an explicit, C-style function API (no method receivers for core logic).
*/
type LSTNodeTransformDispatcher[TNodeKind comparable, TAST any] struct {
	dispatcher *internal.NodeTransformDispatcher[TNodeKind, TAST]
}

/*
LSTNodeTransformContext is the execution context for transforming an LST node tree.

The context is intentionally lightweight and primarily serves as a recursion entry point
for calling back into the dispatcher.
*/
type LSTNodeTransformContext[TNodeKind comparable, TAST any] struct {
	ctx *internal.LSTNodeTransformContext[TNodeKind, TAST]
}

/*
LSTNodeTransformDispatcherCreate creates an empty dispatcher.

Use cases:
- Building compiler frontends where an LST must be transformed into a typed AST
- Validating or normalizing Syntaxa trees into domain-neutral intermediate structures

Time complexity: O(1)
Space complexity: O(1) (initial map allocation is implementation-defined)
*/
func LSTNodeTransformDispatcherCreate[TNodeKind comparable, TAST any]() *LSTNodeTransformDispatcher[TNodeKind, TAST] {
	return &LSTNodeTransformDispatcher[TNodeKind, TAST]{
		dispatcher: internal.NodeTransformDispatcherCreate[TNodeKind, TAST](),
	}
}

/*
LSTNodeTransformDispatcherRegister registers a handler for a specific node kind.

Prerequisites:
- dispatcher must be non-nil
- handler must be non-nil
- dispatcher must not have been marked populated
- a handler for kind must not already exist

Edge cases:
- Panics if prerequisites are violated

Time complexity: O(1) average
Space complexity: O(1) additional (excluding map growth)
*/
func LSTNodeTransformDispatcherRegister[TNodeKind comparable, TAST any](
	dispatcher *LSTNodeTransformDispatcher[TNodeKind, TAST],
	kind TNodeKind,
	handler LSTNodeTransformHandler[TNodeKind, TAST],
) {
	internal.NodeTransformDispatcherRegister(
		dispatcher.dispatcher,
		kind,
		func(
			ctx *internal.LSTNodeTransformContext[TNodeKind, TAST],
			node *syntaxa.SyntaxaLSTNode[TNodeKind],
		) TAST {
			return handler(LSTNodeTransformContext[TNodeKind, TAST]{ctx: ctx}, node)
		},
	)
}

/*
LSTNodeTransformDispatcherMarkPopulated locks the dispatcher against future registrations.

This is meant to enforce a build-then-run workflow: populate all handlers once during
initialization, then treat the dispatcher as immutable during transformation.

Prerequisites:
- dispatcher must be non-nil

Edge cases:
- Panics if dispatcher is nil

Time complexity: O(1)
Space complexity: O(1)
*/
func LSTNodeTransformDispatcherMarkPopulated[TNodeKind comparable, TAST any](
	dispatcher *LSTNodeTransformDispatcher[TNodeKind, TAST],
) {
	internal.NodeTransformDispatcherMarkPopulated(dispatcher.dispatcher)
}

/*
LSTNodeTransformContextCreate creates a new transform context bound to dispatcher.

Prerequisites:
- dispatcher must be non-nil

Edge cases:
- Panics if dispatcher is nil

Time complexity: O(1)
Space complexity: O(1)
*/
func LSTNodeTransformContextCreate[TNodeKind comparable, TAST any](
	dispatcher *LSTNodeTransformDispatcher[TNodeKind, TAST],
	validationEntries *validation.ValidationEntries,
) LSTNodeTransformContext[TNodeKind, TAST] {
	return LSTNodeTransformContext[TNodeKind, TAST]{
		ctx: internal.LSTNodeTransformContextCreate(dispatcher.dispatcher, validationEntries),
	}
}

/*
LSTNodeTransformContextTransform transforms a single LST node by dispatching to the
registered handler for node.Kind().

Use cases:
- Implementing handlers that transform child nodes (recursion)
- Driving a full-tree transform by calling this function on the root node

Time complexity: O(1) average dispatch + handler cost
Space complexity: O(1) not counting recursion inside handlers

Prerequisites:
- ctx must reference a valid dispatcher-bound context
- node must be non-nil

Edge cases:
- Reports an ERROR validation entry if no handler exists for node.Kind()
- Reports a FATAL validation entry if node is nil
- Returns the zero value of TAST in error cases
*/
func LSTNodeTransformContextTransform[TNodeKind comparable, TAST any](
	ctx LSTNodeTransformContext[TNodeKind, TAST],
	node *syntaxa.SyntaxaLSTNode[TNodeKind],
) TAST {
	return internal.LSTNodeTransformContextTransform(ctx.ctx, node)
}

/*
LSTNodeTransformContextIsValid reports whether ctx can be used for transformation.

Time complexity: O(1)
Space complexity: O(1)
*/
func LSTNodeTransformContextIsValid[TNodeKind comparable, TAST any](
	ctx LSTNodeTransformContext[TNodeKind, TAST],
) bool {
	return ctx.ctx != nil
}

/*
LSTNodeTransformContextReport adds a validation entry at node.FullSpan().

Use cases:
- Reporting recoverable transform problems without returning a fatal error

Time complexity: O(1)
Space complexity: O(1) additional (excluding collector growth)
*/
func LSTNodeTransformContextReport[TNodeKind comparable, TAST any](
	ctx LSTNodeTransformContext[TNodeKind, TAST],
	node *syntaxa.SyntaxaLSTNode[TNodeKind],
	severity validation.ValidationSeverity,
	code validation.ValidationCode,
	message string,
	note *string,
) {
	internal.LSTNodeTransformContextReportError(ctx.ctx, node, severity, code, message, note)
}

/*
LSTNodeTransformContextReportDiagnostic reports a DIAGNOSTIC entry.
*/
func LSTNodeTransformContextReportDiagnostic[TNodeKind comparable, TAST any](
	ctx LSTNodeTransformContext[TNodeKind, TAST],
	node *syntaxa.SyntaxaLSTNode[TNodeKind],
	code validation.ValidationCode,
	message string,
	note *string,
) {
	LSTNodeTransformContextReport(ctx, node, validation.VALIDATION_DIAGNOSTIC, code, message, note)
}

/*
LSTNodeTransformContextReportInfo reports an INFO entry.
*/
func LSTNodeTransformContextReportInfo[TNodeKind comparable, TAST any](
	ctx LSTNodeTransformContext[TNodeKind, TAST],
	node *syntaxa.SyntaxaLSTNode[TNodeKind],
	code validation.ValidationCode,
	message string,
	note *string,
) {
	LSTNodeTransformContextReport(ctx, node, validation.VALIDATION_INFO, code, message, note)
}

/*
LSTNodeTransformContextReportNotice reports a NOTICE entry.
*/
func LSTNodeTransformContextReportNotice[TNodeKind comparable, TAST any](
	ctx LSTNodeTransformContext[TNodeKind, TAST],
	node *syntaxa.SyntaxaLSTNode[TNodeKind],
	code validation.ValidationCode,
	message string,
	note *string,
) {
	LSTNodeTransformContextReport(ctx, node, validation.VALIDATION_NOTICE, code, message, note)
}

/*
LSTNodeTransformContextReportWarning reports a WARNING entry.
*/
func LSTNodeTransformContextReportWarning[TNodeKind comparable, TAST any](
	ctx LSTNodeTransformContext[TNodeKind, TAST],
	node *syntaxa.SyntaxaLSTNode[TNodeKind],
	code validation.ValidationCode,
	message string,
	note *string,
) {
	LSTNodeTransformContextReport(ctx, node, validation.VALIDATION_WARNING, code, message, note)
}

/*
LSTNodeTransformContextReportError reports an ERROR entry.
*/
func LSTNodeTransformContextReportError[TNodeKind comparable, TAST any](
	ctx LSTNodeTransformContext[TNodeKind, TAST],
	node *syntaxa.SyntaxaLSTNode[TNodeKind],
	code validation.ValidationCode,
	message string,
	note *string,
) {
	LSTNodeTransformContextReport(ctx, node, validation.VALIDATION_ERROR, code, message, note)
}

/*
LSTNodeTransformContextReportFatal reports a FATAL entry.
*/
func LSTNodeTransformContextReportFatal[TNodeKind comparable, TAST any](
	ctx LSTNodeTransformContext[TNodeKind, TAST],
	node *syntaxa.SyntaxaLSTNode[TNodeKind],
	code validation.ValidationCode,
	message string,
	note *string,
) {
	LSTNodeTransformContextReport(ctx, node, validation.VALIDATION_FATAL, code, message, note)
}
