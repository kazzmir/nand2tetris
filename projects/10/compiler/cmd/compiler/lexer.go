package main

import (
    _ "os"
    "io"
    "fmt"
    "bufio"
    "bytes"
    "errors"
    "unicode"
)

type LexerStateMachine interface {
    Consume(c byte) bool
    Alive() bool
    Reset()
    Token(line uint64, start uint64, end uint64) (Token, error)
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
    TokenDivision
    TokenVar
    TokenComma
    TokenString
    TokenLeftBracket
    TokenRightBracket
    TokenNegation
    TokenTrue
    TokenFalse
    TokenMultiply
    TokenOr
    TokenAnd
    TokenNot
    TokenLessThan
    TokenGreaterThan
)

func (kind *TokenKind) Precedence() int {
    switch *kind {
        /* keywords should have the highest precedence */
        case TokenThis, TokenWhile, TokenMethod,
             TokenVoid, TokenIf, TokenReturn,
             TokenVar, TokenTrue, TokenFalse: return 10
        case TokenLeftParens, TokenRightParens,
             TokenLeftCurly, TokenRightCurly: return 1
        case TokenEquals, TokenDot, TokenSemicolon,
             TokenComma, TokenString: return 1
        case TokenIdentifier: return 1
        case TokenComment: return 0
        case TokenWhitespace: return 0
        case TokenPlus, TokenDivision: return 1
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
    Line uint64
    Start uint64
    End uint64
}

type WhiteSpaceMachine struct {
    LexerStateMachine
    stopped bool
}

func (space *WhiteSpaceMachine) Consume(c byte) bool {
    if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
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

func (space *WhiteSpaceMachine) Token(line uint64, start uint64, end uint64) (Token, error) {
    return Token{
        Kind: TokenWhitespace,
        Value: " ",
        Line: line,
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

func (literal *LiteralMachine) Token(line uint64, start uint64, end uint64) (Token, error) {
    if literal.position == len(literal.Literal) /* && literal.emit */ {
        return Token{
            Kind: literal.Kind,
            Value: "",
            // Value: literal.Literal,
            Line: line,
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

    Name bytes.Buffer
    stopped bool
}

func (identifier *IdentifierMachine) Consume(c byte) bool {
    /* must start with a letter */
    if identifier.Name.Len() == 0 {
        if unicode.IsLetter(rune(c)) {
            identifier.Name.WriteByte(c)
            return true
        }
    /* then it can be numbers of letters */
    } else if unicode.IsLetter(rune(c)) || unicode.IsDigit(rune(c)) {
        identifier.Name.WriteByte(c)
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
    identifier.Name.Reset()
    identifier.Name.Grow(20)
}

func (identifier *IdentifierMachine) Token(line uint64, start uint64, end uint64) (Token, error) {
    if identifier.Name.Len() > 0 {
        return Token{
            Kind: TokenIdentifier,
            Value: string(identifier.Name.Bytes()),
            Line: line,
            Start: start,
            End: end,
        }, nil
    }

    return Token{}, NoToken
}

type NumberMachine struct {
    LexerStateMachine
    Number bytes.Buffer
    stopped bool
}

func (machine *NumberMachine) Consume(c byte) bool {
    if unicode.IsDigit(rune(c)) {
        machine.Number.WriteByte(c)
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
    machine.Number.Reset()
    machine.Number.Grow(5)
}

func (machine *NumberMachine) Token(line uint64, start uint64, end uint64) (Token, error) {
    if machine.Number.Len() > 0 {
        return Token{
            Kind: TokenNumber,
            Value: string(machine.Number.Bytes()),
            Line: line,
            Start: start,
            End: end,
        }, nil
    }

    return Token{}, NoToken
}

type SingleCommentMachine struct {
    LexerStateMachine
    Slashes int
    Newline bool
    stopped bool
}

func (machine *SingleCommentMachine) Token(line uint64, start uint64, end uint64) (Token, error) {
    return Token{
        Kind: TokenWhitespace,
        Value: " ",
        Line: line,
        Start: start,
        End: end,
    }, nil
}

func (machine *SingleCommentMachine) Consume(c byte) bool {
    if machine.Slashes == 0 {
        if c == '/' {
            machine.Slashes = 1
            return true
        }
    } else if machine.Slashes == 1 {
        if c == '/' {
            machine.Slashes = 2
            return true
        }
    } else {
        if machine.Newline {
            machine.stopped = true
            return false
        }

        if c == '\n' {
            machine.Newline = true
        }

        return true
    }

    machine.stopped = true
    return false
}

func (machine *SingleCommentMachine) Alive() bool {
    return !machine.stopped
}

func (machine *SingleCommentMachine) Reset() {
    machine.Slashes = 0
    machine.Newline = false
    machine.stopped = false
}

type BlockCommentMachine struct {
    LexerStateMachine
    OpenComment int // will be 1 if only '/' is seen, and 2 if '/*' is seen
    CloseComment int // will be 1 if only '*' is seen, and 2 if '*/' is seen
    stopped bool
}

func (machine *BlockCommentMachine) Alive() bool {
    return !machine.stopped
}

func (machine *BlockCommentMachine) Reset() {
    machine.OpenComment = 0
    machine.CloseComment = 0
    machine.stopped = false
}

func (machine *BlockCommentMachine) Consume(c byte) bool {
    if machine.OpenComment == 0 {
        if c == '/' {
            machine.OpenComment = 1
            return true
        }
    } else if machine.OpenComment == 1 {
        if c == '*' {
            machine.OpenComment = 2
            return true
        } else {
            /* saw something other than a '*' so this is not a comment */
            machine.stopped = true
            return false
        }
    } else if machine.CloseComment == 0 {
        if c == '*' {
            machine.CloseComment = 1
        }

        return true
    } else if machine.CloseComment == 1 {
        if c == '/' {
            machine.CloseComment = 2
        } else {
            /* saw a '*' but didnt see a '/', so this is not a close comment */
            machine.CloseComment = 0
        }

        return true
    } else if machine.CloseComment == 2 {
        /* saw a '*' and '/' so this comment is done */
        machine.stopped = true
        return false
    }

    return false
}

func (machine *BlockCommentMachine) Token(line uint64, start uint64, end uint64) (Token, error) {
    return Token{
        Kind: TokenWhitespace,
        Value: " ",
        Line: line,
        Start: start,
        End: end,
    }, nil
}

type StringMachine struct {
    LexerStateMachine
    Quote int // 1 for the starting quote, 2 for the ending quote
    Text bytes.Buffer // actual string contents
    stopped bool
}

func (machine *StringMachine) Alive() bool {
    return !machine.stopped
}

func (machine *StringMachine) Consume(c byte) bool {
    if machine.Quote == 0 {
        if c == '"' {
            machine.Quote = 1
            return true
        }
    } else if machine.Quote == 1 {
        if c == '"' {
            machine.Quote = 2
        } else {
            machine.Text.WriteByte(c)
        }

        return true
    }

    machine.stopped = true
    return false
}

func (machine *StringMachine) Token(line uint64, start uint64, end uint64) (Token, error) {
    if machine.Quote == 2 {
        return Token{
            Kind: TokenString,
            Value: string(machine.Text.Bytes()),
            Line: line,
            Start: start,
            End: end,
        }, nil
    }

    return Token{}, fmt.Errorf("did not parse a string")
}

func (machine *StringMachine) Reset() {
    machine.Quote = 0
    machine.stopped = false
}

func makeIdentifierMachine() LexerStateMachine {
    machine := &IdentifierMachine{}
    machine.Reset()
    return machine
}

func makeThisMachine() LexerStateMachine {
    return buildLiteralMachine("this", TokenThis)
}

func makeWhileMachine() LexerStateMachine {
    return buildLiteralMachine("while", TokenWhile)
}

func makeNumberMachine() LexerStateMachine {
    machine := &NumberMachine{}
    machine.Reset()
    return machine
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

func makeVarMachine() LexerStateMachine {
    return buildLiteralMachine("var", TokenVar)
}

func makeCommaMachine() LexerStateMachine {
    return buildLiteralMachine(",", TokenComma)
}

func makeDivisionMachine() LexerStateMachine {
    return buildLiteralMachine("/", TokenReturn)
}

func makeLeftBracketMachine() LexerStateMachine {
    return buildLiteralMachine("[", TokenLeftBracket)
}

func makeRightBracketMachine() LexerStateMachine {
    return buildLiteralMachine("]", TokenRightBracket)
}

func makeNegationMachine() LexerStateMachine {
    return buildLiteralMachine("-", TokenNegation)
}

func makeTrueMachine() LexerStateMachine {
    return buildLiteralMachine("true", TokenTrue)
}

func makeFalseMachine() LexerStateMachine {
    return buildLiteralMachine("false", TokenFalse)
}

func makeMultiplyMachine() LexerStateMachine {
    return buildLiteralMachine("*", TokenMultiply)
}

func makeOrMachine() LexerStateMachine {
    return buildLiteralMachine("|", TokenOr)
}

func makeAndMachine() LexerStateMachine {
    return buildLiteralMachine("&", TokenAnd)
}

func makeNotMachine() LexerStateMachine {
    return buildLiteralMachine("~", TokenNot)
}

func makeLessThanMachine() LexerStateMachine {
    return buildLiteralMachine("<", TokenLessThan)
}

func makeGreaterThanMachine() LexerStateMachine {
    return buildLiteralMachine(">", TokenGreaterThan)
}

func makeSingleCommentMachine() LexerStateMachine {
    machine := &SingleCommentMachine{}
    machine.Reset()
    return machine
}

func makeBlockCommentMachine() LexerStateMachine {
    machine := &BlockCommentMachine{}
    machine.Reset()
    return machine
}

func makeStringMachine() LexerStateMachine {
    machine := &StringMachine{}
    machine.Reset()
    return machine
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
        makeDivisionMachine(),
        makeSingleCommentMachine(),
        makeBlockCommentMachine(),
        makeVarMachine(),
        makeCommaMachine(),
        makeStringMachine(),
        makeLeftBracketMachine(),
        makeRightBracketMachine(),
        makeNegationMachine(),
        makeTrueMachine(),
        makeFalseMachine(),
        makeMultiplyMachine(),
        makeOrMachine(),
        makeAndMachine(),
        makeNotMachine(),
        makeLessThanMachine(),
        makeGreaterThanMachine(),
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

func lexer(machines []LexerStateMachine, reader io.Reader, emitToken chan Token) error {
    defer close(emitToken)

    bufferedReader := bufio.NewReader(reader)

    c, readErr := bufferedReader.ReadByte()

    if readErr != nil && readErr == io.EOF {
        return nil
    }
    if readErr != nil {
        return readErr
    }

    var partial bytes.Buffer
    partial.WriteByte(c)
    var start uint64 = 0
    var end uint64 = 1

    emittingMachines := make([]bool, len(machines))
    aliveMachines := make([]bool, len(machines))
    for i := 0; i < len(aliveMachines); i++ {
        aliveMachines[i] = true
    }

    var line uint64 = 1
    inLine := 1

    if c == '\n' {
        line += 1
    }

    for {
        count := 0

        for i, machine := range machines {
            // if machine.Alive() {
            if aliveMachines[i] {
                emittingMachines[i] = true
                // emittingMachines = append(emittingMachines, machine)
                ok := machine.Consume(c)
                aliveMachines[i] = ok
                if ok {
                    count += 1
                }
            } else {
                emittingMachines[i] = false
            }
        }

        if count == 0 {
            /* all machines died */

            emitters := 0
            emitter := 0
            for i, ok := range emittingMachines {
                if ok {
                    emitters += 1
                    emitter = i
                }
            }

            /* In a lot of cases there is only one emitter, so just find that machine directly */
            if emitters == 1 {
                machine := machines[emitter]
                token, err := machine.Token(line, start, end-1)
                if err != nil {
                    for i := 0; i < 10; i++ {
                        b, err := bufferedReader.ReadByte()
                        if err != nil || b == '\n'{
                            break
                        } else {
                            partial.WriteByte(b)
                        }
                    }

                    return fmt.Errorf("Could not tokenize '%v' at line %v position %v", string(partial.Bytes()), line, inLine)
                }
                emitToken <- token
            } else {
                var longest uint64 = 0
                var possible []Token
                for i, ok := range emittingMachines {
                    if !ok {
                        continue
                    }
                    machine := machines[i]

                    length := end-1 - start
                    token, err := machine.Token(line, start, end-1)

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

                if len(possible) == 0 {
                    for i := 0; i < 10; i++ {
                        b, err := bufferedReader.ReadByte()
                        if err != nil || b == '\n' {
                            break
                        } else {
                            partial.WriteByte(b)
                        }
                    }

                    return fmt.Errorf("Could not tokenize '%v' at line %v position %v", string(partial.Bytes()), line, inLine)
                }

                token := breakTies(possible)
                emitToken <- token
            }

            for i, machine := range machines {
                machine.Reset()
                aliveMachines[i] = true
            }

            // fmt.Printf("Parsed %v\n", token)
            // out = append(out, token)

            partial.Reset()
            partial.Grow(10)
            start = end - 1

            if readErr == io.EOF {
                return nil
            }
        } else {
            /* previously read eof, so now quit */
            if readErr == io.EOF {
                return nil
            }

            c, readErr = bufferedReader.ReadByte()
            end += 1

            if readErr != nil && readErr == io.EOF {
                c = 0
                continue
            }

            partial.WriteByte(c)

            if c == '\n' {
                line += 1
                inLine = 1
            } else {
                inLine += 1
            }

            if readErr != nil {
                return readErr
            }
        }
    }
}

func standardLexer(reader io.Reader, out chan Token) error {
    machines := makeLexerMachines()
    return lexer(machines, reader, out)
}

func lexerTokenSequence(machines []LexerStateMachine, reader io.Reader) ([]Token, error) {
    output := make(chan Token, 1000)
    var err error
    go func(){
        err = lexer(machines, reader, output)
    }()

    var out []Token
    for token := range output {
        out = append(out, token)
    }

    return out, err

}

func standardLexerTokenSequence(reader io.Reader) ([]Token, error) {
    return lexerTokenSequence(makeLexerMachines(), reader)
}
