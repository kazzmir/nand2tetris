package main

import (
    "fmt"
    "io"
)

type FunctionGenerator struct {
    CodeGenerator *CodeGenerator
    /* Map a variable to a local slot */
    LocalVariables map[string]int
    LocalCount int
}

func (function *FunctionGenerator) RegisterVariable(name string){
    function.LocalVariables[name] = function.LocalCount
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

    return value
}

func (function *FunctionGenerator) VisitBoolean(ast *ASTBoolean) (interface{}, error) {
    /* FIXME */
    function.CodeGenerator.Emit <- "push false"
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
    return nil, fmt.Errorf("function generator: do unimplemented")
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
    return nil, fmt.Errorf("function generator: if unimplemented")
}

func (function *FunctionGenerator) VisitIndexExpression(ast *ASTIndexExpression) (interface{}, error) {
    return nil, fmt.Errorf("function generator: index expression unimplemented")
}

func (function *FunctionGenerator) VisitNegation(ast *ASTNegation) (interface{}, error) {
    return nil, fmt.Errorf("function generator: negation unimplemented")
}

func (function *FunctionGenerator) VisitNot(ast *ASTNot) (interface{}, error) {
    return nil, fmt.Errorf("function generator: not unimplemented")
}

func (function *FunctionGenerator) VisitNull(ast *ASTNull) (interface{}, error) {
    return nil, fmt.Errorf("function generator: null unimplemented")
}

func (function *FunctionGenerator) VisitOperator(ast *ASTOperator) (interface{}, error) {
    return nil, fmt.Errorf("function generator: operator unimplemented")
}

func (function *FunctionGenerator) VisitReference(ast *ASTReference) (interface{}, error) {
    emitter := function.CodeGenerator.Emit

    if function.IsLocal(ast.Name) {
        emitter <- fmt.Sprintf("push local %v", function.GetLocal(ast.Name))
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

    for _, name := range ast.Names {
        function.RegisterVariable(name)
    }

    return nil, nil
}

func (function *FunctionGenerator) VisitWhile(ast *ASTWhile) (interface{}, error) {
    return nil, fmt.Errorf("function generator: while unimplemented")
}

func (function *FunctionGenerator) VisitString(ast *ASTString) (interface{}, error) {
    emitter := function.CodeGenerator.Emit
    emitter <- "call String.new 1"
    for _, char_ := range ast.Value {
        emitter <- fmt.Sprintf("push constant %v", int(char_))
        emitter <- "call String.appendChar 2"
    }

    return nil, nil
}

func (function *FunctionGenerator) VisitMethodCall(ast *ASTMethodCall) (interface{}, error) {
    _, err := ast.Left.Visit(function)
    if err != nil {
        return nil, err
    }

    for _, argument := range ast.Call.Arguments {
        _, err := argument.Visit(function)
        if err != nil {
            return nil, err
        }
    }

    function.CodeGenerator.Emit <- fmt.Sprintf("call %v %v", ast.Call.Name, len(ast.Call.Arguments))
    return nil, nil
}

func (function *FunctionGenerator) VisitLet(ast *ASTLet) (interface{}, error) {
    _, err := ast.Expression.Visit(function)
    if err != nil {
        return nil, err
    }

    if ast.ArrayIndex != nil {
        return nil, fmt.Errorf("let: array index unimplemented")
    }

    if function.IsLocal(ast.Name) {
        function.CodeGenerator.Emit <- fmt.Sprintf("pop local %v", function.GetLocal(ast.Name))
        return nil, nil
    }

    return nil, fmt.Errorf("let: unknown name %v", ast.Name)
}

func (function *FunctionGenerator) VisitFunction(ast *ASTFunction) (interface{}, error) {
    localEmitter := make(chan string, 10)
    savedEmit := function.CodeGenerator.Emit
    function.CodeGenerator.Emit = localEmitter

    /* Have to process the body first to find out how many locals there are */
    var functionError error
    go func(){
        defer close(localEmitter)
        _, functionError = ast.Body.Visit(function)
    }()

    var code []string

    for line := range localEmitter {
        code = append(code, line)
    }

    if functionError != nil {
        return nil, functionError
    }

    function.CodeGenerator.Emit = savedEmit
    function.CodeGenerator.Emit <- fmt.Sprintf("function %v.%v %v", function.CodeGenerator.ClassName, ast.Name, function.LocalCount)
    for _, line := range code {
        function.CodeGenerator.Emit <- line
    }

    return nil, nil
}

func (function *FunctionGenerator) VisitMethod(ast *ASTMethod) (interface{}, error) {
    /* has a this object available */
    function.CodeGenerator.Emit <- fmt.Sprintf("function %v.%v %v", function.CodeGenerator.ClassName, ast.Name, len(ast.Parameters))

    _, err := ast.Body.Visit(function)
    if err != nil {
        return nil, err
    }

    return nil, nil
}

type CodeGenerator struct {
    Emit chan(string)
    ClassName string
}

func (generator *CodeGenerator) VisitClass(ast *ASTClass) (interface{}, error) {
    generator.ClassName = ast.Name

    /* FIXME: visit fields/statics first to build up the symbol table,
     * and then visit methods and functions
     */
    for _, body := range ast.Body {
        _, err := body.Visit(generator)
        if err != nil {
            return nil, err
        }
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

func (generator *CodeGenerator) VisitMethod(*ASTMethod) (interface{}, error) {
    return nil, fmt.Errorf("unimplemented method")
}

func (generator *CodeGenerator) VisitBlock(ast *ASTBlock) (interface{}, error) {
    return nil, fmt.Errorf("code generator should not visit block")
}

func (generator *CodeGenerator) VisitFunction(ast *ASTFunction) (interface{}, error) {
    function := FunctionGenerator{
        CodeGenerator: generator,
        LocalVariables: make(map[string]int),
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

func (generator *CodeGenerator) VisitStatic(*ASTStatic) (interface{}, error) {
    return nil, fmt.Errorf("unimplemented static")
}

func (generator *CodeGenerator) VisitField(*ASTField) (interface{}, error) {
    return nil, fmt.Errorf("unimplemented field")
}

func GenerateCode(ast ASTNode, writer io.Writer) error {
    vmChannel := make(chan string, 10)
    generator := CodeGenerator{
        Emit: vmChannel,
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
