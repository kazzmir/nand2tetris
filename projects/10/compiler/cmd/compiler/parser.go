package main

import (
    "fmt"
    "io"
    "strings"
)

type Kind uint32

const (
    ASTKindClass Kind = iota
    ASTKindIdentifier
    ASTKindStaticDeclaration
    ASTKindType
    ASTKindField
    ASTKindFunction
    ASTKindParameter
    ASTKindBlock
    ASTKindLet
    ASTKindVar
    ASTKindDo
    ASTKindReturn
    ASTKindExpression
    ASTKindConstant
    ASTKindCall
    ASTKindMethodCall
    ASTKindThis
    ASTKindReference
    ASTKindOperator
    ASTKindIf
    ASTKindBoolean
    ASTKindString
    ASTKindNull
    ASTKindIndexExpression
    ASTKindNot
    ASTKindNegation
    ASTKindConstructor
    ASTKindMethod
    ASTKindWhile
)

func (kind Kind) Name() string {
    switch kind {
    case ASTKindClass: return "class"
    case ASTKindIdentifier: return "identifier"
    case ASTKindStaticDeclaration: return "static field"
    case ASTKindType: return "type"
    case ASTKindField: return "field"
    case ASTKindFunction: return "function"
    case ASTKindParameter: return "parameter"
    case ASTKindBlock: return "block"
    case ASTKindLet: return "let"
    case ASTKindVar: return "var"
    case ASTKindDo: return "do"
    case ASTKindReturn: return "return"
    case ASTKindExpression: return "expression"
    case ASTKindConstant: return "constant"
    case ASTKindCall: return "call"
    case ASTKindMethodCall: return "method call"
    case ASTKindThis: return "this"
    case ASTKindReference: return "reference"
    case ASTKindOperator: return "operator"
    case ASTKindIf: return "if"
    case ASTKindBoolean: return "boolean"
    case ASTKindString: return "string"
    case ASTKindNull: return "null"
    case ASTKindIndexExpression: return "array index expression"
    case ASTKindNot: return "not"
    case ASTKindNegation: return "negation"
    case ASTKindConstructor: return "constructor"
    case ASTKindMethod: return "method"
    case ASTKindWhile: return "while"
    }

    return "??"
}

type ASTNode interface {
    Kind() Kind
    ToSExpression() string
}

type ASTClass struct {
    ASTNode

    Name string
    Body []ASTNode
}

func (ast *ASTClass) ToSExpression() string {
    var body strings.Builder
    for _, line := range ast.Body {
        body.WriteString(line.ToSExpression())
        body.WriteByte('\n')
    }
    return fmt.Sprintf("(class %v\n%v)", ast.Name, body.String())
}

func (ast *ASTClass) Kind() Kind {
    return ASTKindClass
}

type ASTWhile struct {
    Condition ASTExpression
    Body *ASTBlock
}

func (ast *ASTWhile) ToSExpression() string {
    return fmt.Sprintf("(while %v %v)", ast.Condition.ToSExpression(), ast.Body.ToSExpression())
}

func (ast *ASTWhile) Kind() Kind {
    return ASTKindWhile
}

type ASTConstructor struct {
    Class string
    Name string
    Parameters []*ASTParameter
    Body *ASTBlock
}

func (ast *ASTConstructor) Kind() Kind {
    return ASTKindConstructor
}

func (ast *ASTConstructor) ToSExpression() string {
    var builder strings.Builder

    builder.WriteString("(constructor ")
    builder.WriteString(ast.Class)
    builder.WriteString(" ")
    builder.WriteString(ast.Name)
    builder.WriteString("(")
    for i, parameter := range ast.Parameters {
        if i > 0 {
            builder.WriteByte(' ')
        }
        builder.WriteString(parameter.ToSExpression())
    }
    builder.WriteString(")")
    builder.WriteString(ast.Body.ToSExpression())

    return builder.String()
}

type ASTCall struct {
    Name string
    Arguments []ASTExpression
}

func (ast *ASTCall) Kind() Kind {
    return ASTKindCall
}

func (ast *ASTCall) ToSExpression() string {
    var builder strings.Builder

    builder.WriteString("(")
    builder.WriteString(ast.Name)
    builder.WriteByte(' ')
    for _, argument := range ast.Arguments {
        builder.WriteString(argument.ToSExpression())
        builder.WriteByte(' ')
    }
    builder.WriteString(")")

    return builder.String()
}

type ASTString struct {
    Value string
}

func (ast *ASTString) Kind() Kind {
    return ASTKindString
}

func (ast *ASTString) ToSExpression() string {
    return fmt.Sprintf("\"%v\"", ast.Value)
}

type ASTNull struct {
}

func (ast *ASTNull) Kind() Kind {
    return ASTKindNull
}

