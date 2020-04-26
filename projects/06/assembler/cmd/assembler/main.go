package main

import (
    "os"
    "fmt"
    "flag"
    "bufio"
    "strings"
    "strconv"
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
    ToBinaryString() string
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
    UsesARegister() bool
    ToComputeBinaryString() string
}

type ParsedAssignment struct {
    ParsedCode

    Assign []Register
    Expression ParsedExpression
}

func (assignment *ParsedAssignment) ToBinaryString() string {
    var out strings.Builder
    out.WriteString("111")

    usesA := assignment.Expression.UsesARegister()

    if usesA {
        out.WriteRune('0')
    } else {
        out.WriteRune('1')
    }

    out.WriteString(assignment.Expression.ToComputeBinaryString())

    assignD := false
    assignM := false
    assignA := false

    for _, register := range assignment.Assign {
        switch register {
            case DRegister: assignD = true
            case ARegister: assignA = true
            case MRegister: assignM = true
        }
    }

    if assignA {
        out.WriteRune('1')
    } else {
        out.WriteRune('0')
    }

    if assignD {
        out.WriteRune('1')
    } else {
        out.WriteRune('0')
    }

    if assignM {
        out.WriteRune('1')
    } else {
        out.WriteRune('0')
    }

    /* no jump for an assignment */
    out.WriteString("000")

    return out.String()
}

type ParsedProgram struct {
    Code []ParsedCode
}

