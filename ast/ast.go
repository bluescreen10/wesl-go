package ast

// Node is the base interface for every element in the AST.
// All concrete node types implement this interface through marker methods.
type Node interface {
	node()
}

// IfAttr is a generic conditional node that selects between two branches
// based on a compile-time @if attribute condition. T is the kind of node
// held in the Then/Else branches (Decl, Stmt, Param, etc.).
type IfAttr[T any] struct {
	Cond Expr // compile-time boolean expression that selects the branch
	Then T    // branch taken when Cond evaluates to true
	Else T    // branch taken when Cond evaluates to false
}

// ----------------------------------------------------------------------------
// Decls
type (
	// Decl is the interface implemented by every top-level declaration node.
	Decl interface {
		Node
		declNode()
	}

	// DiagnosticDirective represents a WGSL diagnostic directive that controls
	// how the shader compiler reports diagnostics for a given rule.
	DiagnosticDirective struct {
		Attrs   []*Attribute      // attributes applied to this directive
		Control DiagnosticControl // severity and rule name for the diagnostic
	}

	// DiagnosticControl holds the severity level and rule name for a diagnostic
	// directive.
	DiagnosticControl struct {
		Severity string // diagnostic severity (e.g. "error", "warning", "info", "off")
		RuleName string // qualified rule name (e.g. "derivative_uniformity")
	}

	// EnableDirective represents a WGSL enable directive that activates one or
	// more language extensions.
	EnableDirective struct {
		Attrs      []*Attribute // attributes applied to this directive
		Extensions []string     // names of the extensions to enable
	}

	// FuncDecl represents a function declaration, including its name, parameter
	// list, optional return type, and body.
	FuncDecl struct {
		Name        *Ident         // function name
		Attrs       []*Attribute   // attributes applied to the function (e.g. @vertex)
		Params      []Param        // ordered list of parameters
		ReturnAttrs []*Attribute   // attributes applied to the return type (e.g. @builtin)
		ReturnType  *TypeSpecifier // return type; nil when the function returns nothing
		Body        *BlockStmt     // function body
	}

	// Param is the interface implemented by all function parameter nodes,
	// including conditional @if parameters.
	Param interface {
		Node
		paramNode()
	}

	// FuncParam represents a single named, typed function parameter.
	FuncParam struct {
		Name  string         // parameter name
		Type  *TypeSpecifier // declared parameter type
		Attrs []*Attribute   // attributes applied to the parameter (e.g. @builtin)
	}

	// IfAttrParam is a conditional function parameter whose presence is
	// controlled by an @if attribute evaluated at compile time.
	IfAttrParam IfAttr[Param]

	// IfAttrDecl is a conditional top-level declaration whose presence is
	// controlled by an @if attribute evaluated at compile time.
	IfAttrDecl IfAttr[Decl]

	// ImportDecl groups one or more import statements that bring external
	// symbols into scope.
	ImportDecl struct {
		Imports []ImportedItem // individual imports collected in this declaration
	}

	// ImportedItem is a single fully-qualified import with an optional alias.
	// Path holds every segment including the leading anchor (package/super) and
	// the final symbol name. For example: import package::foo::bar as b
	// → ImportedItem{Path: ["package","foo","bar"], Alias: "b"}
	ImportedItem struct {
		Path  []string // fully-qualified path segments including the anchor and symbol name
		Alias string   // local alias for the imported symbol; empty means use the symbol name
	}

	// RequiresDirective represents a WGSL requires directive that asserts the
	// presence of one or more language features.
	RequiresDirective struct {
		Attrs      []*Attribute // attributes applied to this directive
		Extensions []string     // names of the required features
	}

	// StructDecl represents a struct type declaration with its name, optional
	// attributes, and ordered list of members.
	StructDecl struct {
		Name    *Ident       // struct type name
		Attrs   []*Attribute // attributes applied to the struct
		Members []Member     // ordered list of struct members
	}

	// Member is the interface implemented by all struct member nodes, including
	// conditional @if members.
	Member interface {
		Node
		structMemberNode()
	}

	// StructMember represents a single named, typed field inside a struct.
	StructMember struct {
		Name  string         // field name
		Attrs []*Attribute   // attributes applied to the field (e.g. @builtin, @location)
		Type  *TypeSpecifier // declared field type
	}

	// IfAttrStructMember is a conditional struct field whose presence is
	// controlled by an @if attribute evaluated at compile time.
	IfAttrStructMember IfAttr[Member]

	// TypeAliasDecl represents a type alias declaration that binds a new name
	// to an existing type.
	TypeAliasDecl struct {
		Name  *Ident         // alias name being declared
		Attrs []*Attribute   // attributes applied to the alias
		Type  *TypeSpecifier // the underlying type this alias refers to
	}
)