func (ast *ASTNull) ToSExpression() string {
    return "null"
}

type ASTIndexExpression struct {
    Left ASTExpression
    Index ASTExpression
}

func (ast *ASTIndexExpression) ToSExpression() string {
    return fmt.Sprintf("%v[%v]", ast.Left.ToSExpression(), ast.Index.ToSExpression())
}

func (ast *ASTIndexExpression) Kind() Kind {
    return ASTKindIndexExpression
}

type ASTMethodCall struct {
    Left ASTExpression
    Call *ASTCall
}

func (ast *ASTMethodCall) ToSExpression() string {
    return fmt.Sprintf("(call %v %v)", ast.Left.ToSExpression(), ast.Call.ToSExpression())
}

func (ast *ASTMethodCall) Kind() Kind {
    return ASTKindMethodCall
}

type ASTNot struct {
    Expression ASTExpression
}

func (ast *ASTNot) ToSExpression() string {
    return fmt.Sprintf("(not %v)", ast.Expression.ToSExpression())
}

func (ast *ASTNot) Kind() Kind {
    return ASTKindNot
}

type ASTNegation struct {
    Expression ASTExpression
}

func (ast *ASTNegation) ToSExpression() string {
    return fmt.Sprintf("(- %v)", ast.Expression.ToSExpression())
}

func (ast *ASTNegation) Kind() Kind {
    return ASTKindNegation
}

type ASTBoolean struct {
    Value bool
}

func (ast *ASTBoolean) ToSExpression() string {
    return fmt.Sprintf("%v", ast.Value)
}

func (ast *ASTBoolean) Kind() Kind {
    return ASTKindBoolean
}

type ASTType struct {
    /* will either be an int, char, boolean, or identifier */
    Type ASTNode
}

func (ast *ASTType) ToSExpression() string {
    return ast.Type.ToSExpression()
}

func (ast *ASTType) Kind() Kind {
    return ASTKindType
}

type ASTIf struct {
    Condition ASTExpression
    Then *ASTBlock
    Else *ASTBlock
}

func (ast *ASTIf) ToSExpression() string {
    var builder strings.Builder

    builder.WriteString("(if ")
    builder.WriteString(ast.Condition.ToSExpression())
    builder.WriteString(" ")
    builder.WriteString(ast.Then.ToSExpression())
    if ast.Else != nil {
        builder.WriteString(ast.Else.ToSExpression())
    }

    return builder.String()
}

func (ast *ASTIf) Kind() Kind {
    return ASTKindIf
}

type ASTOperator struct {
    Operator TokenKind // lame to use TokenKind here
    Left ASTExpression
    Right ASTExpression
}

func (ast *ASTOperator) ToSExpression() string {
    return fmt.Sprintf("(%v %v %v)", ast.Operator.Name(), ast.Left.ToSExpression(), ast.Right.ToSExpression())
}

func (ast *ASTOperator) Kind() Kind {
    return ASTKindOperator
}

type ASTThis struct {
}

func (ast *ASTThis) ToSExpression() string {
    return "this"
}

func (ast *ASTThis) Kind() Kind {
    return ASTKindThis
}

type ASTConstant struct {
    Number string
}

func (ast *ASTConstant) ToSExpression() string {
    return ast.Number
}

func (ast *ASTConstant) Kind() Kind {
    return ASTKindConstant
}

type ASTReference struct {
    Name string
}

func (ast *ASTReference) ToSExpression() string {
    return ast.Name
}

func (ast *ASTReference) Kind() Kind {
    return ASTKindReference
}

type ASTMethod struct {
    ReturnType *ASTType
    Name string
    Parameters []*ASTParameter
    Body *ASTBlock
}

func (ast *ASTMethod) Kind() Kind {
    return ASTKindMethod
}

func (ast *ASTMethod) ToSExpression() string {
    var builder strings.Builder

    builder.WriteString("(method ")
    builder.WriteString(ast.Name)
    builder.WriteString("(")
    for _, parameter := range ast.Parameters {
        builder.WriteByte(' ')
        builder.WriteString(parameter.ToSExpression())
    }
    builder.WriteString(") -> ")
    builder.WriteString(ast.ReturnType.ToSExpression())
    builder.WriteByte(' ')
    builder.WriteString(ast.Body.ToSExpression())

    return builder.String()
}

type ASTFunction struct {
    ReturnType *ASTType
    Name string
    Parameters []*ASTParameter
    Body *ASTBlock
}

func (ast *ASTFunction) ToSExpression() string {
    var builder strings.Builder

    builder.WriteString("(function ")
    builder.WriteString(ast.Name)

    builder.WriteString(" (")
    for _, parameter := range ast.Parameters {
        builder.WriteString(parameter.ToSExpression())
        builder.WriteString(" ")
    }
    builder.WriteString(") -> ")
    builder.WriteString(ast.ReturnType.ToSExpression())
    builder.WriteString(ast.Body.ToSExpression())
    builder.WriteString(")")

    return builder.String()
}

