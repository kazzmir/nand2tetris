package main

import (
    "os"
    "fmt"
    "flag"
    "io/ioutil"
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
    if strings.Contains(line, "//") {
        index := strings.Index(line, "//")
        line = line[0:index]
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
    UsesMRegister() bool
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

    usesM := assignment.Expression.UsesMRegister()

    if usesM {
        out.WriteRune('1')
    } else {
        out.WriteRune('0')
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

    // fmt.Printf("d=%v m=%v a=%v\n", assignD, assignM, assignA)

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

func (program *ParsedProgram) FixupLabels(labels *LabelManager) error {
    variableAllocator := VariableAllocator{
        CurrentSlot: 16,
        Mapping: make(map[string]int32),
    }

    for _, code := range program.Code {
        memory, ok := code.(*ParsedMemoryReference)
        if ok {
            if memory.Constant == -1 {
                label, err := labels.Lookup(memory.LabelReference)

                /* If its not a defined label then it must have been a variable */
                if err != nil {
                    memory.Constant = variableAllocator.Get(memory.LabelReference)
                } else {
                    memory.Constant = label
                }
            }
        }
    }

    return nil
}

func (program *ParsedProgram) InstructionCount() int32 {
    return int32(len(program.Code))
}

func (program *ParsedProgram) ToBinaryString() string {
    var out strings.Builder
    for _, code := range program.Code {
        out.WriteString(fmt.Sprintf("%v\n", code.ToBinaryString()))
    }
    return out.String()
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

func (constant *ParsedConstant) UsesMRegister() bool {
    return false
}

func (constant *ParsedConstant) ToComputeBinaryString() string {
    switch constant.Value {
        case 0: return "101010"
        case 1: return "111111"
        case -1: return "111010"
        default: fmt.Sprintf("unknown constant %v", constant.Value)
    }

    return "fail"
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

func (register *ParsedSingleRegister) UsesMRegister() bool {
    return register.Register == MRegister
}

type ParsedUnary struct {
    ParsedExpression
    Operation Operation
    Value ParsedExpression
}

func (unary *ParsedUnary) ToComputeBinaryString() string {
    switch unary.Operation {
        case OperationNot:
            register, ok := unary.Value.(*ParsedSingleRegister)
            if !ok {
                return "invalid unary operation"
            }

            switch register.Register {
                case DRegister: return "0011001"
                case ARegister, MRegister: return "110001"
            }

            return "invalid unary operation"

        case OperationNegate:
            register, ok := unary.Value.(*ParsedSingleRegister)
            if !ok {
                return "invalid unary operation"
            }

            switch register.Register {
                case DRegister: return "001111"
                case ARegister, MRegister: return "110011"
            }

            return "invalid unary operation"

        default: fmt.Sprintf("unimplemented operation %v", unary.Operation)
    }

    return "invalid unary operation"
}

func (unary *ParsedUnary) UsesMRegister() bool {
    return unary.Value.UsesMRegister()
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

            left, ok := binary.Left.(*ParsedSingleRegister)
            if !ok {
                return "invalid left side of subtract"
            }

            rightNumber, ok := binary.Right.(*ParsedConstant)
            if ok {
                /* d-1, a-1 */

                if rightNumber.Value != 1 {
                    return "invalid subtract constant"
                }

                switch left.Register {
                    case DRegister: return "001110"
                    case ARegister, MRegister: return "110010"
                }

            } else {
                /* a-d, d-a */
                rightRegister, ok := binary.Right.(*ParsedSingleRegister)
                if !ok {
                    return "invalid right side of subtract"
                }

                if left.Register == DRegister && (rightRegister.Register == ARegister || rightRegister.Register == MRegister) {
                    return "010011"
                }
                if (left.Register == ARegister || left.Register == MRegister) && rightRegister.Register == DRegister {
                    return "000111"
                }

                return fmt.Sprintf("[unknown subtraction between %v and %v]", left.Register, rightRegister.Register)
            }

            return "unimplemented subtract"
        case OperationBinaryAnd:
            /* d&a */
            return "000000"
        case OperationBinaryOr:
            /* d|a */
            /* FIXME: check the registers */
            return "010101"
        default:
            return "invalid operation"
    }
}

func (binary *ParsedBinary) UsesMRegister() bool {
    return binary.Left.UsesMRegister() || binary.Right.UsesMRegister()
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
        case '0': rightValue = &ParsedConstant{Value: 0}
        case '1': rightValue = &ParsedConstant{Value: 1}
        default:
            rightRegister, err := parseRegister(rune(right))
            if err != nil {
                return nil, err
            }
            rightValue = &ParsedSingleRegister{Register: rightRegister}
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

type Jump int
const (
    JGT Jump = iota
    JEQ
    JLT
    JGE
    JNE
    JLE
    JMP
    NoJump
    InvalidJump
)

func parseJumpType(jump string) (Jump, error) {
    switch jump {
        case "null": return NoJump, nil
        case "JGT": return JGT, nil
        case "JEQ": return JEQ, nil
        case "JGE": return JGE, nil
        case "JLT": return JLT, nil
        case "JNE": return JNE, nil
        case "JLE": return JLE, nil
        case "JMP": return JMP, nil
        default: return InvalidJump, fmt.Errorf("unknown jump type '%v'", jump)
    }
}

type ParsedJump struct {
    ParsedExpression

    Expression ParsedExpression
    Jump Jump
}

func (jump *ParsedJump) ToBinaryString() string {
    var out strings.Builder

    out.WriteString("111")

    usesM := jump.Expression.UsesMRegister()

    if usesM {
        out.WriteRune('1')
    } else {
        out.WriteRune('0')
    }

    out.WriteString(jump.Expression.ToComputeBinaryString())

    /* no destination */
    out.WriteString("000")

    switch jump.Jump {
        case NoJump: out.WriteString("000")
        case JGT: out.WriteString("001")
        case JEQ: out.WriteString("010")
        case JGE: out.WriteString("011")
        case JLT: out.WriteString("100")
        case JNE: out.WriteString("101")
        case JLE: out.WriteString("110")
        case JMP: out.WriteString("111")
        default:
            out.WriteString("invalid jump")
    }

    return out.String()
}

func parseJump(code RawCode) (ParsedJump, error) {
    /* jump := value ; jump_op
     * jump_op := null | JGT | JEQ | JLT | JNE | JLE | JMP
     */
    parts := strings.Split(code.Text, ";")
    if len(parts) != 2 {
        return ParsedJump{}, fmt.Errorf("expected a jump to be separated by a ;")
    }
    expression := parts[0]
    jump := parts[1]

    parsedExpression, err := parseExpression(expression)
    if err != nil {
        return ParsedJump{}, err
    }

    parsedJump, err := parseJumpType(jump)
    if err != nil {
        return ParsedJump{}, err
    }

    return ParsedJump{
        Expression: parsedExpression,
        Jump: parsedJump,
    }, nil
}

type ParsedMemoryReference struct {
    ParsedCode
    Constant int32
    LabelReference string // reference to a label
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

func isSpecialMemory(value string) bool {
    switch value {
        case "SCREEN", "KBD", "THIS", "THAT", "SP", "LCL", "ARG": return true
        default: return false
    }
}

func isRamSlot(ram string) bool {
    if len(ram) == 0 {
        return false
    }

    if ram[0] != 'R' {
        return false
    }

    if !isNumber(ram[1:]) {
        return false
    }

    return true
}

func parseRamSlot(ram string) int32 {
    value, err := strconv.Atoi(ram[1:])
    if err != nil {
        return -1
    }
    return int32(value)
}

type VariableAllocator struct {
    CurrentSlot int32
    Mapping map[string]int32
}

func (allocator *VariableAllocator) Get(variable string) int32 {
    slot, ok := allocator.Mapping[variable]
    if ok {
        return slot
    }

    allocator.Mapping[variable] = allocator.CurrentSlot
    allocator.CurrentSlot += 1
    return allocator.Mapping[variable]
}

func isAllCaps(value string) bool {
    return value == strings.ToUpper(value)
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

    if isRamSlot(value) {
        return ParsedMemoryReference{Constant: parseRamSlot(value)}, nil
    }

    /* could be a variable reference, a label reference, or a special reference like
     * SCREEN, KBD, etc
     */

    if isSpecialMemory(value) {
        switch value {
            case "SP": return ParsedMemoryReference{Constant: 0}, nil
            case "LCL": return ParsedMemoryReference{Constant: 1}, nil
            case "ARG": return ParsedMemoryReference{Constant: 2}, nil
            case "THIS": return ParsedMemoryReference{Constant: 3}, nil
            case "THAT": return ParsedMemoryReference{Constant: 4}, nil
            case "SCREEN": return ParsedMemoryReference{Constant: 0x4000}, nil
            case "KBD": return ParsedMemoryReference{Constant: 0x6000}, nil
        }

        return ParsedMemoryReference{}, fmt.Errorf("unimplemented special memory reference on line %v '%v'", code.SourceLine, code.Text)
    }

    /* otherwise its a label or variable */
    return ParsedMemoryReference{LabelReference: value, Constant: -1}, nil
}

type LabelManager struct {
    Labels map[string]int32
}

func (manager *LabelManager) Lookup(label string) (int32, error) {
    value, ok := manager.Labels[label]
    if !ok {
        return -1, fmt.Errorf("unknown label '%v'", label)
    }
    return value, nil
}

func (manager *LabelManager) SetLabel(label string, count int32) error {
    _, ok := manager.Labels[label]
    if ok {
        return fmt.Errorf("Existing label named '%v'", label)
    }

    manager.Labels[label] = count
    return nil
}

func parseLabel(raw RawCode) (string, error) {
    label := raw.Text
    if strings.HasPrefix(label, "(") && strings.HasSuffix(label, ")") {
        return label[1:len(label)-1], nil
    } else {
        return "", fmt.Errorf("unknown label syntax '%v'", label)
    }
}

func parse(raw RawProgram) (ParsedProgram, error) {
    

    labelManager := LabelManager {
        Labels: make(map[string]int32),
    }

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
        } else if strings.HasPrefix(code.Text, "(") {
            label, err := parseLabel(code)
            if err != nil {
                return parsed, err
            }
            labelManager.SetLabel(label, parsed.InstructionCount())
        } else {
            return parsed, fmt.Errorf("Error line %v: unknown line '%v'", code.SourceLine, code.Text)
        }
    }

    parsed.FixupLabels(&labelManager)

    return parsed, nil
}

func replaceExtension(path string, extension string) string {
    parts := strings.Split(path, ".")
    return fmt.Sprintf("%v.%v", parts[0], extension)
}

func writeToHack(data string, asmPath string) error {
    hackPath := replaceExtension(asmPath, "hack")
    err := ioutil.WriteFile(hackPath, []byte(data), 0644)
    if err == nil {
        fmt.Printf("Assembled to %v\n", hackPath)
    }
    return err
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

    fmt.Printf("%v", parsed.ToBinaryString())

    err = writeToHack(parsed.ToBinaryString(), path)
    return err
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
