package optimizer

import (
	"github.com/zylo-lang/zylo/internal/ast"
)

// Optimizer performs AST optimizations
type Optimizer struct{}

// NewOptimizer creates a new optimizer instance
func NewOptimizer() *Optimizer {
	return &Optimizer{}
}

// Optimize applies all optimizations to the AST
func (o *Optimizer) Optimize(program *ast.Program) {
	o.constantFolding(program)
	o.deadCodeElimination(program)
	o.constantPropagation(program)
}

// constantFolding performs constant folding on expressions
func (o *Optimizer) constantFolding(node ast.Node) ast.Node {
	switch n := node.(type) {
	case *ast.Program:
		for i, stmt := range n.Statements {
			n.Statements[i] = o.constantFolding(stmt).(ast.Statement)
		}
		return n
	case *ast.ExpressionStatement:
		n.Expression = o.constantFolding(n.Expression).(ast.Expression)
		return n
	case *ast.VarStatement:
		if n.Value != nil {
			n.Value = o.constantFolding(n.Value).(ast.Expression)
		}
		return n
	case *ast.ReturnStatement:
		if n.ReturnValue != nil {
			n.ReturnValue = o.constantFolding(n.ReturnValue).(ast.Expression)
		}
		return n
	case *ast.IfStatement:
		n.Condition = o.constantFolding(n.Condition).(ast.Expression)
		n.Consequence = o.constantFolding(n.Consequence).(*ast.BlockStatement)
		if n.Alternative != nil {
			n.Alternative = o.constantFolding(n.Alternative).(*ast.BlockStatement)
		}
		return n
	case *ast.BlockStatement:
		for i, stmt := range n.Statements {
			n.Statements[i] = o.constantFolding(stmt).(ast.Statement)
		}
		return n
	case *ast.InfixExpression:
		n.Left = o.constantFolding(n.Left).(ast.Expression)
		n.Right = o.constantFolding(n.Right).(ast.Expression)
		return o.foldInfixExpression(n)
	case *ast.PrefixExpression:
		n.Right = o.constantFolding(n.Right).(ast.Expression)
		return o.foldPrefixExpression(n)
	case *ast.CallExpression:
		for i, arg := range n.Arguments {
			n.Arguments[i] = o.constantFolding(arg).(ast.Expression)
		}
		return n
	case *ast.ListLiteral:
		for i, elem := range n.Elements {
			n.Elements[i] = o.constantFolding(elem).(ast.Expression)
		}
		return n
	case *ast.MapLiteral:
		for k, v := range n.Pairs {
			n.Pairs[k] = o.constantFolding(v).(ast.Expression)
		}
		return n
	default:
		return n
	}
}

// foldInfixExpression attempts to fold constant infix expressions
func (o *Optimizer) foldInfixExpression(expr *ast.InfixExpression) ast.Expression {
	// Try arithmetic folding first
	if folded := o.foldArithmeticExpression(expr); folded != nil {
		return folded
	}

	// Try comparison folding
	if folded := o.foldComparisonExpression(expr); folded != nil {
		return folded
	}

	return expr
}

// foldArithmeticExpression folds arithmetic operations on constants
func (o *Optimizer) foldArithmeticExpression(expr *ast.InfixExpression) ast.Expression {
	leftLit, leftOk := expr.Left.(*ast.NumberLiteral)
	rightLit, rightOk := expr.Right.(*ast.NumberLiteral)

	if !leftOk || !rightOk {
		return nil
	}

	leftVal, leftIsInt := leftLit.Value.(int64)
	rightVal, rightIsInt := rightLit.Value.(int64)

	if !leftIsInt || !rightIsInt {
		// For now, only handle integers
		return nil
	}

	var result int64
	switch expr.Operator {
	case "+":
		result = leftVal + rightVal
	case "-":
		result = leftVal - rightVal
	case "*":
		result = leftVal * rightVal
	case "/":
		if rightVal != 0 {
			result = leftVal / rightVal
		} else {
			return nil // Avoid division by zero
		}
	case "%":
		if rightVal != 0 {
			result = leftVal % rightVal
		} else {
			return nil
		}
	default:
		return nil
	}

	return &ast.NumberLiteral{
		Token: expr.Token,
		Value: result,
	}
}