func (ast *ASTFunction) Kind() Kind {
    return ASTKindFunction
}

type ASTParameter struct {
    Type *ASTType
    Name string
}

func (ast *ASTParameter) Kind() Kind {
    return ASTKindParameter
}

func (ast *ASTParameter) ToSExpression() string {
    return fmt.Sprintf("(%v %v)", ast.Name, ast.Type.ToSExpression())
}

type ASTBlock struct {
    Statements []ASTNode
}

func (ast *ASTBlock) ToSExpression() string {
    var builder strings.Builder
    builder.WriteString("(block\n")
    for _, statement := range ast.Statements {
        builder.WriteString(statement.ToSExpression())
        builder.WriteByte('\n')
    }
    builder.WriteString(")")
    return builder.String()
}

func (ast *ASTBlock) Kind() Kind {
    return ASTKindBlock
}

type ASTVar struct {
    Type *ASTType
    Names []string
}

func (ast *ASTVar) ToSExpression() string {
    var builder strings.Builder
    builder.WriteString("(var ")
    builder.WriteString(ast.Type.ToSExpression())
    for _, name := range ast.Names {
        builder.WriteByte(' ')
        builder.WriteString(name)
    }
    builder.WriteString(")")
    return builder.String()
}

func (ast *ASTVar) Kind() Kind {
    return ASTKindVar
}

type ASTExpression interface {
    ASTNode
}

type ASTLet struct {
    Name string
    ArrayIndex ASTExpression
    Expression ASTExpression
}

func (ast *ASTLet) ToSExpression() string {
    var part strings.Builder
    part.WriteString(ast.Name)
    if ast.ArrayIndex != nil {
        part.WriteString("[")
        part.WriteString(ast.ArrayIndex.ToSExpression())
        part.WriteString("]")
    }
    return fmt.Sprintf("(let %v = %v)", part.String(), ast.Expression.ToSExpression())
}

func (ast *ASTLet) Kind() Kind {
    return ASTKindLet
}

type ASTDo struct {
    Expression ASTExpression
}

func (ast *ASTDo) ToSExpression() string {
    return fmt.Sprintf("(do %v)", fmt.Sprintf(ast.Expression.ToSExpression()))
}

func (ast *ASTDo) Kind() Kind {
    return ASTKindDo
}

type ASTReturn struct {
    Expression ASTExpression
}

func (ast *ASTReturn) ToSExpression() string {
    var builder strings.Builder

    builder.WriteString("(return")
    if ast.Expression != nil {
        builder.WriteByte(' ')
        builder.WriteString(ast.Expression.ToSExpression())
    }
    builder.WriteString(")")

    return builder.String()
}

func (ast *ASTReturn) Kind() Kind {
    return ASTKindReturn
}

type ASTStatic struct {
    Type *ASTType
    Names []*ASTIdentifier
}

func (ast *ASTStatic) ToSExpression() string {
    var builder strings.Builder

    builder.WriteString("(static ")
    builder.WriteString(ast.Type.ToSExpression())
    for _, name := range ast.Names {
        builder.WriteByte(' ')
        builder.WriteString(name.ToSExpression())
    }
    builder.WriteString(")")

    return builder.String()
}

func (ast *ASTStatic) Kind() Kind {
    return ASTKindStaticDeclaration
}

type ASTIdentifier struct {
    Name string
}

func (ast *ASTIdentifier) Kind() Kind {
    return ASTKindIdentifier
}

func (ast *ASTIdentifier) ToSExpression() string {
    return ast.Name
}

type ASTField struct {
    Type *ASTType
    Names []string
}

func (ast *ASTField) ToSExpression() string {
    var builder strings.Builder

    builder.WriteString("(field ")
    builder.WriteString(ast.Type.ToSExpression())
    for _, name := range ast.Names {
        builder.WriteByte(' ')
        builder.WriteString(name)
    }
    builder.WriteString(")")

    return builder.String()
}

func (ast *ASTField) Kind() Kind {
    return ASTKindField
}

func isExpression(ast ASTNode) bool {
    switch ast.Kind() {
        case ASTKindCall, ASTKindReference, ASTKindThis, ASTKindOperator,
             ASTKindMethodCall, ASTKindNegation, ASTKindNot: return true
        default: return false
    }
}

/* provides single token lookahead and filters whitespace */
type TokenStream struct {
    tokens chan Token
    next Token
    hasNext bool
}

