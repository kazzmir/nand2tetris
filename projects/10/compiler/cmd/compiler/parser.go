package main

import (
    "fmt"
)

type Kind uint32

const (
    ASTKindClass Kind = iota
)

type ASTNode interface {
    Kind() Kind
}

type ASTClass struct {
    ASTNode

    Name string
}

func (ast *ASTClass) Kind() Kind {
    return ASTKindClass
}

func parse(tokens chan Token) (ASTNode, error) {
    filteredTokens := make(chan Token, 1000)

    go func(){
        for token := range tokens {
            if token.Kind != TokenWhitespace {
                filteredTokens <- token
            }
        }

        close(filteredTokens)
    }()

    return parseClass(filteredTokens)
}

func consumeToken(tokens chan Token, kind TokenKind) error {
    token, ok := <-tokens
    if !ok {
        return fmt.Errorf("out of tokens")
    }

    if token.Kind != kind {
        return fmt.Errorf("expected token %v but found %v", kind.Name(), token.Kind.Name())
    }

    return nil
}

func parseClass(tokens chan Token) (ASTNode, error) {
    class, ok := <-tokens
    if !ok {
        return nil, fmt.Errorf("expected a 'class' keyword")
    }

    if class.Kind != TokenClass {
        return nil, fmt.Errorf("expected a 'class' keyword but got %v", class)
    }

    name, ok := <-tokens

    if !ok {
        return nil, fmt.Errorf("out of tokens: expected an identifier to follow the 'class' keyword")
    }

    if name.Kind != TokenIdentifier {
        return nil, fmt.Errorf("expected an identifier to follow the 'class' keyword: %v", name)
    }

    err := consumeToken(tokens, TokenLeftCurly)
    if err != nil {
        return nil, fmt.Errorf("expected a '{' after the class name: %v", err)
    }

    err = consumeToken(tokens, TokenRightCurly)
    if err != nil {
        return nil, fmt.Errorf("expected a '}' to close the class body: %v", err)
    }

    return &ASTClass{
        Name: name.Value,
    }, nil
}
