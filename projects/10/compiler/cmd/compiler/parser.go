package main

import (
    "fmt"
    "io"
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
)

type ASTNode interface {
    Kind() Kind
}

type ASTClass struct {
    ASTNode

    Name string
    Body []ASTNode
}

func (ast *ASTClass) Kind() Kind {
    return ASTKindClass
}

type ASTType struct {
    ASTNode
    /* will either be an int, char, boolean, or identifier */
    Type ASTNode
}

func (ast *ASTType) Kind() Kind {
    return ASTKindType
}

type ASTFunction struct {
    ASTNode
    ReturnType *ASTType
    Name string
    Parameters []*ASTParameter
    Body *ASTBlock
}

func (ast *ASTFunction) Kind() Kind {
    return ASTKindFunction
}

type ASTParameter struct {
    ASTNode
    /* TODO */
}

func (ast *ASTParameter) Kind() Kind {
    return ASTKindParameter
}

type ASTBlock struct {
    ASTNode
    Statements []ASTNode
}

func (ast *ASTBlock) Kind() Kind {
    return ASTKindBlock
}

type ASTVar struct {
    ASTNode
    Type *ASTType
    Name string
}

func (ast *ASTVar) Kind() Kind {
    return ASTKindVar
}

type ASTLet struct {
    ASTNode
}

func (ast *ASTLet) Kind() Kind {
    return ASTKindLet
}

type ASTDo struct {
    ASTNode
}

func (ast *ASTDo) Kind() Kind {
    return ASTKindDo
}

type ASTReturn struct {
    ASTNode
}

func (ast *ASTReturn) Kind() Kind {
    return ASTKindReturn
}

type ASTStatic struct {
    ASTNode
    Type *ASTType
    Names []*ASTIdentifier
}

func (ast *ASTStatic) Kind() Kind {
    return ASTKindStaticDeclaration
}

type ASTIdentifier struct {
    ASTNode
    Name string
}

func (ast *ASTIdentifier) Kind() Kind {
    return ASTKindIdentifier
}

type ASTField struct {
    ASTNode
}

func (ast *ASTField) Kind() Kind {
    return ASTKindField
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

    if lexerError != nil {
        return nil, lexerError
    }

    return class, err
}

func consumeToken(tokens *TokenStream, kind TokenKind) error {
    token, err := tokens.Consume()
    if err != nil {
        return fmt.Errorf("out of tokens")
    }

    if token.Kind != kind {
        return fmt.Errorf("expected token %v but found %v", kind.Name(), token.Kind.Name())
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

    return nil, fmt.Errorf("expected a type to be one of int, char, boolean, or identifier")
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

func parseFieldDeclaration(tokens *TokenStream) (*ASTField, error) {
    return nil, nil
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

    name, err := tokens.Consume()
    if err != nil {
        return nil, err
    }

    if name.Kind != TokenIdentifier {
        return nil, fmt.Errorf("expected a var declaration to have an identifier but got %v", name.String())
    }

    err = consumeToken(tokens, TokenSemicolon)
    if err != nil {
        return nil, err
    }

    return &ASTVar{
        Type: typeNode,
        Name: name.Value,
    }, nil
}

func parseLetDeclaration(tokens *TokenStream) (*ASTLet, error) {
    return nil, fmt.Errorf("unimplemented")
}

func parseDo(tokens *TokenStream) (*ASTDo, error) {
    return nil, fmt.Errorf("unimplemented")
}

func parseReturn(tokens *TokenStream) (*ASTReturn, error) {
    return nil, fmt.Errorf("unimplemented")
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
                letDeclaration, err := parseLetDeclaration(tokens)
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

    err = consumeToken(tokens, TokenLeftParens)
    if err != nil {
        return nil, err
    }

    /* TODO: parse parameters */
    var parameters []*ASTParameter

    err = consumeToken(tokens, TokenRightParens)
    if err != nil {
        return nil, err
    }

    body, err := parseBlock(tokens)

    return &ASTFunction{
        ReturnType: typeNode,
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
            case TokenFunction:
                function, err := parseFunction(tokens)
                if err != nil {
                    return nil, err
                }
                out = append(out, function)
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
        return nil, fmt.Errorf("expected a 'class' keyword but got %v", class)
    }

    name, err := tokens.Consume()

    if err != nil {
        return nil, fmt.Errorf("expected an identifier to follow the 'class' keyword: %v", err)
    }

    if name.Kind != TokenIdentifier {
        return nil, fmt.Errorf("expected an identifier to follow the 'class' keyword: %v", name)
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