func (*DiagnosticDirective) declNode() {}
func (*EnableDirective) declNode()     {}
func (*FuncDecl) declNode()            {}
func (*ImportDecl) declNode()          {}
func (*IfAttrDecl) declNode()          {}
func (*RequiresDirective) declNode()   {}
func (*StructDecl) declNode()          {}
func (*TypeAliasDecl) declNode()       {}
func (*VarStmt) declNode()             {}
func (*ValStmt) declNode()             {}
func (*ConstAssertStmt) declNode()     {}

func (*DiagnosticDirective) node() {}
func (*EnableDirective) node()     {}
func (*FuncDecl) node()            {}
func (*ImportDecl) node()          {}
func (*IfAttrDecl) node()          {}
func (*RequiresDirective) node()   {}
func (*StructDecl) node()          {}
func (*TypeAliasDecl) node()       {}

func (*IfAttrStructMember) structMemberNode() {}
func (*StructMember) structMemberNode()       {}

func (*IfAttrStructMember) node() {}
func (*StructMember) node()       {}

func (*FuncParam) paramNode()   {}
func (*IfAttrParam) paramNode() {}

func (*FuncParam) node()   {}
func (*IfAttrParam) node() {}

// ----------------------------------------------------------------------------
// Stmt
type (
	// Stmt is the interface implemented by every statement node.
	Stmt interface {
		Node
		stmtNode()
	}

	// AssignmentStmt represents a simple or compound assignment (e.g. x = y,
	// x += y). A nil LHS stands for the phony assignment target _.
	AssignmentStmt struct {
		Attrs []*Attribute // attributes applied to the statement
		LHS   Expr         // left-hand side of the assignment; nil for phony target
		RHS   Expr         // right-hand side value being assigned
		Op    string       // assignment operator, e.g. "=", "+=", "-="
	}

	// BreakStmt represents a break statement that exits the nearest enclosing
	// loop. When Cond is non-nil it is a "break if" conditional break.
	BreakStmt struct {
		Attrs []*Attribute // attributes applied to the statement
		Cond  Expr         // optional condition for "break if"; nil for unconditional break
	}

	// ConstAssertStmt represents a compile-time assertion that the given
	// expression evaluates to true.
	ConstAssertStmt struct {
		Attrs []*Attribute // attributes applied to the statement
		Expr  Expr         // boolean expression that must be true at compile time
	}

	// ContinueStmt represents a continue statement that skips the rest of the
	// current loop iteration.
	ContinueStmt struct {
		Attrs []*Attribute // attributes applied to the statement
	}

	// ContinuingStmt represents a continuing block that runs at the end of each
	// loop iteration.
	ContinuingStmt struct {
		Attrs []*Attribute // attributes applied to the statement
		Body  *BlockStmt   // statements executed at the end of each iteration
	}

	// BlockStmt represents a brace-enclosed sequence of statements.
	BlockStmt struct {
		Attrs []*Attribute // attributes applied to the block
		Stmts []Stmt       // ordered list of statements in the block
	}

	// DiscardStmt represents a discard statement that terminates the current
	// fragment shader invocation.
	DiscardStmt struct {
		Attrs []*Attribute // attributes applied to the statement
	}

	// EmptyStmt represents a no-op statement (a bare semicolon).
	EmptyStmt struct{}

	// ForStmt represents a for-loop with an optional initializer, condition, and
	// update statement.
	ForStmt struct {
		Attrs  []*Attribute // attributes applied to the loop
		Init   Stmt         // optional initializer executed before the first iteration; nil if absent
		Cond   Expr         // optional loop condition evaluated before each iteration; nil means infinite
		Update Stmt         // optional statement executed after each iteration; nil if absent
		Body   *BlockStmt   // loop body
	}

	// FuncCallStmt represents a function call used as a statement (the return
	// value, if any, is discarded).
	FuncCallStmt struct {
		Attrs []*Attribute // attributes applied to the statement
		Call  *CallExpr    // the function call expression being executed
	}

	// IfStmt represents an if/else-if/else conditional statement.
	IfStmt struct {
		Attrs  []*Attribute // attributes applied to the if statement
		Cond   Expr         // condition that determines which branch executes
		Then   *BlockStmt   // block executed when Cond is true
		ElseIf *IfStmt      // chained else-if clause; nil if absent
		Else   *BlockStmt   // block executed when all conditions are false; nil if absent
	}

	// IfAttrStmt is a conditional statement whose presence is controlled by an
	// @if attribute evaluated at compile time.
	IfAttrStmt IfAttr[Stmt]

	// IncDecStmt represents an increment or decrement expression statement
	// (e.g. x++ or x--).
	IncDecStmt struct {
		Attrs []*Attribute // attributes applied to the statement
		LHS   Expr         // expression being incremented or decremented
		Op    string       // operator: "++" or "--"
	}

	// LoopStmt represents an unconditional loop that repeats until an explicit
	// break.
	LoopStmt struct {
		Attrs     []*Attribute // attributes applied to the loop
		BodyAttrs []*Attribute // attributes applied to the loop body block
		Body      *BlockStmt   // loop body executed on each iteration
	}

	// ReturnStmt represents a return statement, optionally carrying a value.
	ReturnStmt struct {
		Attrs []*Attribute // attributes applied to the statement
		Value Expr         // value returned to the caller; nil for void returns
	}

	// SwitchStmt represents a switch statement that dispatches on an expression
	// value across one or more case clauses.
	SwitchStmt struct {
		Attrs   []*Attribute // attributes applied to the switch statement
		Expr    Expr         // expression whose value is matched against the clauses
		Clauses []Clause     // ordered list of case and default clauses
	}

	// Clause is the interface implemented by all switch clause nodes.
	Clause interface {
		Node
		switchClauseNode()
	}

	// CaseClause represents a single case or default clause inside a switch
	// statement. A nil Selectors slice indicates the default clause.
	CaseClause struct {
		Attrs     []*Attribute // attributes applied to the clause
		Selectors []Expr       // case selector expressions; nil for the default clause
		Body      *BlockStmt   // statements executed when this clause matches
	}

	// IfAttrClause is a conditional switch clause whose presence is controlled
	// by an @if attribute evaluated at compile time.
	IfAttrClause IfAttr[Clause]

	// VarStmt declares a mutable local variable with an optional type and
	// initializer.
	VarStmt struct {
		Attrs        []*Attribute   // attributes applied to the declaration
		TemplateArgs []Expr         // optional template arguments for the var keyword (e.g. address space)
		Name         *Ident         // declared variable name
		Type         *TypeSpecifier // explicit type annotation; nil when inferred from the initializer
		Init         Expr           // optional initializer expression; nil if not provided
	}

	// ValStmt declares an immutable local binding using let or const.
	// Keyword is "let" for runtime-constant bindings and "const" for
	// compile-time constants.
	ValStmt struct {
		Attrs   []*Attribute   // attributes applied to the declaration
		Keyword string         // binding keyword: "let" or "const"
		Name    *Ident         // declared binding name
		Type    *TypeSpecifier // explicit type annotation; nil when inferred from the initializer
		Init    Expr           // initializer expression
	}

	// WhileStmt represents a while-loop that repeats as long as Cond is true.
	WhileStmt struct {
		Attrs []*Attribute // attributes applied to the loop
		Cond  Expr         // loop condition evaluated before each iteration
		Body  *BlockStmt   // loop body executed when Cond is true
	}
)

