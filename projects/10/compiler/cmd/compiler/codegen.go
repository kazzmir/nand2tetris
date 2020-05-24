package main

import (
    "fmt"
    "io"
)

/* holds the slot in whatever segment this variable lives in
 * and the type of the variable
 */
type VariableMapping struct {
    Slot int
    Type string
}

type FunctionGenerator struct {
    CodeGenerator *CodeGenerator
    /* Map a variable to a local slot */
    LocalVariables map[string]VariableMapping
    LocalCount int
    Parameters map[string]VariableMapping
    ParameterCount int
    gensym int
}

func (function *FunctionGenerator) Gensym(name string) string {
    out := fmt.Sprintf("%v_%v", name, function.gensym)
    function.gensym += 1
    return out
}

func (function *FunctionGenerator) RegisterParameter(name string, type_ string){
    function.Parameters[name] = VariableMapping{
        Slot: function.ParameterCount,
        Type: type_,
    }
    function.ParameterCount += 1
}

func (function *FunctionGenerator) IsParameter(name string) bool {
    _, ok := function.Parameters[name]
    return ok
}

func (function *FunctionGenerator) GetParameter(name string) int {
    mapping, ok := function.Parameters[name]
    if !ok {
        return -1
    }

    return mapping.Slot
}

func (function *FunctionGenerator) RegisterVariable(name string, type_ string){
    function.LocalVariables[name] = VariableMapping{
        Slot: function.LocalCount,
        Type: type_,
    }
    function.LocalCount += 1
}

func (function *FunctionGenerator) IsLocal(name string) bool {
    _, ok := function.LocalVariables[name]
    return ok
}

func (function *FunctionGenerator) GetLocal(name string) int {
    value, ok := function.LocalVariables[name]
    if !ok {
        return -1
    }

    return value.Slot
}

func (function *FunctionGenerator) VisitBoolean(ast *ASTBoolean) (interface{}, error) {
    if ast.Value {
        function.CodeGenerator.Emit <- "push constant 0"
        function.CodeGenerator.Emit <- "not"
    } else {
        function.CodeGenerator.Emit <- "push constant 0"
    }
    return nil, nil
}

func (function *FunctionGenerator) VisitCall(ast *ASTCall) (interface{}, error) {
    return nil, fmt.Errorf("function generator: call unimplemented")
}

func (function *FunctionGenerator) VisitClass(ast *ASTClass) (interface{}, error) {
    return nil, fmt.Errorf("function generator: should not visit class")
}

func (function *FunctionGenerator) VisitConstant(ast *ASTConstant) (interface{}, error) {
    function.CodeGenerator.Emit <- fmt.Sprintf("push constant %v", ast.Number)
    return nil, nil
}

func (function *FunctionGenerator) VisitConstructor(ast *ASTConstructor) (interface{}, error) {
    return nil, fmt.Errorf("function generator: should not visit constructor")
}

func (function *FunctionGenerator) VisitDo(ast *ASTDo) (interface{}, error) {
    _, err := ast.Expression.Visit(function)
    if err != nil {
        return nil, err
    }

    function.CodeGenerator.Emit <- "pop temp 0"
    return nil, nil
}

func (function *FunctionGenerator) VisitField(ast *ASTField) (interface{}, error) {
    return nil, fmt.Errorf("function generator: should not visit field")
}

func (function *FunctionGenerator) VisitBlock(ast *ASTBlock) (interface{}, error) {
    for _, statement := range ast.Statements {
        _, err := statement.Visit(function)
        if err != nil {
            return nil, err
        }
    }

    return nil, nil
}

func (function *FunctionGenerator) VisitIdentifier(ast *ASTIdentifier) (interface{}, error) {
    return nil, fmt.Errorf("function generator: unimplemented identifier")
}

