package internal

import (
	"fmt"

	"compkit/validation"
	"syntaxa"
)

// --------------------------------------------------------------- DISPATCHER

type TransformHandler[TNodeKind comparable, TAST any] func(
	ctx *LSTNodeTransformContext[TNodeKind, TAST],
	node *syntaxa.SyntaxaLSTNode[TNodeKind],
) TAST

type NodeTransformDispatcher[TNodeKind comparable, TAST any] struct {
	handlers           map[TNodeKind]TransformHandler[TNodeKind, TAST]
	populationComplete bool
}

func NodeTransformDispatcherCreate[TNodeKind comparable, TAST any]() *NodeTransformDispatcher[TNodeKind, TAST] {
	return &NodeTransformDispatcher[TNodeKind, TAST]{
		handlers: make(map[TNodeKind]TransformHandler[TNodeKind, TAST]),
	}
}

func NodeTransformDispatcherRegister[TNodeKind comparable, TAST any](
	n *NodeTransformDispatcher[TNodeKind, TAST],
	kind TNodeKind,
	handler TransformHandler[TNodeKind, TAST],
) {
	if n == nil {
		panic("dispatcher is nil -> cannot register")
	}

	if handler == nil {
		panic("handler is nil -> cannot register")
	}

	if n.populationComplete {
		panic("dispatcher marked as populated -> cannot register anymore")
	}

	if _, exist := n.handlers[kind]; exist {
		panic(fmt.Errorf("handler for kind '%v' already exists", kind))
	}

	n.handlers[kind] = handler
}

func NodeTransformDispatcherMarkPopulated[TNodeKind comparable, TAST any](
	n *NodeTransformDispatcher[TNodeKind, TAST],
) {
	if n == nil {
		panic("dispatcher is nil -> cannot mark populated")
	}

	n.populationComplete = true
}

func nodeTransformDispatcherDispatch[TNodeKind comparable, TAST any](
	n *NodeTransformDispatcher[TNodeKind, TAST],
	ctx *LSTNodeTransformContext[TNodeKind, TAST],
	node *syntaxa.SyntaxaLSTNode[TNodeKind],
) TAST {
	var zero TAST
	if n == nil {
		if node != nil {
			LSTNodeTransformContextReportError(
				ctx,
				node,
				validation.VALIDATION_FATAL,
				"COMPKIT_TRANSFORM_DISPATCHER_NIL",
				"internal error: dispatcher is nil",
				nil,
			)
		}
		return zero
	}

	handler, exists := n.handlers[node.Kind()]

	if !exists {
		LSTNodeTransformContextReportError(
			ctx,
			node,
			validation.VALIDATION_ERROR,
			"COMPKIT_TRANSFORM_NO_HANDLER",
			fmt.Sprintf("no handler registered for kind '%v'", node.Kind()),
			nil,
		)
		return zero
	}

	return handler(ctx, node)
}

// --------------------------------------------------------------- TRANSFORMER

type LSTNodeTransformContext[TNodeKind comparable, TAST any] struct {
	dispatcher        *NodeTransformDispatcher[TNodeKind, TAST]
	validationEntries *validation.ValidationEntries
}

func LSTNodeTransformContextCreate[TNodeKind comparable, TAST any](
	dispatcher *NodeTransformDispatcher[TNodeKind, TAST],
	validationEntries *validation.ValidationEntries,
) *LSTNodeTransformContext[TNodeKind, TAST] {
	return &LSTNodeTransformContext[TNodeKind, TAST]{
		dispatcher:        dispatcher,
		validationEntries: validationEntries,
	}
}

func LSTNodeTransformContextTransform[TNodeKind comparable, TAST any](
	l *LSTNodeTransformContext[TNodeKind, TAST],
	node *syntaxa.SyntaxaLSTNode[TNodeKind],
) TAST {
	var zero TAST

	if l == nil {
		panic("context is nil")
	}

	if l.validationEntries == nil {
		panic("validation entries is nil")
	}

	if node == nil {
		validation.ValidationEntriesAddFatal(
			l.validationEntries,
			"COMPKIT_TRANSFORM_NODE_NIL",
			"engine error: transforming nil node",
			nil,
			syntaxa.Span{},
		)
		return zero
	}

	return nodeTransformDispatcherDispatch(l.dispatcher, l, node)
}

func LSTNodeTransformContextReportError[TNodeKind comparable, TAST any](
	l *LSTNodeTransformContext[TNodeKind, TAST],
	node *syntaxa.SyntaxaLSTNode[TNodeKind],
	severity validation.ValidationSeverity,
	code validation.ValidationCode,
	message string,
	note *string,
) {
	if l == nil {
		panic("context is nil -> cannot report validation entry")
	}
	if l.validationEntries == nil {
		panic("validation entries is nil -> cannot report validation entry")
	}

	if node == nil {
		panic("node is nil -> cannot report validation entry")
	}

	validation.ValidationEntriesAdd(l.validationEntries, severity, code, message, note, node.FullSpan())
}