func (stream *TokenStream) Next() (Token, error) {
    if stream.hasNext {
        return stream.next, nil
    }

    for {
        token, ok := <-stream.tokens
        if !ok {
            return Token{}, fmt.Errorf("out of tokens")
        }

        if token.Kind == TokenWhitespace {
            continue
        }

        stream.next = token
        stream.hasNext = true
        return stream.next, nil
    }
}

func (stream *TokenStream) Consume() (Token, error) {
    if stream.hasNext {
        stream.hasNext = false
        return stream.next, nil
    }

    for {
        token, ok := <-stream.tokens
        if !ok {
            return Token{}, fmt.Errorf("out of tokens")
        }

        if token.Kind == TokenWhitespace {
            continue
        }

        return token, nil
    }
}

func parse(reader io.Reader) (ASTNode, error) {
    tokens := make(chan Token, 1000)

    var lexerError error

    go func(){
        lexerError = standardLexer(reader, tokens)
    }()

    stream := &TokenStream{
        tokens: tokens,
        hasNext: false,
    }

    class, err := parseClass(stream)

    if err != nil {
        return nil, err
    }

    var unparsedToken Token
    unparsed := false
    for {
        token, empty := stream.Consume()
        if empty != nil {
            break
        }
        unparsedToken = token
        unparsed = true
    }

    if lexerError != nil {
        return nil, lexerError
    }

    if unparsed {
        return nil, fmt.Errorf("unparsed token %v", unparsedToken)
    }

    return class, err
}

func consumeToken(tokens *TokenStream, kind TokenKind) error {
    token, err := tokens.Consume()
    if err != nil {
        return fmt.Errorf("out of tokens")
    }

    if token.Kind != kind {
        return fmt.Errorf("expected token %v but found %v", kind.Name(), token.String())
    }

    return nil
}

func parseTypeNode(tokens *TokenStream) (*ASTType, error) {
    next, err := tokens.Consume()
    if err != nil {
        return nil, err
    }

    switch next.Kind {
        case TokenInt:
            return &ASTType{
                Type: &ASTIdentifier{Name: "int"},
            }, nil
        case TokenChar:
            return &ASTType{
                Type: &ASTIdentifier{Name: "char"},
            }, nil
        case TokenBoolean:
            return &ASTType{
                Type: &ASTIdentifier{Name: "boolean"},
            }, nil
        case TokenIdentifier:
            return &ASTType{
                Type: &ASTIdentifier{Name: next.Value},
            }, nil
        case TokenVoid:
            return &ASTType{
                Type: &ASTIdentifier{Name: "void"},
            }, nil
    }

    return nil, fmt.Errorf("expected a type to be one of int, char, boolean, or identifier but was %v", next.String())
}

/* static <type> <identifier> ...; */
func parseStaticDeclaration(tokens *TokenStream) (*ASTStatic, error) {
    static, err := tokens.Consume()
    if err != nil {
        return nil, fmt.Errorf("could not parse static field: %v", err)
    }

    if static.Kind != TokenStatic {
        return nil, fmt.Errorf("expected a 'static' keyword but found %v", static.String())
    }

    typeNode, err := parseTypeNode(tokens)
    if err != nil {
        return nil, err
    }

    var names []*ASTIdentifier

    for {
        next, err := tokens.Consume()
        if err != nil {
            return nil, err
        }

        if next.Kind == TokenSemicolon {
            break
        }

        if next.Kind == TokenIdentifier {
            names = append(names, &ASTIdentifier{Name: next.Value})
        }
    }

    return &ASTStatic{
        Type: typeNode,
        Names: names,
    }, nil
}

func parseIdentifierList(tokens *TokenStream) ([]string, error) {
    var names []string

    name, err := tokens.Consume()
    if err != nil {
        return nil, err
    }

    if name.Kind != TokenIdentifier {
        return nil, fmt.Errorf("expected an identifier but got %v", name.String())
    }

    names = append(names, name.Value)

    for {
        next, err := tokens.Next()
        if err != nil {
            return nil, err
        }

        if next.Kind != TokenComma {
            return names, nil
        }

        /* consume the comma */
        tokens.Consume()

        name, err = tokens.Consume()
        if err != nil {
            return nil, err
        }

        if name.Kind != TokenIdentifier {
            return nil, fmt.Errorf("expected an identifier but got %v", name.String())
        }

        names = append(names, name.Value)
    }

    return names, nil
}

func parseFieldDeclaration(tokens *TokenStream) (*ASTField, error) {
    err := consumeToken(tokens, TokenField)
    if err != nil {
        return nil, err
    }

    typeNode, err := parseTypeNode(tokens)
    if err != nil {
        return nil, err
    }

    names, err := parseIdentifierList(tokens)
    if err != nil {
        return nil, err
    }

    err = consumeToken(tokens, TokenSemicolon)
    if err != nil {
        return nil, err
    }

    return &ASTField{
        Type: typeNode,
        Names: names,
    }, nil
}