func (function *FunctionGenerator) VisitIf(ast *ASTIf) (interface{}, error) {
    /*
    Condition ASTExpression
    Then *ASTBlock
    Else *ASTBlock
    */

    labelElse := function.Gensym("if_else")
    labelDone := function.Gensym("if_done")

    /* check if the condition is true or false */
    _, err := ast.Condition.Visit(function)
    if err != nil {
        return nil, err
    }

    /* negate the condition */
    function.CodeGenerator.Emit <- "not"
    if ast.Else != nil {
        /* condition failed, so jump to else */
        function.CodeGenerator.Emit <- fmt.Sprintf("if-goto %v", labelElse)
    } else {
        /* there is no else block, so skip directly to done */
        function.CodeGenerator.Emit <- fmt.Sprintf("if-goto %v", labelDone)
    }

    /* otherwise the condition is true, so fall through here */

    _, err = ast.Then.Visit(function)
    if err != nil {
        return nil, err
    }

    if ast.Else != nil {
        /* from the then block, skip over the else and go directly to done */
        function.CodeGenerator.Emit <- fmt.Sprintf("goto %v", labelDone)
        function.CodeGenerator.Emit <- fmt.Sprintf("label %v", labelElse)

        _, err = ast.Else.Visit(function)
        if err != nil {
            return nil, err
        }
    }

    function.CodeGenerator.Emit <- fmt.Sprintf("label %v", labelDone)

    return nil, nil
}

func (function *FunctionGenerator) VisitIndexExpression(ast *ASTIndexExpression) (interface{}, error) {
    _, err := ast.Left.Visit(function)
    if err != nil {
        return nil, err
    }

    _, err = ast.Index.Visit(function)
    if err != nil {
        return nil, err
    }

    function.CodeGenerator.Emit <- "add"
    function.CodeGenerator.Emit <- "pop pointer 1"
    function.CodeGenerator.Emit <- "push that 0"

    return nil, nil
}

func (function *FunctionGenerator) VisitNegation(ast *ASTNegation) (interface{}, error) {
    _, err := ast.Expression.Visit(function)
    if err != nil {
        return nil, err
    }
    function.CodeGenerator.Emit <- "neg"

    return nil, nil
}

func (function *FunctionGenerator) VisitNot(ast *ASTNot) (interface{}, error) {
    return nil, fmt.Errorf("function generator: not unimplemented")
}

func (function *FunctionGenerator) VisitNull(ast *ASTNull) (interface{}, error) {
    function.CodeGenerator.Emit <- "push constant 0"
    return nil, nil
}

func (function *FunctionGenerator) VisitOperator(ast *ASTOperator) (interface{}, error) {
    _, err := ast.Left.Visit(function)
    if err != nil {
        return nil, err
    }

    _, err = ast.Right.Visit(function)
    if err != nil {
        return nil, err
    }

    switch ast.Operator {
    case TokenPlus:
        function.CodeGenerator.Emit <- "add"
        return nil, nil
    case TokenNegation:
        function.CodeGenerator.Emit <- "sub"
        return nil, nil
    case TokenDivision:
        function.CodeGenerator.Emit <- "call Math.divide 2"
        return nil, nil
    case TokenMultiply:
        function.CodeGenerator.Emit <- "call Math.multiply 2"
        return nil, nil
    case TokenLessThan:
        function.CodeGenerator.Emit <- "lt"
        return nil, nil
    case TokenGreaterThan:
        function.CodeGenerator.Emit <- "gt"
        return nil, nil
    case TokenOr:
        function.CodeGenerator.Emit <- "or"
        return nil, nil
    }

    return nil, fmt.Errorf("function generator: unknown operator %v", ast.Operator.Name())
}