// foldComparisonExpression folds comparison operations on constants
func (o *Optimizer) foldComparisonExpression(expr *ast.InfixExpression) ast.Expression {
	leftLit, leftOk := expr.Left.(*ast.NumberLiteral)
	rightLit, rightOk := expr.Right.(*ast.NumberLiteral)

	if !leftOk || !rightOk {
		return nil
	}

	leftVal, leftIsInt := leftLit.Value.(int64)
	rightVal, rightIsInt := rightLit.Value.(int64)

	if !leftIsInt || !rightIsInt {
		return nil
	}

	var result bool
	switch expr.Operator {
	case "==":
		result = leftVal == rightVal
	case "!=":
		result = leftVal != rightVal
	case "<":
		result = leftVal < rightVal
	case "<=":
		result = leftVal <= rightVal
	case ">":
		result = leftVal > rightVal
	case ">=":
		result = leftVal >= rightVal
	default:
		return nil
	}

	return &ast.BooleanLiteral{
		Token: expr.Token,
		Value: result,
	}
}

// foldPrefixExpression attempts to fold constant prefix expressions
func (o *Optimizer) foldPrefixExpression(expr *ast.PrefixExpression) ast.Expression {
	if expr.Operator != "-" {
		return expr
	}

	numLit, ok := expr.Right.(*ast.NumberLiteral)
	if !ok {
		return expr
	}

	val, isInt := numLit.Value.(int64)
	if !isInt {
		return expr
	}

	return &ast.NumberLiteral{
		Token: expr.Token,
		Value: -val,
	}
}

// deadCodeElimination removes unreachable code
func (o *Optimizer) deadCodeElimination(node ast.Node) ast.Node {
	switch n := node.(type) {
	case *ast.Program:
		var optimizedStatements []ast.Statement
		for _, stmt := range n.Statements {
			optimizedStmt := o.deadCodeElimination(stmt)
			if optimizedStmt != nil {
				optimizedStatements = append(optimizedStatements, optimizedStmt.(ast.Statement))
			}
		}
		n.Statements = optimizedStatements
		return n
	case *ast.IfStatement:
		n.Condition = o.deadCodeElimination(n.Condition).(ast.Expression)
		n.Consequence = o.deadCodeElimination(n.Consequence).(*ast.BlockStatement)

		// Check if condition is a constant boolean
		if boolLit, ok := n.Condition.(*ast.BooleanLiteral); ok {
			if boolLit.Value {
				// If condition is always true, replace with consequence block
				return n.Consequence
			} else {
				// If condition is always false, eliminate the if statement
				return nil
			}
		}

		if n.Alternative != nil {
			n.Alternative = o.deadCodeElimination(n.Alternative).(*ast.BlockStatement)
		}
		return n
	case *ast.BlockStatement:
		var optimizedStatements []ast.Statement
		for _, stmt := range n.Statements {
			optimizedStmt := o.deadCodeElimination(stmt)
			if optimizedStmt != nil {
				optimizedStatements = append(optimizedStatements, optimizedStmt.(ast.Statement))
			}
		}
		n.Statements = optimizedStatements
		return n
	default:
		return n
	}
}

// constantPropagation performs constant propagation
func (o *Optimizer) constantPropagation(node ast.Node) ast.Node {
	// For now, implement a simple version
	// In a full implementation, this would track variable assignments and substitute uses
	switch n := node.(type) {
	case *ast.Program:
		for i, stmt := range n.Statements {
			n.Statements[i] = o.constantPropagation(stmt).(ast.Statement)
		}
		return n
	case *ast.BlockStatement:
		for i, stmt := range n.Statements {
			n.Statements[i] = o.constantPropagation(stmt).(ast.Statement)
		}
		return n
	default:
		return n
	}
}