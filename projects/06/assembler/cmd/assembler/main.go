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

type ParsedExpression struct {
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

func parseExpression(expression string) (ParsedExpression, error) {
    return ParsedExpression{}, fmt.Errorf("unimplemented")
}

func parseAssignment(raw RawCode) (ParsedAssignment, error) {
    parts := strings.Split(raw.Text, "=")
    if len(parts) != 2 {
        return ParsedAssignment{}, fmt.Errorf("Error line %v: multiple '=' characters in an assignment statement", raw.SourceLine)
    }

    assigned, err := parseAssignedVariables(parts[0])
    if err != nil {
        return ParsedAssignment{}, fmt.Errorf("Error line %v: %v", raw.SourceLine, err)
    }
    expression, err := parseExpression(parts[1])
    if err != nil {
        return ParsedAssignment{}, fmt.Errorf("Error line %v: %v", raw.SourceLine, err)
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