func (function *FunctionGenerator) VisitReference(ast *ASTReference) (interface{}, error) {
    emitter := function.CodeGenerator.Emit

    if function.IsLocal(ast.Name) {
        emitter <- fmt.Sprintf("push local %v", function.GetLocal(ast.Name))
        return nil, nil
    }

    if function.IsParameter(ast.Name) {
        emitter <- fmt.Sprintf("push argument %v", function.GetParameter(ast.Name))
        return nil, nil
    }

    if function.CodeGenerator.IsField(ast.Name) {
        emitter <- fmt.Sprintf("push this %v", function.CodeGenerator.GetField(ast.Name))
        return nil, nil
    }

    if function.CodeGenerator.IsStatic(ast.Name) {
        emitter <- fmt.Sprintf("push static %v", function.CodeGenerator.GetStatic(ast.Name))
        return nil, nil
    }

    return nil, fmt.Errorf("function generator: unknown reference %v", ast.Name)
}

func (function *FunctionGenerator) VisitReturn(ast *ASTReturn) (interface{}, error) {
    if ast.Expression != nil {
        _, err := ast.Expression.Visit(function)
        if err != nil {
            return nil, err
        }
    }

    function.CodeGenerator.Emit <- "return"
    return nil, nil
}

func (function *FunctionGenerator) VisitStatic(ast *ASTStatic) (interface{}, error) {
    return nil, fmt.Errorf("function generator: static should not be visited")
}

func (function *FunctionGenerator) VisitThis(ast *ASTThis) (interface{}, error) {
    return nil, fmt.Errorf("function generator: this unimplemented")
}

func (function *FunctionGenerator) VisitVar(ast *ASTVar) (interface{}, error) {

    type_, ok := ast.Type.Type.(*ASTIdentifier)
    if !ok {
        return nil, fmt.Errorf("unknown type node %v", ast.Type.ToSExpression())
    }

    for _, name := range ast.Names {
        function.RegisterVariable(name, type_.Name)
    }

    return nil, nil
}

func (function *FunctionGenerator) VisitWhile(ast *ASTWhile) (interface{}, error) {
    labelStart := function.Gensym("while_start")
    labelEnd := function.Gensym("while_end")

    emitter := function.CodeGenerator.Emit
    emitter <- fmt.Sprintf("label %v", labelStart)

    _, err := ast.Condition.Visit(function)
    if err != nil {
        return nil, err
    }

    emitter <- "not"
    emitter <- fmt.Sprintf("if-goto %v", labelEnd)

    _, err = ast.Body.Visit(function)
    if err != nil {
        return nil, err
    }

    emitter <- fmt.Sprintf("goto %v", labelStart)
    emitter <- fmt.Sprintf("label %v", labelEnd)

    return nil, nil
}

func (function *FunctionGenerator) VisitString(ast *ASTString) (interface{}, error) {
    emitter := function.CodeGenerator.Emit
    emitter <- fmt.Sprintf("push constant %v", len(ast.Value))
    emitter <- "call String.new 1"
    for _, char_ := range ast.Value {
        emitter <- fmt.Sprintf("push constant %v", int(char_))
        emitter <- "call String.appendChar 2"
    }

    return nil, nil
}

func (function *FunctionGenerator) GetType(ast *ASTReference) (string, error) {
    if function.IsLocal(ast.Name) {
        mapping := function.LocalVariables[ast.Name]
        return mapping.Type, nil
    }

    return "", fmt.Errorf("unknown reference %v", ast.Name)
}

func (function *FunctionGenerator) VisitMethodCall(ast *ASTMethodCall) (interface{}, error) {
    var name string

    passThis := 0
    reference, ok := ast.Left.(*ASTReference)
    /* try to determine if this is a call to a static function or a method call on
     * a live variable. for a variable call like foo.xyz(), we implicitly pass this as
     * argument 1, so the number of arguments passed in is 1 + actual arguments, e.g:
     * foo.xyz(4) -> has 2 arguments: this and the number 4
     */
    if ok {
        _, err := reference.Visit(function)
        if err != nil {
            /* hopefully its some existing type.. */
            name = fmt.Sprintf("%v.%v", reference.Name, ast.Call.Name)
        } else {
            type_, err := function.GetType(reference)
            if err != nil {
                return nil, fmt.Errorf("unknown type for reference %v", reference.Name)
            }
            name = fmt.Sprintf("%v.%v", type_, ast.Call.Name)
            passThis = 1
        }
    } else {
        _, err := ast.Left.Visit(function)
        if err != nil {
            return nil, err
        }
    }

    for _, argument := range ast.Call.Arguments {
        _, err := argument.Visit(function)
        if err != nil {
            return nil, err
        }
    }

    function.CodeGenerator.Emit <- fmt.Sprintf("call %v %v", name, len(ast.Call.Arguments) + passThis)
    return nil, nil
}