func parseVarDeclaration(tokens *TokenStream) (*ASTVar, error) {
    err := consumeToken(tokens, TokenVar)
    if err != nil {
        return nil, err
    }

    typeNode, err := parseTypeNode(tokens)
    if err != nil {
        return nil, err
    }

    var names []string

    for {
        name, err := tokens.Next()
        if err != nil {
            return nil, err
        }

        if name.Kind == TokenSemicolon {
            break
        }

        tokens.Consume()

        if name.Kind != TokenIdentifier {
            return nil, fmt.Errorf("expected an identifier in a var declaration: %v", name.String())
        }

        names = append(names, name.Value)

        /* check if the next token is a comma, in which case just consume it */
        next, err := tokens.Next()
        if err != nil {
            return nil, err
        }

        if next.Kind == TokenComma {
            tokens.Consume()
        }
    }

    err = consumeToken(tokens, TokenSemicolon)
    if err != nil {
        return nil, err
    }

    if len(names) == 0 {
        return nil, fmt.Errorf("no identifiers given in a var declaration")
    }

    return &ASTVar{
        Type: typeNode,
        Names: names,
    }, nil
}

/* name(<expression>, ...) */
func parseCall(tokens *TokenStream) (*ASTCall, error) {
    name, err := tokens.Consume()
    if err != nil {
        return nil, err
    }

    if name.Kind != TokenIdentifier {
        return nil, fmt.Errorf("call must start with identifier but got %v", name.String())
    }

    err = consumeToken(tokens, TokenLeftParens)
    if err != nil {
        return nil, err
    }

    var arguments []ASTExpression

    for {
        next, err := tokens.Next()
        if err != nil {
            return nil, err
        }

        if next.Kind == TokenRightParens {
            break
        }

        expression, err := parseExpression(tokens)
        if err != nil {
            return nil, err
        }

        arguments = append(arguments, expression)

        next, err = tokens.Next()
        if err != nil {
            return nil, err
        }

        if next.Kind == TokenComma {
            tokens.Consume()
            continue
        }
    }

    err = consumeToken(tokens, TokenRightParens)
    if err != nil {
        return nil, err
    }

    return &ASTCall{
        Name: name.Value,
        Arguments: arguments,
    }, nil
}

func parseExpressionNoOp(tokens *TokenStream) (ASTExpression, error) {
    next, err := tokens.Next()
    if err != nil {
        return nil, err
    }

    switch next.Kind {
        case TokenLeftParens:
            tokens.Consume()
            expression, err := parseExpression(tokens)
            if err != nil {
                return nil, err
            }
            err = consumeToken(tokens, TokenRightParens)
            if err != nil {
                return nil, err
            }
            return expression, nil
        case TokenTrue, TokenFalse:
            tokens.Consume()
            isTrue := next.Kind == TokenTrue
            return &ASTBoolean{Value: isTrue}, nil
        case TokenString:
            tokens.Consume()
            return &ASTString{Value: next.Value}, nil
        case TokenNull:
            tokens.Consume()
            return &ASTNull{}, nil
        case TokenNumber:
            number := next
            tokens.Consume()
            return &ASTConstant{Number: number.Value}, nil
        case TokenThis, TokenIdentifier:
            /* either a variable reference or a x.y() call,
             * or a method call f()
             */

            id, _ := tokens.Consume()

            var left ASTExpression

            switch id.Kind {
                case TokenThis:
                    left = &ASTThis{}
                case TokenIdentifier:
                    left = &ASTReference{Name: id.Value}
                default:
                    return nil, fmt.Errorf("unknown token on the left side of a dot expression: %v", id.String())
            }

            /* Use an ASTThis node? */

            next, err = tokens.Next()
            switch next.Kind {
                case TokenDot:
                    tokens.Consume()

                    call, err := parseCall(tokens)
                    if err != nil {
                        return nil, err
                    }

                    return &ASTMethodCall{
                        Left: left,
                        Call: call,
                    }, nil
                case TokenLeftParens:
                    var arguments []ASTExpression

                    tokens.Consume()

                    for {
                        next, err := tokens.Next()
                        if err != nil {
                            return nil, err
                        }

                        if next.Kind == TokenRightParens {
                            break
                        }

                        expression, err := parseExpression(tokens)
                        if err != nil {
                            return nil, err
                        }

                        arguments = append(arguments, expression)

                        next, err = tokens.Next()
                        if err != nil {
                            return nil, err
                        }

                        if next.Kind == TokenComma {
                            tokens.Consume()
                            continue
                        }
                    }

                    err = consumeToken(tokens, TokenRightParens)
                    if err != nil {
                        return nil, err
                    }

                    return &ASTCall{
                        Name: id.Value,
                        Arguments: arguments,
                    }, nil
            }

            return left, nil
        default:
            return nil, fmt.Errorf("unknown token in expression: %v", next.String())
    }
}

