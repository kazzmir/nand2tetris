package main

import (
    _ "os"
    "io"
    "fmt"
    "bufio"
    "errors"
    "unicode"
)

type LexerStateMachine interface {
    Consume(c byte) bool
    Alive() bool
    Reset()
    Token(start uint64, end uint64) (Token, error)
}

type TokenKind int

const (
    TokenThis TokenKind = iota
    TokenWhile
    TokenWhitespace
    TokenComment
    TokenIf
    TokenIdentifier
    TokenNumber
    TokenPlus
    TokenMethod
    TokenVoid
    TokenLeftParens
    TokenRightParens
    TokenLeftCurly
    TokenRightCurly
    TokenEquals
    TokenDot
    TokenSemicolon
    TokenReturn
)

func (kind *TokenKind) Precedence() int {
    switch *kind {
        /* keywords should have the highest precedence */
        case TokenThis, TokenWhile, TokenMethod,
             TokenVoid, TokenIf, TokenReturn: return 10
        case TokenLeftParens, TokenRightParens,
             TokenLeftCurly, TokenRightCurly: return 1
        case TokenEquals, TokenDot, TokenSemicolon: return 1
        case TokenIdentifier: return 1
        case TokenComment: return 0
        case TokenWhitespace: return 0
        case TokenPlus: return 1
    }

    return 0
}

func removeWhitespaceTokens(tokens []Token) []Token {
    var out []Token = nil

    for _, token := range tokens {
        if token.Kind != TokenWhitespace && token.Kind != TokenComment {
            out = append(out, token)
        }
    }

    return out
}

type Token struct {
    Kind TokenKind
    Value string
    /* source location within the input */
    Start uint64
    End uint64
}

type WhiteSpaceMachine struct {
    LexerStateMachine
    stopped bool
}

func (space *WhiteSpaceMachine) Consume(c byte) bool {
    if c == ' ' || c == '\t' || c == '\n' {
        return true
    }

    space.stopped = true

    return false
}

func (space *WhiteSpaceMachine) Alive() bool {
    return !space.stopped
}

func (space *WhiteSpaceMachine) Reset() {
    space.stopped = false
}

func (space *WhiteSpaceMachine) Token(start uint64, end uint64) (Token, error) {
    return Token{
        Kind: TokenWhitespace,
        Value: " ",
        Start: start,
        End: end,
    }, nil
}

type LiteralMachine struct {
    LexerStateMachine
    Literal string
    Kind TokenKind

    position int
    stopped bool
}

func (literal *LiteralMachine) Consume(c byte) bool {
    if literal.position < len(literal.Literal) && literal.Literal[literal.position] == c {
        literal.position += 1
        return true
    }

    literal.stopped = true

    return false
}

func (literal *LiteralMachine) Reset() {
    literal.stopped = false
    literal.position = 0
}

var NoToken error = errors.New("no-token")

func (literal *LiteralMachine) Token(start uint64, end uint64) (Token, error) {
    if literal.position == len(literal.Literal) /* && literal.emit */ {
        return Token{
            Kind: literal.Kind,
            Value: "",
            // Value: literal.Literal,
            Start: start,
            End: end,
        }, nil
    } else {
        return Token{}, NoToken
    }
}

func (literal *LiteralMachine) Alive() bool {
    return !literal.stopped
}

func buildLiteralMachine(literal string, kind TokenKind) LexerStateMachine {
    return &LiteralMachine{
        Literal: literal,
        Kind: kind,
        position: 0,
        stopped: false,
    }
}

type IdentifierMachine struct {
    LexerStateMachine

    Name []byte
    stopped bool
}

func (identifier *IdentifierMachine) Consume(c byte) bool {
    /* must start with a letter */
    if len(identifier.Name) == 0 {
        if unicode.IsLetter(rune(c)) {
            identifier.Name = append(identifier.Name, c)
            return true
        }
    /* then it can be numbers of letters */
    } else if unicode.IsLetter(rune(c)) || unicode.IsDigit(rune(c)) {
        identifier.Name = append(identifier.Name, c)
        return true
    }

    identifier.stopped = true
    return false
}

func (identifier *IdentifierMachine) Alive() bool {
    return !identifier.stopped
}

func (identifier *IdentifierMachine) Reset() {
    identifier.stopped = false
    identifier.Name = nil
}

func (identifier *IdentifierMachine) Token(start uint64, end uint64) (Token, error) {
    if len(identifier.Name) > 0 {
        return Token{
            Kind: TokenIdentifier,
            Value: string(identifier.Name),
            Start: start,
            End: end,
        }, nil
    }

    return Token{}, NoToken
}

type NumberMachine struct {
    LexerStateMachine
    Number []byte
    stopped bool
}

func (machine *NumberMachine) Consume(c byte) bool {
    if unicode.IsDigit(rune(c)) {
        machine.Number = append(machine.Number, c)
        return true
    }

    machine.stopped = true
    return false
}

func (machine *NumberMachine) Alive() bool {
    return !machine.stopped
}

func (machine *NumberMachine) Reset() {
    machine.stopped = false
    machine.Number = nil
}

func (machine *NumberMachine) Token(start uint64, end uint64) (Token, error) {
    if len(machine.Number) > 0 {
        return Token{
            Kind: TokenNumber,
            Value: string(machine.Number),
            Start: start,
            End: end,
        }, nil
    }

    return Token{}, NoToken
}

func makeIdentifierMachine() LexerStateMachine {
    return &IdentifierMachine{Name: nil, stopped: false}
}

func makeThisMachine() LexerStateMachine {
    return buildLiteralMachine("this", TokenThis)
}