func (function *FunctionGenerator) VisitLet(ast *ASTLet) (interface{}, error) {
    _, err := ast.Expression.Visit(function)
    if err != nil {
        return nil, err
    }

    if ast.ArrayIndex != nil {
        /* compute the array offset */
        reference := &ASTReference{Name: ast.Name}
        _, err := reference.Visit(function)
        if err != nil {
            return nil, err
        }

        _, err = ast.ArrayIndex.Visit(function)
        if err != nil {
            return nil, err
        }

        function.CodeGenerator.Emit <- "add"
        /* set up 'that' register */
        function.CodeGenerator.Emit <- "pop pointer 1"
        /* assign 'that' */
        function.CodeGenerator.Emit <- "pop that 0"

        return nil, nil
    } else if function.IsLocal(ast.Name) {
        function.CodeGenerator.Emit <- fmt.Sprintf("pop local %v", function.GetLocal(ast.Name))
        return nil, nil
    } else if function.CodeGenerator.IsStatic(ast.Name) {
        function.CodeGenerator.Emit <- fmt.Sprintf("pop static %v", function.CodeGenerator.GetStatic(ast.Name))
        return nil, nil
    }

    return nil, fmt.Errorf("let: unknown name %v", ast.Name)
}

func (function *FunctionGenerator) processFunctionOrMethod(ast ASTNode) (interface{}, error) {
    localEmitter := make(chan string, 10)
    savedEmit := function.CodeGenerator.Emit
    function.CodeGenerator.Emit = localEmitter

    var parameters []*ASTParameter
    var body *ASTBlock
    var name string
    methodAST, isMethod := ast.(*ASTMethod)
    functionAST, isFunction := ast.(*ASTFunction)

    if isMethod {
        parameters = methodAST.Parameters
        body = methodAST.Body
        name = methodAST.Name
    } else if isFunction {
        parameters = functionAST.Parameters
        body = functionAST.Body
        name = functionAST.Name
    } else {
        return nil, fmt.Errorf("not a method or function")
    }

    for _, parameter := range parameters {
        type_, ok := parameter.Type.Type.(*ASTIdentifier)
        if !ok {
            return nil, fmt.Errorf("unknown type %v", parameter.Type.ToSExpression())
        }
        function.RegisterParameter(parameter.Name, type_.Name)
    }

    /* Have to process the body first to find out how many locals there are */
    var functionError error
    go func(){
        defer close(localEmitter)
        _, functionError = body.Visit(function)
    }()

    var code []string

    for line := range localEmitter {
        code = append(code, line)
    }

    if functionError != nil {
        return nil, functionError
    }

    function.CodeGenerator.Emit = savedEmit
    function.CodeGenerator.Emit <- fmt.Sprintf("function %v.%v %v", function.CodeGenerator.ClassName, name, function.LocalCount)

    if isMethod {
        /* set up the 'this' register */
        function.CodeGenerator.Emit <- "push argument 0"
        function.CodeGenerator.Emit <- "pop pointer 0"
    }

    for _, line := range code {
        function.CodeGenerator.Emit <- line
    }

    return nil, nil
}

func (function *FunctionGenerator) VisitFunction(ast *ASTFunction) (interface{}, error) {
    return function.processFunctionOrMethod(ast)
}