func (program *ParsedProgram) DumpAsBinaryString() {
    for _, code := range program.Code {
        fmt.Printf("%v\n", code.ToBinaryString())
    }
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

func (register *ParsedSingleRegister) ToComputeBinaryString() string {
    switch register.Register {
        case ARegister, MRegister:
            return "110000"
        case DRegister:
            return "001100"
        default:
            return "invalid"
    }
}

func (register *ParsedSingleRegister) UsesARegister() bool {
    return register.Register == ARegister
}

type ParsedUnary struct {
    ParsedExpression
    Operation Operation
    Value ParsedExpression
}

func (unary *ParsedUnary) UsesARegister() bool {
    return unary.Value.UsesARegister()
}

type ParsedBinary struct {
    ParsedExpression
    Operation Operation
    Left ParsedExpression
    Right ParsedExpression
}

func isConstant(expression ParsedExpression) bool {
    _, ok := expression.(*ParsedConstant)
    return ok
}

func (binary *ParsedBinary) ToComputeBinaryString() string {
    switch binary.Operation {
        case OperationAdd:
            /* d+1, a+1, d+a */

            left, ok := binary.Left.(*ParsedSingleRegister)
            if !ok {
                return "invalid binary"
            }

            if isConstant(binary.Right) {
                right, ok := binary.Right.(*ParsedConstant)
                if !ok {
                    return "invalid binary"
                }

                if right.Value != 1 {
                    return "invalid binary"
                }

                switch left.Register {
                    /* d+1 */
                    case DRegister:
                        return "011111"

                    /* a+1 */
                    case ARegister, MRegister:
                        return "110111"
                    default: return "invalid binary"
                }

            } else {
                right, ok := binary.Right.(*ParsedSingleRegister)
                if !ok {
                    return "invalid binary"
                }

                if right.Register == DRegister {
                    return "invalid binary"
                }

                return "000010"
            }

        case OperationSubtract:
            /* d-1, a-1, d-a, a-d */

            return "unimplemented subtract"
        case OperationBinaryAnd:
            /* d&a */
            return "unimplemented and"
        case OperationBinaryOr:
            /* d|a */
            return "unimplemented or"
        default:
            return "invalid operation"
    }
}

func (binary *ParsedBinary) UsesARegister() bool {
    return binary.Left.UsesARegister() || binary.Right.UsesARegister()
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

    return &ParsedUnary{Operation: OperationNegate, Value: &ParsedSingleRegister{Register: register}}, nil
}

func parseNot(expression string) (ParsedExpression, error) {
    if len(expression) == 0 {
        return nil, fmt.Errorf("no expression given after '!'")
    }

    register, err := parseRegister(rune(expression[0]))

    if err != nil {
        return nil, err
    }

    return &ParsedUnary{Operation: OperationNot, Value: &ParsedSingleRegister{Register: register}}, nil
}

func maybeParseOperation(register Register, expression string) (ParsedExpression, error) {
    if len(expression) == 0 {
        return &ParsedSingleRegister{Register: register}, nil
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
            register, err = parseRegister(rune(right))
            if err != nil {
                return nil, err
            }
            rightValue = &ParsedSingleRegister{Register: register}
    }

    if rightValue == nil {
        return nil, fmt.Errorf("expected a value to follow the operation")
    }

    return &ParsedBinary{
        Operation: operation,
        Left: &ParsedSingleRegister{Register: register},
        Right: rightValue},
        nil
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

type ParsedJump struct {
    ParsedExpression
}

func (jump *ParsedJump) ToBinaryString() string {
    return "jump unimplemented"
}

func parseJump(code RawCode) (ParsedJump, error) {
    /* jump := value ; jump_op
     * jump_op := null | JGT | JEQ | JLT | JNE | JLE JMP
     */
    return ParsedJump{}, fmt.Errorf("unimplemented")
}

type ParsedMemoryReference struct {
    ParsedCode
    Constant int32
}

func (memory *ParsedMemoryReference) ToBinaryString() string {
    binary := strconv.FormatInt(int64(memory.Constant), 2)
    zeroPad := 16 - len(binary)
    if zeroPad < 1 {
        return fmt.Sprintf("invalid memory size %v", memory.Constant)
    }

    var builder strings.Builder
    for i := 0; i < zeroPad; i++ {
        builder.WriteRune('0')
    }
    builder.WriteString(binary)
    return builder.String()
}

func isNumber(value string) bool {
    _, err := strconv.Atoi(value)
    return err == nil
}

func parseMemoryConstant(value string) (int32, error) {
    out, err := strconv.Atoi(value)
    return int32(out), err
}

func parseMemoryReference(code RawCode) (ParsedMemoryReference, error) {
    line := code.Text

    if len(line) == 0 {
        return ParsedMemoryReference{}, fmt.Errorf("not a memory reference")
    }

    if line[0] != '@' {
        return ParsedMemoryReference{}, fmt.Errorf("not a memory reference")
    }

    value := line[1:]
    if isNumber(value) {
        parsed, err := parseMemoryConstant(value)
        if err != nil {
            return ParsedMemoryReference{}, err
        }
        return ParsedMemoryReference{Constant: parsed}, nil
    }

    return ParsedMemoryReference{}, fmt.Errorf("unimplemented memory reference on line %v", code.SourceLine)
}

func parse(raw RawProgram) (ParsedProgram, error) {
    // variableAllocator := 16
    var parsed ParsedProgram

    for _, code := range raw.Code {
        /* line := assignment | label declaration | jump | variable/explicit A value
         * assignment := X=Y
         * label declaration := (FOO)
         * jump :=
         * variable/explicit A value := @2 | @foo
         */
        if strings.ContainsRune(code.Text, '=') {
            converted, err := parseAssignment(code)
            if err != nil {
                return parsed, err
            }
            parsed.Add(&converted)
        } else if strings.ContainsRune(code.Text, ';') {
            converted, err := parseJump(code)
            if err != nil {
                return parsed, err
            }
            parsed.Add(&converted)
        } else if strings.HasPrefix(code.Text, "@") {
            converted, err := parseMemoryReference(code)
            if err != nil {
                return parsed, err
            }
            parsed.Add(&converted)
        } else {
            return parsed, fmt.Errorf("Error line %v: unknown line '%v'", code.SourceLine, code.Text)
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

    parsed.DumpAsBinaryString()

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