func (*AssignmentStmt) stmtNode()  {}
func (*BreakStmt) stmtNode()       {}
func (*BlockStmt) stmtNode()       {}
func (*ConstAssertStmt) stmtNode() {}
func (*ContinueStmt) stmtNode()    {}
func (*ContinuingStmt) stmtNode()  {}
func (*DiscardStmt) stmtNode()     {}
func (*EmptyStmt) stmtNode()       {}
func (*ForStmt) stmtNode()         {}
func (*FuncCallStmt) stmtNode()    {}
func (*IfStmt) stmtNode()          {}
func (*IfAttrStmt) stmtNode()      {}
func (*IncDecStmt) stmtNode()      {}
func (*LoopStmt) stmtNode()        {}
func (*ReturnStmt) stmtNode()      {}
func (*SwitchStmt) stmtNode()      {}
func (*VarStmt) stmtNode()         {}
func (*ValStmt) stmtNode()         {}
func (*WhileStmt) stmtNode()       {}

func (*AssignmentStmt) node()  {}
func (*BreakStmt) node()       {}
func (*BlockStmt) node()       {}
func (*ConstAssertStmt) node() {}
func (*ContinueStmt) node()    {}
func (*ContinuingStmt) node()  {}
func (*DiscardStmt) node()     {}
func (*EmptyStmt) node()       {}
func (*ForStmt) node()         {}
func (*FuncCallStmt) node()    {}
func (*IfStmt) node()          {}
func (*IfAttrStmt) node()      {}
func (*IncDecStmt) node()      {}
func (*LoopStmt) node()        {}
func (*ReturnStmt) node()      {}
func (*SwitchStmt) node()      {}
func (*VarStmt) node()         {}
func (*ValStmt) node()         {}
func (*WhileStmt) node()       {}