func (function *FunctionGenerator) VisitMethod(ast *ASTMethod) (interface{}, error) {
    return function.processFunctionOrMethod(ast)
}

type CodeGenerator struct {
    Emit chan(string)
    ClassName string
    Classes map[string]bool

    Fields map[string]int
    FieldCount int

    Statics map[string]int
    StaticCount int
}

func (generator *CodeGenerator) RegisterClass(name string){
    generator.Classes[name] = true
}

func (generator *CodeGenerator) IsClass(name string) bool {
    _, ok := generator.Classes[name]
    return ok
}

func (generator *CodeGenerator) RegisterStatic(name string){
    generator.Statics[name] = generator.StaticCount
    generator.StaticCount += 1
}

func (generator *CodeGenerator) IsStatic(name string) bool {
    _, ok := generator.Statics[name]
    return ok
}

func (generator *CodeGenerator) GetStatic(name string) int {
    value, ok := generator.Statics[name]
    if !ok {
        return -1
    }
    return value
}

func (generator *CodeGenerator) RegisterField(name string){
    generator.Fields[name] = generator.FieldCount
    generator.FieldCount += 1
}

func (generator *CodeGenerator) IsField(name string) bool {
    _, ok := generator.Fields[name]
    return ok
}

func (generator *CodeGenerator) GetField(name string) int {
    index, ok := generator.Fields[name]
    if !ok {
        return -1
    }

    return index
}

func (generator *CodeGenerator) VisitClass(ast *ASTClass) (interface{}, error) {
    generator.RegisterClass(ast.Name)
    generator.ClassName = ast.Name

    for _, body := range ast.Body {
        field, ok := body.(*ASTField)
        if ok {
            _, err := field.Visit(generator)
            if err != nil {
                return nil, err
            }
        }

        static, ok := body.(*ASTStatic)
        if ok {
            _, err := static.Visit(generator)
            if err != nil {
                return nil, err
            }
        }
    }

    for _, body := range ast.Body {
        function, ok := body.(*ASTFunction)
        if ok {
            _, err := function.Visit(generator)
            if err != nil {
                return nil, err
            }

            continue
        }

        method, ok := body.(*ASTMethod)
        if ok {
            _, err := method.Visit(generator)
            if err != nil {
                return nil, err
            }

            continue
        }

        _, ok = body.(*ASTField)
        if ok {
            continue
        }

        _, ok = body.(*ASTStatic)
        if ok {
            continue
        }


        return nil, fmt.Errorf("code generator: unknown class body %v", body.ToSExpression())
    }

    return nil, nil
}

func (generator *CodeGenerator) VisitIdentifier(*ASTIdentifier) (interface{}, error) {
    return nil, fmt.Errorf("unimplemented identifier")
}

func (generator *CodeGenerator) VisitBoolean(*ASTBoolean) (interface{}, error) {
    return nil, fmt.Errorf("unimplemented boolean")
}

func (generator *CodeGenerator) VisitString(ast *ASTString) (interface{}, error) {
    return nil, fmt.Errorf("code generator should not visit string")
}

func (generator *CodeGenerator) VisitNull(*ASTNull) (interface{}, error) {
    return nil, fmt.Errorf("unimplemented null")
}

func (generator *CodeGenerator) VisitCall(*ASTCall) (interface{}, error) {
    return nil, fmt.Errorf("unimplemented call")
}

func (generator *CodeGenerator) VisitIndexExpression(*ASTIndexExpression) (interface{}, error) {
    return nil, fmt.Errorf("unimplemented index expression")
}

func (generator *CodeGenerator) VisitVar(*ASTVar) (interface{}, error) {
    return nil, fmt.Errorf("code generator should not visit var")
}

func (generator *CodeGenerator) VisitMethodCall(ast *ASTMethodCall) (interface{}, error) {
    return nil, fmt.Errorf("code generator should not visit method call")
}

func (generator *CodeGenerator) VisitNot(*ASTNot) (interface{}, error) {
    return nil, fmt.Errorf("unimplemented not")
}

