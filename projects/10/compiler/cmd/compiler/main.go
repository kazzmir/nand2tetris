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
    TokenIdentifier
)

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
    if c == ' ' || c == '\t' {
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
    emit bool
}

func (literal *LiteralMachine) Consume(c byte) bool {
    // fmt.Printf("literal '%v' consuming %v\n", literal.Literal, c)

    if literal.position < len(literal.Literal) && literal.Literal[literal.position] == c {
        literal.position += 1
        return true
    }

    literal.stopped = true
    literal.emit = !unicode.IsLetter(rune(c))

    return false
}

func (literal *LiteralMachine) Reset() {
    literal.stopped = false
    literal.position = 0
    literal.emit = false
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
        emit: false,
    }
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
        buildLiteralMachine("this", TokenThis),
        buildLiteralMachine("while", TokenWhile),
    }
}

func breakTies(tokens []Token) Token {
    return tokens[0]
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

                // fmt.Printf("Check token %v\n", token)

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
            out = append(out, token)

            partial = nil
            start = end

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

/*
machine.consume(c) -> token

ThisMachine.emit(ThisToken())
ThisMachine.emit(ThisToken())
IdentifierMachine.emit(IdentifierToken())

"thisthithis foobar"

ThisThi.emit(ThisThiToken()) = "thisthi"
*/

func compile(path string) error {
    /*
    identifier := regex.Compile("\\w+")
    number := regex.Compile("\\d+")
    keywordThis := regex.Compile("this")
    keywordThat := regex.Compile("that")
    */
    return nil
}

func compileAll(paths []string) error {
    for _, path := range paths {
        err := compile(path)
        if err != nil {
            return err
        }
    }

    return nil
}

func main(){
    // TestL()
    /*
    if len(os.Args) == 1 {
        fmt.Printf("Give a directory of .jack files or a list of .jack files")
        return
    } else {
        err := compileAll(os.Args[1:])
        if err != nil {
            fmt.Printf("Error: %v\n", err)
        }
    }
    */
}