func parseExpressionArrayIndex(tokens *TokenStream) (ASTExpression, error) {
    left, err := parseExpressionNoOp(tokens)
    if err != nil {
        return nil, err
    }

    next, err := tokens.Next()
    if err != nil {
        return nil, err
    }

    /* array index */
    if next.Kind == TokenLeftBracket {
        tokens.Consume()

        index, err := parseExpression(tokens)
        if err != nil {
            return nil, err
        }

        err = consumeToken(tokens, TokenRightBracket)
        if err != nil {
            return nil, err
        }

        return &ASTIndexExpression{
            Left: left,
            Index: index,
        }, nil
    } else {
        return left, nil
    }
}

func parseExpressionUnary(tokens *TokenStream) (ASTExpression, error) {
    next, err := tokens.Next()
    if err != nil {
        return nil, err
    }

    switch next.Kind {
        case TokenNot, TokenNegation:
            tokens.Consume()
            expression, err := parseExpressionArrayIndex(tokens)
            if err != nil {
                return nil, err
            }

            if next.Kind == TokenNot {
                return &ASTNot{
                    Expression: expression,
                }, nil
            } else if next.Kind == TokenNegation {
                return &ASTNegation{
                    Expression: expression,
                }, nil
            } else {
                return nil, fmt.Errorf("internal error")
            }
        default:
            return parseExpressionArrayIndex(tokens)
    }
}

func parseExpression(tokens *TokenStream) (ASTExpression, error) {
    left, err := parseExpressionUnary(tokens)
    if err != nil {
        return nil, err
    }

    for {
        next, err := tokens.Next()
        if err != nil {
            return nil, err
        }

        /* check for an operator */
        switch next.Kind {
            case TokenPlus, TokenDivision, TokenMultiply,
                 TokenLessThan, TokenGreaterThan,
                 TokenOr, TokenAnd, TokenNegation, TokenEquals:
                operator, err := tokens.Consume()

                right, err := parseExpressionUnary(tokens)
                if err != nil {
                    return nil, fmt.Errorf("could not parse operator expression: %v", err)
                }

                left = &ASTOperator{
                    Operator: operator.Kind,
                    Left: left,
                    Right: right,
                }
            default:
                return left, nil
        }
    }
}

/* let <name> [<array-expression>] = <expression> ;
 */
func parseLet(tokens *TokenStream) (*ASTLet, error) {
    err := consumeToken(tokens, TokenLet)
    if err != nil {
        return nil, err
    }

    name, err := tokens.Consume()
    if err != nil {
        return nil, err
    }

    if name.Kind != TokenIdentifier {
        return nil, fmt.Errorf("expected a name to follow 'let': %v", name)
    }

    next, err := tokens.Next()
    if err != nil {
        return nil, err
    }

    var arrayIndex ASTExpression

    if next.Kind == TokenLeftBracket {
        err = consumeToken(tokens, TokenLeftBracket)
        if err != nil {
            return nil, err
        }

        arrayIndex, err = parseExpression(tokens)
        if err != nil {
            return nil, err
        }

        err = consumeToken(tokens, TokenRightBracket)
        if err != nil {
            return nil, err
        }
    }

    err = consumeToken(tokens, TokenEquals)
    if err != nil {
        return nil, err
    }

    expression, err := parseExpression(tokens)
    if err != nil {
        return nil, fmt.Errorf("could not parse expression on the right hand side of a let: %v", err)
    }

    err = consumeToken(tokens, TokenSemicolon)
    if err != nil {
        return nil, fmt.Errorf("missing a semicolon after a let: %v", err)
    }

    return &ASTLet{
        Name: name.Value,
        ArrayIndex: arrayIndex,
        Expression: expression,
    }, nil
}

/* do <expression>; */
func parseDo(tokens *TokenStream) (*ASTDo, error) {
    err := consumeToken(tokens, TokenDo)
    if err != nil {
        return nil, err
    }

    expression, err := parseExpression(tokens)
    if err != nil {
        return nil, err
    }

    err = consumeToken(tokens, TokenSemicolon)
    if err != nil {
        return nil, err
    }

    return &ASTDo{
        Expression: expression,
    }, nil
}

/* return [<expression>] */
func parseReturn(tokens *TokenStream) (*ASTReturn, error) {
    ret, err := tokens.Consume()
    if err != nil {
        return nil, err
    }

    if ret.Kind != TokenReturn {
        return nil, fmt.Errorf("expected 'return' but found %v", ret.String())
    }

    next, err := tokens.Next()
    if err != nil {
        return nil, err
    }

    var expression ASTExpression
    if next.Kind != TokenSemicolon {
        expression, err = parseExpression(tokens)
        if err != nil {
            return nil, err
        }
    }

    err = consumeToken(tokens, TokenSemicolon)
    if err != nil {
        return nil, err
    }

    return &ASTReturn{
        Expression: expression,
    }, nil
}

