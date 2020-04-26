package main

import (
    "os"
    "fmt"
    "flag"
    "bufio"
    "strings"
)

type RawCode struct {
    Text string
    Line uint64
    SourceLine uint64
}

/* Represents the program in its unprocessed form, except that comments
 * and lines with only whitespace (non-code lines) are removed.
 */
type RawProgram struct {
    Code []RawCode
}

func (raw *RawProgram) AddLine(line string, sourceLine uint64){
    if strings.HasPrefix(line, "//") {
        return
    }

    trimmed := strings.TrimSpace(line)

    if len(trimmed) > 0 {
        code := RawCode{
            Text: trimmed,
            Line: uint64(len(raw.Code)),
            SourceLine: sourceLine,
        }

        raw.Code = append(raw.Code, code)
    }
}

func (raw *RawProgram) Dump() {
    for _, code := range raw.Code {
        fmt.Printf("%v: %v\n", code.Line, code.Text)
    }
}

type ParsedCode interface {
}

type Register int
const (
    ARegister Register = iota
    MRegister
    DRegister
    InvalidRegister
)

type Operation int
const (
    OperationAdd Operation = iota
    OperationSubtract
    OperationBinaryAnd
    OperationBinaryOr
    OperationNegate
    OperationNot
    InvalidOperation
)

type ParsedExpression interface {
}

type ParsedAssignment struct {
    ParsedCode

    Assign []Register
    Expression ParsedExpression
}

type ParsedProgram struct {
    Code []ParsedCode
}

func (program *ParsedProgram) Add(code ParsedCode) {
    program.Code = append(program.Code, code)
}

func parseRegister(name rune) (Register, error) {
    switch name {
        case 'A': return ARegister, nil
        case 'M': return MRegister, nil
        case 'D': return DRegister, nil
        default: return InvalidRegister, fmt.Errorf("unknown register name '%v'", name)
    }
}

func parseAssignedVariables(variables string) ([]Register, error) {
    var out []Register = nil

    for _, name := range variables {
        named, err := parseRegister(name)
        if err != nil {
            return nil, err
        }

        out = append(out, named)
    }

    if len(out) == 0 {
        return nil, fmt.Errorf("no variables found on the left hand side of an assignment")
    }

    return out, nil
}

type ParsedConstant struct {
    ParsedExpression
    Value int
}

type ParsedSingleRegister struct {
    ParsedExpression
    Register Register
}

type ParsedUnary struct {
    ParsedExpression
    Operation Operation
    Value ParsedExpression
}

type ParsedBinary struct {
    ParsedExpression
    Operation Operation
    Left ParsedExpression
    Right ParsedExpression
}

func parseNegation(expression string) (ParsedExpression, error) {
    if len(expression) == 0 {
        return nil, fmt.Errorf("no expression given after the '-' sign")
    }

    if expression[0] == '1' {
        return &ParsedConstant{Value: -1}, nil
    }

    register, err := parseRegister(rune(expression[0]))
    if err != nil {
        return nil, err
    }

    return ParsedUnary{Operation: OperationNegate, Value: register}, nil
}

func parseNot(expression string) (ParsedExpression, error) {
    if len(expression) == 0 {
        return nil, fmt.Errorf("no expression given after '!'")
    }

    register, err := parseRegister(rune(expression[0]))

    if err != nil {
        return nil, err
    }

    return ParsedUnary{Operation: OperationNot, Value: register}, nil
}