func makeWhileMachine() LexerStateMachine {
    return buildLiteralMachine("while", TokenWhile)
}

func makeNumberMachine() LexerStateMachine {
    return &NumberMachine{stopped: false}
}

func makeMethodMachine() LexerStateMachine {
    return buildLiteralMachine("method", TokenMethod)
}

func makeVoidMachine() LexerStateMachine {
    return buildLiteralMachine("void", TokenVoid)
}

func makePlusMachine() LexerStateMachine {
    return buildLiteralMachine("+", TokenPlus)
}

func makeLeftParensMachine() LexerStateMachine {
    return buildLiteralMachine("(", TokenLeftParens)
}

func makeRightParensMachine() LexerStateMachine {
    return buildLiteralMachine(")", TokenRightParens)
}

func makeLeftCurlyMachine() LexerStateMachine {
    return buildLiteralMachine("{", TokenLeftParens)
}

func makeRightCurlyMachine() LexerStateMachine {
    return buildLiteralMachine("}", TokenRightCurly)
}

func makeEqualsMachine() LexerStateMachine {
    return buildLiteralMachine("=", TokenEquals)
}

func makeDotMachine() LexerStateMachine {
    return buildLiteralMachine(".", TokenDot)
}

func makeSemicolonMachine() LexerStateMachine {
    return buildLiteralMachine(";", TokenSemicolon)
}

func makeIfMachine() LexerStateMachine {
    return buildLiteralMachine("if", TokenIf)
}

func makeReturnMachine() LexerStateMachine {
    return buildLiteralMachine("return", TokenReturn)
}

/* for each state machine, call machine(c). it returns token and bool which is
 * the completed token, and true = can match more, or false = cannot match c
 *
 * "thisisaX"
 *      ^ this machine emits This("this"), false
 *          ^ identifier machine emits Identifier("thisisaX"), true
 *
 * 12+4
 * DigitMachine.consume(1).consume(2).consume(+) -> DigitToken("12")
 *
 * ThisToken("this", 2, 5)
 * IdentifierToken("thishelp", 2, 9)
 * IshToken("ish", 4, 6)
 *
 * get a list of tokens, sort by starting position, from the end of the sorted list
 * check if token n-1 is contained entirely by token n, in which case throw out n-1.
 * then check token n vs n-2, etc until token n-i starts before token n. then do the
 * same with token n-i.
 *
 * for each machine, call machine(c), it returns true/false, true for 'can consume more'
 * and false for 'cannot consume more' aka consuming c would transition to an error state.
 *
 * when all machines have reached an error state, save c for the next loop, then check
 * which machine has the longest token, using special rules to break ties like This("this")
 * should win over IdentifierToken("this").
 */

func makeLexerMachines() []LexerStateMachine {
    return []LexerStateMachine{
        &WhiteSpaceMachine{stopped: false},
        makeThisMachine(),
        makeWhileMachine(),
        makeIdentifierMachine(),
        makeNumberMachine(),
        makePlusMachine(),
        makeMethodMachine(),
        makeVoidMachine(),
        makeLeftParensMachine(),
        makeRightParensMachine(),
        makeLeftCurlyMachine(),
        makeRightCurlyMachine(),
        makeIfMachine(),
        makeEqualsMachine(),
        makeDotMachine(),
        makeSemicolonMachine(),
        makeReturnMachine(),
    }
}

func breakTies(tokens []Token) Token {
    var out Token

    precedence := -1

    for _, token := range tokens {
        if token.Kind.Precedence() > precedence {
            out = token
            precedence = token.Kind.Precedence()
        }
    }

    return out
}

func lexer(machines []LexerStateMachine, reader io.Reader) ([]Token, error) {
    var out []Token

    bufferedReader := bufio.NewReader(reader)

    c, readErr := bufferedReader.ReadByte()

    if readErr != nil && readErr == io.EOF {
        return out, nil
    }
    if readErr != nil {
        return nil, readErr
    }

    partial := []byte{c}
    var start uint64 = 0
    var end uint64 = 1

    for {
        count := 0

        var emittingMachines []LexerStateMachine

        for _, machine := range machines {
            if machine.Alive() {
                emittingMachines = append(emittingMachines, machine)
                ok := machine.Consume(c)
                if ok {
                    count += 1
                }
            }
        }

        if count == 0 {
            /* all machines died */

            var longest uint64 = 0
            var possible []Token
            for _, machine := range emittingMachines {
                length := end-1 - start
                token, err := machine.Token(start, end-1)

                if err == nil {
                    if length == 0 {
                        continue
                    }

                    if length == longest {
                        possible = append(possible, token)
                    } else if length > longest {
                        possible = []Token{token}
                        longest = length
                    }
                }
            }

            for _, machine := range machines {
                machine.Reset()
            }

            if len(possible) == 0 {
                return nil, fmt.Errorf("Could not tokenize '%v' from position %v to %v", string(partial), start, end-1)
            }

            token := breakTies(possible)
            // fmt.Printf("Parsed %v\n", token)
            out = append(out, token)

            partial = nil
            start = end - 1

            if readErr == io.EOF {
                return out, nil
            }
        } else {
            /* previously read eof, so now quit */
            if readErr == io.EOF {
                return out, nil
            }

            c, readErr = bufferedReader.ReadByte()
            end += 1

            if readErr != nil && readErr == io.EOF {
                c = 0
                continue
            }

            partial = append(partial, c)

            if readErr != nil {
                return nil, readErr
            }
        }
    }
}

func standardLexer(reader io.Reader) ([]Token, error) {
    machines := makeLexerMachines()
    return lexer(machines, reader)
}