func parseIf(tokens *TokenStream) (*ASTIf, error) {
    err := consumeToken(tokens, TokenIf)
    if err != nil {
        return nil, err
    }

    err = consumeToken(tokens, TokenLeftParens)
    if err != nil {
        return nil, err
    }

    condition, err := parseExpression(tokens)
    if err != nil {
        return nil, err
    }

    err = consumeToken(tokens, TokenRightParens)
    if err != nil {
        return nil, err
    }

    thenBlock, err := parseBlock(tokens)
    if err != nil {
        return nil, err
    }

    else_, err := tokens.Next()
    if err != nil {
        return nil, err
    }

    var elseBlock *ASTBlock

    if else_.Kind == TokenElse {
        tokens.Consume()
        elseBlock, err = parseBlock(tokens)
        if err != nil {
            return nil, err
        }
    }

    return &ASTIf{
        Condition: condition,
        Then: thenBlock,
        Else: elseBlock,
    }, nil
}

func parseWhile(tokens *TokenStream) (*ASTWhile, error) {
    err := consumeToken(tokens, TokenWhile)
    if err != nil {
        return nil, err
    }

    err = consumeToken(tokens, TokenLeftParens)
    if err != nil {
        return nil, err
    }

    condition, err := parseExpression(tokens)
    if err != nil {
        return nil, err
    }

    err = consumeToken(tokens, TokenRightParens)
    if err != nil {
        return nil, err
    }

    body, err := parseBlock(tokens)
    if err != nil {
        return nil, err
    }

    return &ASTWhile{
        Condition: condition,
        Body: body,
    }, nil
}

func parseBlock(tokens *TokenStream) (*ASTBlock, error) {
    var statements []ASTNode

    err := consumeToken(tokens, TokenLeftCurly)
    if err != nil {
        return nil, err
    }

    for {
        next, err := tokens.Next()
        if err != nil {
            return nil, err
        }

        switch next.Kind {
            case TokenVar:
                varDeclaration, err := parseVarDeclaration(tokens)
                if err != nil {
                    return nil, err
                }
                statements = append(statements, varDeclaration)
            case TokenLet:
                letDeclaration, err := parseLet(tokens)
                if err != nil {
                    return nil, err
                }
                statements = append(statements, letDeclaration)
            case TokenDo:
                do, err := parseDo(tokens)
                if err != nil {
                    return nil, err
                }
                statements = append(statements, do)
            case TokenWhile:
                while, err := parseWhile(tokens)
                if err != nil {
                    return nil, err
                }
                statements = append(statements, while)
            case TokenIf:
                if_, err := parseIf(tokens)
                if err != nil {
                    return nil, err
                }
                statements = append(statements, if_)
            case TokenRightCurly:
                err := consumeToken(tokens, TokenRightCurly)
                if err != nil {
                    return nil, err
                }
                return &ASTBlock{
                    Statements: statements,
                }, nil
            case TokenReturn:
                ret, err := parseReturn(tokens)
                if err != nil {
                    return nil, err
                }
                statements = append(statements, ret)
            default:
                return nil, fmt.Errorf("unexpected token %v in a block", next.String())
        }
    }
}

func parseParameterList(tokens *TokenStream) ([]*ASTParameter, error) {
    err := consumeToken(tokens, TokenLeftParens)
    if err != nil {
        return nil, err
    }

    var parameters []*ASTParameter
    /* TODO: parse parameters */

    for {
        if len(parameters) > 0 {
            next, err := tokens.Next()
            if err != nil {
                return nil, err
            }

            if next.Kind != TokenComma {
                break
            }

            tokens.Consume()
        }

        next, err := tokens.Next()
        if err != nil {
            return nil, err
        }

        if next.Kind == TokenRightParens {
            break
        }

        type_, err := parseTypeNode(tokens)
        if err != nil {
            return nil, err
        }

        name, err := tokens.Consume()
        if err != nil {
            return nil, err
        }

        if name.Kind != TokenIdentifier {
            return nil, fmt.Errorf("expected an identifier but got %v", name.String())
        }

        parameters = append(parameters, &ASTParameter{Type: type_, Name: name.Value})
    }

    err = consumeToken(tokens, TokenRightParens)
    if err != nil {
        return nil, err
    }

    return parameters, nil
}