func (*IfAttrClause) switchClauseNode() {}
func (*CaseClause) switchClauseNode()   {}

func (*IfAttrClause) node() {}
func (*CaseClause) node()   {}

// ----------------------------------------------------------------------------
// Expr
type (
	// Expr is the interface implemented by every expression node.
	Expr interface {
		Node
		exprNode()
	}

	// AddrOfExpr represents the address-of unary expression (&operand), which
	// produces a pointer to the storage location of its operand.
	AddrOfExpr struct {
		Operand Expr // expression whose address is taken
	}

	// BinaryExpr represents a binary infix expression such as arithmetic,
	// comparison, or logical operations.
	BinaryExpr struct {
		Op    string // binary operator (e.g. "+", "==", "&&")
		Left  Expr   // left-hand operand
		Right Expr   // right-hand operand
	}

	// CallExpr represents a function or constructor call, optionally with
	// template arguments.
	CallExpr struct {
		Callee       *Ident // name of the function or type constructor being called
		TemplateArgs []Expr // optional template arguments enclosed in angle brackets
		Args         []Expr // positional call arguments
	}

	// DerefExpr represents the pointer dereference unary expression (*operand),
	// which yields the value at the address held by the operand.
	DerefExpr struct {
		Operand Expr // pointer expression being dereferenced
	}

	// Ident represents an identifier, either a simple name or a qualified path.
	// For simple names Path is nil and Val holds the full name. For qualified
	// references such as "package::foo::bar", Path is ["package","foo"] and Val
	// is "bar".
	Ident struct {
		Path []string // leading path segments for qualified names; nil for simple identifiers
		Val  string   // final symbol name, or the full name for simple identifiers
	}

	// IndexExpr represents a subscript expression (base[index]).
	IndexExpr struct {
		Base  Expr // expression being indexed
		Index Expr // index value
	}

	// LitExpr represents a literal value token such as a number or boolean
	// constant.
	LitExpr struct {
		Val string // source text of the literal (e.g. "42", "3.14", "true")
	}

	// MemberExpr represents a struct field access expression (base.member).
	MemberExpr struct {
		Base   Expr   // expression whose field is being accessed
		Member string // name of the field being accessed
	}

	// ParenExpr represents an explicitly parenthesized expression. It is kept
	// in the AST to allow round-trip printing without altering precedence.
	ParenExpr struct {
		Inner Expr // the expression enclosed in parentheses
	}

	// UnaryExpr represents a unary prefix expression.
	UnaryExpr struct {
		Op      string // unary operator (e.g. "-", "!", "~")
		Operand Expr   // expression the operator is applied to
	}
)

func (*AddrOfExpr) exprNode() {}
func (*BinaryExpr) exprNode() {}
func (*CallExpr) exprNode()   {}
func (*DerefExpr) exprNode()  {}
func (*Ident) exprNode()      {}
func (*IndexExpr) exprNode()  {}
func (*LitExpr) exprNode()    {}
func (*MemberExpr) exprNode() {}
func (*ParenExpr) exprNode()  {}
func (*UnaryExpr) exprNode()  {}

func (*AddrOfExpr) node() {}
func (*BinaryExpr) node() {}
func (*CallExpr) node()   {}
func (*DerefExpr) node()  {}
func (*Ident) node()      {}
func (*IndexExpr) node()  {}
func (*LitExpr) node()    {}
func (*MemberExpr) node() {}
func (*ParenExpr) node()  {}
func (*UnaryExpr) node()  {}

// ----------------------------------------------------------------------------
// Type, Attributes, Identifiers, Values, etc.

type (
	// Attribute represents a WGSL attribute (e.g. @location(0), @vertex).
	Attribute struct {
		Node
		Name string // attribute name without the leading @ (e.g. "location", "vertex")
		Args []Expr // optional attribute arguments
	}

	// TypeSpecifier represents a type reference, optionally parameterized with
	// template arguments (e.g. vec3<f32>, array<u32, 4>).
	TypeSpecifier struct {
		Node
		Name         *Ident // name of the type being referenced
		TemplateArgs []Expr // optional template arguments (e.g. element type for vec/array)
	}
)

// File is the root AST node for a single parsed source file.
type File struct {
	Node
	Decls []Decl // top-level declarations in source order
}

func (*Attribute) node()     {}
func (*TypeSpecifier) node() {}
func (*TypeSpecifier) exprNode() {}
func (*File) node()          {}