func maybeParseOperation(register Register, expression string) (ParsedExpression, error) {
    if len(expression) == 0 {
        return ParsedSingleRegister{Register: register}, nil
    }

    if len(expression) != 2 {
        return nil, fmt.Errorf("expected an operation and a value: '%v'", expression)
    }

    var op = expression[0]
    var right = expression[1]

    var operation Operation

    switch op {
        case '+': operation = OperationAdd
        case '-': operation = OperationSubtract
        case '&': operation = OperationBinaryAnd
        case '|': operation = OperationBinaryOr
        default: return nil, fmt.Errorf("invalid operation '%v'", expression[0])
    }

    var rightValue ParsedExpression

    switch right {
        case '0': rightValue = ParsedConstant{Value: 0}
        case '1': rightValue = ParsedConstant{Value: 1}
        default:
            var err error
            rightValue, err = parseRegister(rune(right))
            if err != nil {
                return nil, err
            }
    }

    if rightValue == nil {
        return nil, fmt.Errorf("expected a value to follow the operation")
    }

    return ParsedBinary{Operation: operation, Left: register, Right: rightValue}, nil
}

func parseExpression(expression string) (ParsedExpression, error) {
    /* expression := variable op one-or-variable | 0 | 1 | -1 | variable | ! variable | - variable
     * variable = A | M | D
     * one-or-variable = variable | 0 | 1
     * op = + | - | & | '|'
     *
     * the same variable cannot appear twice, such as D+D
     */

    if len(expression) == 0 {
        return nil, fmt.Errorf("no expression given")
    }

    switch expression[0] {
        case '0': return &ParsedConstant{Value: 0}, nil
        case '1': return &ParsedConstant{Value: 1}, nil
        case '-': return parseNegation(expression[1:])
        case '!': return parseNot(expression[1:])
    }

    register1, err := parseRegister(rune(expression[0]))
    if err != nil {
        return nil, err
    }

    return maybeParseOperation(register1, expression[1:])
}

func parseAssignment(raw RawCode) (ParsedAssignment, error) {
    parts := strings.Split(raw.Text, "=")
    if len(parts) != 2 {
        return ParsedAssignment{}, fmt.Errorf("Error line %v: multiple '=' characters in an assignment statement", raw.SourceLine)
    }

    assigned, err := parseAssignedVariables(parts[0])
    if err != nil {
        return ParsedAssignment{}, fmt.Errorf("Error line %v: '%v' %v", raw.SourceLine, raw.Text, err)
    }
    expression, err := parseExpression(strings.TrimSpace(parts[1]))
    if err != nil {
        return ParsedAssignment{}, fmt.Errorf("Error line %v: '%v' %v", raw.SourceLine, raw.Text, err)
    }

    return ParsedAssignment{
        Assign: assigned,
        Expression: expression,
    }, nil
}

func parse(raw RawProgram) (ParsedProgram, error) {
    var parsed ParsedProgram

    for _, code := range raw.Code {
        if strings.ContainsRune(code.Text, '=') {
            converted, err := parseAssignment(code)
            if err != nil {
                return parsed, err
            }
            parsed.Add(&converted)
        }
    }

    return parsed, nil
}

func process(path string) error {
    fmt.Printf("Assembling '%v'\n", path)

    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()

    var rawProgram RawProgram

    scanner := bufio.NewScanner(file)
    var sourceLine uint64
    for scanner.Scan() {
        line := scanner.Text()
        // fmt.Printf("%v: %v\n", i, line)
        sourceLine += 1

        rawProgram.AddLine(line, sourceLine)
    }

    err = scanner.Err()
    if err != nil {
        return err
    }

    rawProgram.Dump()

    parsed, err := parse(rawProgram)
    if err != nil {
        return err
    }

    _ = parsed

    return nil
}

func help() {
    fmt.Printf(`Help:
 $ assembler file.asm ...

nand2tetris assembler by Jon Rafkind (jon@rafkind.com)
`)
}

func main(){
    flag.Parse()

    rest := os.Args[len(os.Args) - flag.NArg():]
    if len(rest) == 0 {
        fmt.Printf("Give a file to process\n\n")
        help()
        return
    }

    for _, path := range rest {
        err := process(path)
        if err != nil {
            fmt.Printf("Error: Could not process '%v': %v\n", path, err)
        }
    }
}