func parseMethod(tokens *TokenStream) (*ASTMethod, error) {
    err := consumeToken(tokens, TokenMethod)
    if err != nil {
        return nil, err
    }

    typeNode, err := parseTypeNode(tokens)
    if err != nil {
        return nil, err
    }

    name, err := tokens.Consume()
    if err != nil {
        return nil, err
    }

    if name.Kind != TokenIdentifier {
        return nil, fmt.Errorf("expected an identifier for the method name but got %v", name.String())
    }

    parameters, err := parseParameterList(tokens)
    if err != nil {
        return nil, err
    }

    body, err := parseBlock(tokens)
    if err != nil {
        return nil, err
    }

    return &ASTMethod{
        ReturnType: typeNode,
        Name: name.Value,
        Parameters: parameters,
        Body: body,
    }, nil
}

func parseFunction(tokens *TokenStream) (*ASTFunction, error) {
    function, err := tokens.Consume()
    if err != nil {
        return nil, err
    }

    if function.Kind != TokenFunction {
        return nil, fmt.Errorf("expected 'function' but got %v", function.String())
    }

    typeNode, err := parseTypeNode(tokens)
    if err != nil {
        return nil, err
    }

    name, err := tokens.Consume()
    if err != nil {
        return nil, err
    }

    if name.Kind != TokenIdentifier {
        return nil, fmt.Errorf("expected an identifier for the function name but got %v", name.String())
    }

    parameters, err := parseParameterList(tokens)
    if err != nil {
        return nil, err
    }

    body, err := parseBlock(tokens)
    if err != nil {
        return nil, err
    }

    return &ASTFunction{
        ReturnType: typeNode,
        Name: name.Value,
        Parameters: parameters,
        Body: body,
    }, nil
}

func parseConstructor(tokens *TokenStream) (*ASTConstructor, error) {
    err := consumeToken(tokens, TokenConstructor)
    if err != nil {
        return nil, err
    }

    /* the class name must be the same as the file */
    class, err := tokens.Consume()
    if err != nil {
        return nil, err
    }

    if class.Kind != TokenIdentifier {
        return nil, fmt.Errorf("expected an identifier but found %v", class)
    }

    /* I think name always has to be 'new', so we could check it here */
    name, err := tokens.Consume()
    if err != nil {
        return nil, err
    }

    if name.Kind != TokenIdentifier {
        return nil, fmt.Errorf("expected an identifier but found %v", name)
    }

    parameters, err := parseParameterList(tokens)
    if err != nil {
        return nil, err
    }

    body, err := parseBlock(tokens)
    if err != nil {
        return nil, err
    }

    return &ASTConstructor{
        Class: class.Value,
        Name: name.Value,
        Parameters: parameters,
        Body: body,
    }, nil
}

func parseClassBody(tokens *TokenStream) ([]ASTNode, error) {
    var out []ASTNode

    for {
        first, err := tokens.Next()
        if err != nil {
            return nil, fmt.Errorf("out of tokens")
        }

        switch first.Kind {
            case TokenStatic:
                static, err := parseStaticDeclaration(tokens)
                if err != nil {
                    return nil, err
                }
                out = append(out, static)
            case TokenField:
                field, err := parseFieldDeclaration(tokens)
                if err != nil {
                    return nil, err
                }
                out = append(out, field)
            case TokenMethod:
                method, err := parseMethod(tokens)
                if err != nil {
                    return nil, err
                }
                out = append(out, method)
            case TokenFunction:
                function, err := parseFunction(tokens)
                if err != nil {
                    return nil, err
                }
                out = append(out, function)
            case TokenConstructor:
                constructor, err := parseConstructor(tokens)
                if err != nil {
                    return nil, err
                }
                out = append(out, constructor)
            default:
                return out, nil
        }
    }
}

func parseClass(tokens *TokenStream) (*ASTClass, error) {
    class, err := tokens.Consume()
    if err != nil {
        return nil, fmt.Errorf("expected a 'class' keyword: %v", err)
    }

    if class.Kind != TokenClass {
        return nil, fmt.Errorf("expected a 'class' keyword but got %v", class.String())
    }

    name, err := tokens.Consume()

    if err != nil {
        return nil, fmt.Errorf("expected an identifier to follow the 'class' keyword: %v", err)
    }

    if name.Kind != TokenIdentifier {
        return nil, fmt.Errorf("expected an identifier to follow the 'class' keyword: %v", name.String())
    }

    err = consumeToken(tokens, TokenLeftCurly)
    if err != nil {
        return nil, fmt.Errorf("expected a '{' after the class name: %v", err)
    }

    classElements, err := parseClassBody(tokens)
    if err != nil {
        return nil, fmt.Errorf("unable to parse class body: %v", err)
    }

    err = consumeToken(tokens, TokenRightCurly)
    if err != nil {
        return nil, fmt.Errorf("expected a '}' to close the class body: %v", err)
    }

    return &ASTClass{
        Name: name.Value,
        Body: classElements,
    }, nil
}