func (generator *CodeGenerator) VisitNegation(*ASTNegation) (interface{}, error) {
    return nil, fmt.Errorf("unimplemented negation")
}

func (generator *CodeGenerator) VisitOperator(*ASTOperator) (interface{}, error) {
    return nil, fmt.Errorf("unimplemented operator")
}

func (generator *CodeGenerator) VisitThis(*ASTThis) (interface{}, error) {
    return nil, fmt.Errorf("unimplemented this")
}

func (generator *CodeGenerator) VisitConstant(*ASTConstant) (interface{}, error) {
    return nil, fmt.Errorf("unimplemented constant")
}

func (generator *CodeGenerator) VisitReference(*ASTReference) (interface{}, error) {
    return nil, fmt.Errorf("code generator: reference should not be visited")
}

func (generator *CodeGenerator) VisitWhile(*ASTWhile) (interface{}, error) {
    return nil, fmt.Errorf("unimplemented while")
}

func (generator *CodeGenerator) VisitConstructor(*ASTConstructor) (interface{}, error) {
    return nil, fmt.Errorf("unimplemented constructor")
}

func (generator *CodeGenerator) VisitIf(*ASTIf) (interface{}, error) {
    return nil, fmt.Errorf("unimplemented if")
}

func (generator *CodeGenerator) VisitMethod(ast *ASTMethod) (interface{}, error) {
    function := FunctionGenerator{
        CodeGenerator: generator,
        LocalVariables: make(map[string]VariableMapping),
        Parameters: make(map[string]VariableMapping),
        ParameterCount: 1,
    }

    return ast.Visit(&function)
}

func (generator *CodeGenerator) VisitBlock(ast *ASTBlock) (interface{}, error) {
    return nil, fmt.Errorf("code generator should not visit block")
}

func (generator *CodeGenerator) VisitFunction(ast *ASTFunction) (interface{}, error) {
    function := FunctionGenerator{
        CodeGenerator: generator,
        LocalVariables: make(map[string]VariableMapping),
        Parameters: make(map[string]VariableMapping),
    }

    return ast.Visit(&function)
}

func (generator *CodeGenerator) VisitLet(ast *ASTLet) (interface{}, error) {
    return nil, fmt.Errorf("code generator should not visit let")
}

func (generator *CodeGenerator) VisitDo(*ASTDo) (interface{}, error) {
    return nil, fmt.Errorf("unimplemented do")
}

func (generator *CodeGenerator) VisitReturn(*ASTReturn) (interface{}, error) {
    return nil, fmt.Errorf("unimplemented return")
}

func (generator *CodeGenerator) VisitStatic(ast *ASTStatic) (interface{}, error) {
    for _, name := range ast.Names {
        generator.RegisterStatic(name.Name)
    }

    return nil, nil
}

func (generator *CodeGenerator) VisitField(ast *ASTField) (interface{}, error) {

    for _, name := range ast.Names {
        generator.RegisterField(name)
    }

    return nil, nil
}

func GenerateCode(ast ASTNode, writer io.Writer) error {
    vmChannel := make(chan string, 10)
    classes := make(map[string]bool)
    classes["Keyboard"] = true
    classes["Array"] = true
    classes["Output"] = true
    generator := CodeGenerator{
        Emit: vmChannel,
        Fields: make(map[string]int),
        Statics: make(map[string]int),
        Classes: classes,
    }
    var codegenError error
    go func(){
        defer close(vmChannel)
        _, codegenError = ast.Visit(&generator)
    }()

    /* drain the channel so it doesn't get lost forever */
    drain := func () {
        for range vmChannel {
        }
    }

    for vm := range vmChannel {
        _, err := io.WriteString(writer, vm)
        if err != nil {
            go drain()
            return err
        }
        _, err = writer.Write([]byte{'\n'})
        if err != nil {
            go drain()
            return err
        }
    }

    return codegenError
}
