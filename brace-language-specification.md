# RFC: BRACE Configuration Language Specification

## Abstract

BRACE is a configuration markup language designed to address limitations in existing formats like YAML, TOML, and JSON. It provides whitespace-insensitive syntax, clear data structure definitions, directive-based behaviors, and table organization while maintaining JSON compatibility.

## 1. Introduction

BRACE (Brackets, References, Arrays, Configuration, Environment) eliminates common configuration file pain points:
- No whitespace sensitivity issues
- Clear syntax for complex nested structures
- Extensible directive system for dynamic behaviors
- Table-based organization for logical grouping
- Direct JSON output capability

## 2. Language Overview

### 2.1 Basic Syntax
- Case-sensitive identifiers
- C-style comments (`//` and `/* */`)
- Semicolons optional
- Whitespace insensitive

### 2.2 Data Types
- **String**: Double-quoted (`"text"`) or triple-quoted (`"""multiline"""`)
- **Number**: Integer or decimal (`123`, `-45.67`)
- **Boolean**: `true` or `false`
- **Null**: `null`
- **Object**: `{ key = value, ... }`
- **Array**: `[ value1, value2, ... ]`

### 2.3 Special Constructs
- **Directives**: `@directive` - Compile-time behaviors
- **Constants**: `@const` - Reusable values with namespacing
- **References**: `:namespace.CONSTANT` - Constant references
- **Tables**: `#table.subtable` - Organizational sections

## 3. EBNF Grammar

```ebnf
(* BRACE Configuration Language Grammar *)

braceFile = versionDirective, { item } ;

versionDirective = "@brace", string ;

item = assignment
     | directive
     | table
     | comment ;

assignment = identifier, "=", value ;

value = string
      | number
      | boolean
      | null
      | object
      | array
      | reference ;

(* Basic Types *)
string = doubleQuotedString | tripleQuotedString ;
doubleQuotedString = '"', { stringChar }, '"' ;
tripleQuotedString = '"""', { multilineChar }, '"""' ;
stringChar = ? any character except '"' and newline ? ;
multilineChar = ? any character ? ;

number = [ "-" ], ( integer | decimal ) ;
integer = digit, { digit } ;
decimal = digit, { digit }, ".", digit, { digit } ;
digit = "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9" ;

boolean = "true" | "false" ;
null = "null" ;

(* Complex Types *)
object = "{", [ objectMember, { ",", objectMember } ], "}" ;
objectMember = identifier, "=", value ;

array = "[", [ value, { ",", value } ], "]" ;

reference = ":", [ namespace, "." ], identifier ;
namespace = identifier ;

(* Directives *)
directive = "@", directiveName, [ directiveParams ] ;
directiveName = "const" | "env" | identifier ;

directiveParams = string, objectBody
                | objectBody
                | "(" paramList ")" ;

paramList = value, { ",", value } ;
objectBody = "{", { assignment }, "}" ;

(* Tables *)
table = "#", tablePath, objectBody ;
tablePath = identifier, { ".", identifier } ;

(* Comments *)
comment = singleLineComment | multiLineComment ;
singleLineComment = "//", { ? any character except newline ? }, newline ;
multiLineComment = "/*", { ? any character ? }, "*/" ;

(* Identifiers *)
identifier = letter, { letter | digit | "_" } ;
letter = "a" | "b" | "c" | "d" | "e" | "f" | "g" | "h" | "i" | "j" 
       | "k" | "l" | "m" | "n" | "o" | "p" | "q" | "r" | "s" | "t" 
       | "u" | "v" | "w" | "x" | "y" | "z" | "A" | "B" | "C" | "D" 
       | "E" | "F" | "G" | "H" | "I" | "J" | "K" | "L" | "M" | "N" 
       | "O" | "P" | "Q" | "R" | "S" | "T" | "U" | "V" | "W" | "X" 
       | "Y" | "Z" ;
```

## 4. Directive Specifications

### 4.1 @const Directive
Declares reusable constants with optional namespacing.

**Syntax:**
```
@const { assignments }
@const "namespace" { assignments }
```

**Behavior:**
- Without namespace: constants stored in `global` namespace
- With namespace: constants stored in specified namespace
- Constants referenced via `:namespace.CONSTANT` or `:CONSTANT` (global)

### 4.2 @env Directive
Retrieves environment variable values with optional defaults.

**Syntax:**
```
@env("VAR_NAME")
@env("VAR_NAME", defaultValue)
```

**Behavior:**
- Returns environment variable value if exists
- Returns default value if provided and variable doesn't exist
- Compilation error if variable doesn't exist and no default provided

### 4.3 @brace Directive
Specifies BRACE language version for compatibility.

**Syntax:**
```
@brace "version"
```

## 5. Table System

Tables provide hierarchical organization of configuration data.

**Syntax:**
```
#tableName { assignments }
#parentTable.childTable { assignments }
```

**Output:** Tables convert to nested JSON objects.

## 6. Compilation Process

### 6.1 Phase 1: Lexical Analysis
- Tokenize input according to grammar
- Handle comments (strip from output)
- Validate basic syntax

### 6.2 Phase 2: Parsing
- Build Abstract Syntax Tree (AST)
- Validate structure against grammar
- Identify directives, tables, and assignments

### 6.3 Phase 3: Directive Processing
- Execute `@const` directives to build symbol table
- Process `@env` directives with environment lookups
- Validate all references can be resolved

### 6.4 Phase 4: Reference Resolution
- Replace all `:namespace.CONSTANT` references with actual values
- Validate all references exist in declared namespaces

### 6.5 Phase 5: JSON Generation
- Convert tables to nested JSON objects
- Output valid JSON to stdout

## 7. Error Handling

### 7.1 Compilation Errors
- Missing environment variables (no default provided)
- Undefined constant references
- Invalid syntax
- Type mismatches in arrays

### 7.2 Error Reporting
- Line and column numbers for syntax errors
- Clear messages for missing references
- Environment variable resolution failures

## 8. JSON Output Format

BRACE compiles to standard JSON with the following mappings:

| BRACE Type | JSON Type |
|------------|-----------|
| String | String |
| Number | Number |
| Boolean | Boolean |
| Null | null |
| Object | Object |
| Array | Array |
| Table | Object (nested) |

## 9. Example Compilation

**Input (BRACE):**
```brace
@brace "0.0.1"
@const { VERSION = "1.0.0" }
app_version = :VERSION
#database {
    host = "localhost"
    port = 5432
}
```

**Output (JSON):**
```json
{
  "app_version": "1.0.0",
  "database": {
    "host": "localhost",
    "port": 5432
  }
}
```

## 10. Implementation Notes

- Arrays must contain homogeneous types
- Identifiers must start with letter, contain only alphanumeric and underscore
- String interpolation not supported (by design)
- Directive system designed for extensibility
- Memory-efficient parsing recommended for large files

## 11. Security Considerations

- Environment variable access should be controlled
- File inclusion directives (if added) need sandboxing
- Validate all external data sources

## 12. Future Extensions

- Include directive for file composition
- Conditional compilation directives
- Schema validation directives
- Custom directive plugins