# Changelog
## [v0.12.17](https://github.com/itchyny/gojq/compare/v0.12.16..v0.12.17) (2024-12-01)
* implement `add/1`, `skip/2` functions
* implement `--library-path` option as the alias of `-L` option
* fix `reduce` syntax to emit results for each initial value
* fix `last/1` to yield no values when the argument yields no values
* fix `limit/2` to emit an error on negative count
* fix `@uri` and `@urid` formats not to convert space between plus sign
* fix resolving search paths of import statements in the query
* improve time functions to accept fewer element arrays

## [v0.12.16](https://github.com/itchyny/gojq/compare/v0.12.15..v0.12.16) (2024-06-01)
* fix offset of query parsing error on multi-byte characters
* fix tests of `exp10` and `atan2` failing on some platforms
* fix `debug/1` to be available only when `debug/0` is defined
* improve parser to allow binary operators as object values
* improve compiler to emit error if query is missing

## [v0.12.15](https://github.com/itchyny/gojq/compare/v0.12.14..v0.12.15) (2024-04-01)
* implement `ltrim`, `rtrim`, and `trim` functions
* implement `gojq.ParseError` for getting the offset and token of query parsing error
* implement `gojq.HaltError` for detecting halt errors and stopping outer iteration
* fix object construction with duplicate keys (`{x:0,y:1} | {a:.x,a:.y}`)
* fix `halt` and `halt_error` functions to stop the command execution immediately
* fix variable scope of binding syntax (`"a" as $v | def f: $v; "b" as $v | f`)
* fix pre-defined variables to be available in initial modules (`$ARGS` in `~/.jq`)
* fix `ltrimstr` and `rtrimstr` functions to emit error on non-string input
* fix `nearbyint` and `rint` functions to round ties to even
* improve parser to allow `reduce`, `foreach`, `if`, `try`-`catch` syntax as object values
* remove `pow10` in favor of `exp10`, define `scalbn` and `scalbln` by `ldexp`

## [v0.12.14](https://github.com/itchyny/gojq/compare/v0.12.13..v0.12.14) (2023-12-01)
* implement `abs`, `pick`, and `debug/1` functions
* implement `--raw-output0` option, and remove `--nul-output` (`-0`) option
* fix string multiplication by zero to emit an empty string
* fix zero divided by zero to emit an error, not `nan`
* fix modulo operator to emit `nan` if either side is `nan`
* fix `implode` function to emit replacement characters on invalid code points
* fix `stderr` function to output strings in raw format
* fix `error` function to throw an error even for `null`
* fix `walk` function on multiple outputs arguments
* fix `--from-file` option to work with `--args` and `--jsonargs` options
* fix the default module search path `../lib` relative to the executable
* improve query parser to support comment continuation with backslash
* improve `modulemeta` function to include defined function names in the module
* improve search path of `import` and `include` directives to support `$ORIGIN` expansion
* remove deprecated `leaf_paths` function

## [v0.12.13](https://github.com/itchyny/gojq/compare/v0.12.12..v0.12.13) (2023-06-01)
* implement `@urid` format string to decode URI values
* fix functions returning arrays not to emit nil slices (`flatten`, `group_by`,
  `unique`, `unique_by`, `nth`, `indices`, `path`, and `modulemeta.deps`)

## [v0.12.12](https://github.com/itchyny/gojq/compare/v0.12.11..v0.12.12) (2023-03-01)
* fix assignment operator (`=`) with overlapping paths and multiple values (`[[]] | .. = ..`)
* fix crash on multiplying large numbers to an empty string (`9223372036854775807 * ""`)
* improve zsh completion file

## [v0.12.11](https://github.com/itchyny/gojq/compare/v0.12.10..v0.12.11) (2022-12-24)
* fix crash on assignment operator (`=`) with multiple values (`. = (0,0)`)
* fix `isnormal` and `normals` functions against subnormal numbers

## [v0.12.10](https://github.com/itchyny/gojq/compare/v0.12.9..v0.12.10) (2022-12-01)
* fix `break` in `try`-`catch` query (`label $x | try break $x catch .`)
* fix path value validation for `getpath` function (`path(getpath([[0]][0]))`)
* fix path value validation for custom iterator functions
* fix `walk` function with argument emitting multiple values (`[1],{x:1} | walk(.,0)`)
* fix `@csv`, `@tsv`, `@sh` to escape the null character (`["\u0000"] | @csv,@tsv,@sh`)
* improve performance of assignment operator (`=`), update-assignment operator (`|=`),
  `map_values`, `del`, `delpaths`, `walk`, `ascii_downcase`, and `ascii_upcase` functions

## [v0.12.9](https://github.com/itchyny/gojq/compare/v0.12.8..v0.12.9) (2022-09-01)
* fix `fromjson` to emit error on unexpected trailing string
* fix path analyzer on variable argument evaluation (`def f($x): .y; path(f(.x))`)
* fix raw input option `--raw-input` (`-R`) to keep carriage returns and support 64KiB+ lines

## [v0.12.8](https://github.com/itchyny/gojq/compare/v0.12.7..v0.12.8) (2022-06-01)
* implement `gojq.Compare` for comparing values in custom internal functions
* implement `gojq.TypeOf` for obtaining type name of values in custom internal functions
* implement `gojq.Preview` for previewing values for error messages of custom internal functions
* fix query lexer to parse string literals as JSON to support surrogate pairs (`"\ud83d\ude04"`)
* fix priority bug of declared and builtin functions (`def empty: .; null | select(.)`)
* fix string indexing by index out of bounds to emit `null` (`"abc" | .[3]`)
* fix array binding pattern not to match against strings (`"abc" as [$a] ?// $a | $a`)
* fix `sub` and `gsub` functions to emit results in the same order of jq
* fix `fromjson` to keep integer precision (`"10000000000000000" | fromjson + 1`)
* fix stream option to raise error against incomplete JSON input
* improve array updating index and string repetition to increase limitations
* improve `mktime` to support nanoseconds, just like `gmtime` and `now`
* improve query lexer to report unterminated string literals
* improve performance of string indexing and slicing by reducing allocations
* improve performance of object and array indexing, slicing, and iteration,
  by validating path values by comparing data addresses. This change improves jq
  compatibility of path value validation (`{} | {}.x = 0`, `[0] | [.[]][] = 1`).
  Also optimize constant indexing and slicing by specialized instruction
* improve performance of `add` (on array of strings), `flatten`, `min`, `max`,
  `sort`, `unique`, `join`, `to_entries`, `from_entries`, `indices`, `index`,
  `rindex`, `startswith`, `endswith`, `ltrimstr`, `rtrimstr`, `explode`,
  `capture`, `sub`, and `gsub` functions

## [v0.12.7](https://github.com/itchyny/gojq/compare/v0.12.6..v0.12.7) (2022-03-01)
* fix precedence of try expression against operators (`try 0 * error(0)`)
* fix iterator suffix with optional operator (`0 | .x[]?`)
* fix stream option with slurp option or `input`, `inputs` functions
* fix the command flag parser to support equal sign in short options with argument
* fix string conversion of query including empty strings in module and import metadata
* improve performance of `isempty` function

## [v0.12.6](https://github.com/itchyny/gojq/compare/v0.12.5..v0.12.6) (2021-12-01)
* implement options for consuming remaining arguments (`--args`, `--jsonargs`, `$ARGS.positional`)
* fix `delpaths` function with overlapped paths
* fix `--exit-status` flag with `halt`, `halt_error` functions
* fix `input_filename` function with null input option
* fix path value validation for `nan`
* fix crash on branch optimization (`if 0 then . else 0|0 end`)
* add validation on regular expression flags to reject unsupported ones
* improve performance of `range`, `join`, `flatten` functions
* improve constant value optimization for object with quoted keys
* remove dependency on forked `go-flags` package

## [v0.12.5](https://github.com/itchyny/gojq/compare/v0.12.4..v0.12.5) (2021-09-01)
* implement `input_filename` function for the command
* fix priority bug of declared functions and arguments (`def g: 1; def f(g): g; f(2)`)
* fix label handling to catch the correct break error (`first((0, 0) | first(0))`)
* fix `null|error` and `error(null)` to behave like `empty` (`null | [0, error, error(null), 1]`)
* fix integer division to keep precision when divisible (`1 / 1 * 1000000000000000000000`)
* fix modulo operator on negative number and large number (`(-1) % 10000000000`)
* fix combination of slurp (`--slurp`) and raw input option (`--raw-input`) to keep newlines
* change the default module paths to `~/.jq`, `$ORIGIN/../lib/gojq`, `$ORIGIN/lib`
  where `$ORIGIN` is the directory where the executable is located in
* improve command argument parser to recognize query with leading hyphen,
  allow hyphen for standard input, and force posix style on Windows
* improve `@base64d` to allow input without padding characters
* improve `fromdate`, `fromdateiso8601` to parse date time strings with timezone offset
* improve `halt_error` to print error values without prefix
* improve `sub`, `gsub` to allow the replacement string emitting multiple values
* improve encoding `\b` and `\f` in strings
* improve module loader for search path in query, and absolute path
* improve query lexer to support string literal including newlines
* improve performance of `index`, `rindex`, `indices`, `transpose`, and `walk` functions
* improve performance of value preview in errors and debug mode
* improve runtime performance including tail call optimization
* switch Docker base image to `distroless/static:debug`

## [v0.12.4](https://github.com/itchyny/gojq/compare/v0.12.3..v0.12.4) (2021-06-01)
* fix numeric conversion of large floating-point numbers in modulo operator
* implement a compiler option for adding custom iterator functions
* implement `gojq.NewIter` function for creating a new iterator from values
* implement `$ARGS.named` for listing command line variables
* remove `debug` and `stderr` functions from the library
* stop printing newlines on `stderr` function for jq compatibility

## [v0.12.3](https://github.com/itchyny/gojq/compare/v0.12.2..v0.12.3) (2021-04-01)
* fix array slicing with infinities and large numbers (`[0][-infinite:infinite], [0][:1e20]`)
* fix multiplying strings and modulo by infinities on MIPS 64 architecture
* fix git revision information in Docker images
* release multi-platform Docker images for ARM 64
* switch to `distroless` image for Docker base image

## [v0.12.2](https://github.com/itchyny/gojq/compare/v0.12.1..v0.12.2) (2021-03-01)
* implement `GOJQ_COLORS` environment variable to configure individual colors
* respect `--color-output` (`-C`) option even if `NO_COLOR` is set
* implement `gojq.ValueError` interface for custom internal functions
* fix crash on timestamps in YAML input
* fix calculation on `infinite` (`infinite-infinite | isnan`)
* fix comparison on `nan` (`nan < nan`)
* fix validation of `implode` (`[-1] | implode`)
* fix number normalization for custom JSON module loader
* print error line numbers on invalid JSON and YAML
* improve `strftime`, `strptime` for time zone offsets
* improve performance on reading a large JSON file given by command line argument
* improve performance and reduce memory allocation of the lexer, compiler and executor

## [v0.12.1](https://github.com/itchyny/gojq/compare/v0.12.0..v0.12.1) (2021-01-17)
* skip adding `$HOME/.jq` to module paths when `$HOME` is unset
* fix optional operator followed by division operator (`1?/1`)
* fix undefined format followed by optional operator (`@foo?`)
* fix parsing invalid consecutive dots while scanning a number (`0..[empty]`)
* fix panic on printing a query with `%#v`
* improve performance and reduce memory allocation of `query.String()`
* change all methods of `ModuleLoader` optional

## [v0.12.0](https://github.com/itchyny/gojq/compare/v0.11.2..v0.12.0) (2020-12-24)
* implement tab indentation option (`--tab`)
* implement a compiler option for adding custom internal functions
* implement `gojq.Marshal` function for jq-flavored encoding
* fix slurp option with JSON file arguments
* fix escaping characters in object keys
* fix normalizing negative `int64` to `int` on 32-bit architecture
* fix crash on continuing iteration after emitting an error
* `iter.Next()` does not normalize `NaN` and infinities anymore. Library users
  should take care of them. To handle them for encoding as JSON bytes, use
  `gojq.Marshal`. Also, `iter.Next()` does not clone values deeply anymore for
  performance reason. Users must not update the elements of the returned arrays
  and objects
* improve performance of outputting JSON values by about 3.5 times

## [v0.11.2](https://github.com/itchyny/gojq/compare/v0.11.1..v0.11.2) (2020-10-01)
* fix build for 32bit architecture
* release to [GitHub Container Registry](https://github.com/users/itchyny/packages/container/package/gojq)

## [v0.11.1](https://github.com/itchyny/gojq/compare/v0.11.0..v0.11.1) (2020-08-22)
* improve compatibility of `strftime`, `strptime` functions with jq
* fix YAML input with numbers in keys
* fix crash on multiplying a large number or `infinite` to a string
* fix crash on error while slicing a string (`""[:{}]`)
* fix crash on modulo by a number near 0.0 (`1 % 0.1`)
* include `CREDITS` file in artifacts

## [v0.11.0](https://github.com/itchyny/gojq/compare/v0.10.4..v0.11.0) (2020-07-08)
* improve parsing performance significantly
* rewrite the parser from `participle` library to `goyacc` generated parser
* release to [itchyny/gojq - Docker Hub](https://hub.docker.com/r/itchyny/gojq)
* support string interpolation for object pattern key

## [v0.10.4](https://github.com/itchyny/gojq/compare/v0.10.3..v0.10.4) (2020-06-30)
* implement variable in object key (`. as $x | { $x: 1 }`)
* fix modify operator (`|=`) with `try` `catch` expression
* fix optional operator (`?`) with alternative operator (`//`) in `map_values` function
* fix normalizing numeric types for library users
* export `gojq.NewModuleLoader` function for library users

## [v0.10.3](https://github.com/itchyny/gojq/compare/v0.10.2..v0.10.3) (2020-06-06)
* implement `add`, `unique_by`, `max_by`, `min_by`, `reverse` by internal
  functions for performance and reducing the binary size
* improve performance of `setpath`, `delpaths` functions
* fix assignment against nested slicing (`[1,2,3] | .[1:][:1] = [5]`)
* limit the array index of assignment operator
* optimize constant arrays and objects

## [v0.10.2](https://github.com/itchyny/gojq/compare/v0.10.1..v0.10.2) (2020-05-24)
* implement `sort_by`, `group_by`, `bsearch` by internal functions for performance
  and reducing the binary size
* fix object construction and constant object to allow trailing commas
* fix `tonumber` function to allow leading zeros
* minify the builtin functions to reduce the binary size

## [v0.10.1](https://github.com/itchyny/gojq/compare/v0.10.0..v0.10.1) (2020-04-24)
* fix array addition not to modify the left hand side

## [v0.10.0](https://github.com/itchyny/gojq/compare/v0.9.0..v0.10.0) (2020-04-02)
* implement various functions (`format`, `significand`, `modulemeta`, `halt_error`)
* implement `input`, `inputs` functions
* implement stream option (`--stream`)
* implement slicing with object (`.[{"start": 1, "end": 2}]`)
* implement `NO_COLOR` environment variable support
* implement `nul` output option (`-0`, `--nul-output`)
* implement exit status option (`-e`, `--exit-status`)
* implement `search` field of module meta object
* implement combination of `--yaml-input` and `--slurp`
* improve string token lexer and support nested string interpolation
* improve the exit code for jq compatibility
* improve default module search paths for jq compatibility
* improve documentation for the usage as a library
* change methods of `ModuleLoader` optional, implement `LoadModuleWithMeta` and `LoadJSONWithMeta`
* fix number normalization for JSON arguments (`--argjson`, `--slurpfile`)
* fix `0/0` and `infinite/infinite`
* fix `error` function against `null`

## [v0.9.0](https://github.com/itchyny/gojq/compare/v0.8.0..v0.9.0) (2020-03-15)
* implement various functions (`infinite`, `isfinite`, `isinfinite`, `finites`, `isnormal`, `normals`)
* implement environment variables loader as a compiler option
* implement `$NAME::NAME` syntax for imported JSON variable
* fix modify operator with empty against array (`[range(9)] | (.[] | select(. % 2 > 0)) |= empty`)
* fix variable and function scopes (`{ x: 1 } | . as $x | (.x as $x | $x) | ., $x`)
* fix path analyzer
* fix type check in `startswith` and `endswith`
* ignore type error of `ltrimstr` and `rtrimstr`
* remove nano seconds from `mktime` output
* trim newline at the end of error messages
* improve documents and examples

## [v0.8.0](https://github.com/itchyny/gojq/compare/v0.7.0..v0.8.0) (2020-03-02)
* implement format strings (`@text`, `@json`, `@html`, `@uri`, `@csv`, `@tsv`,
  `@sh`, `@base64`, `@base64d`)
* implement modules feature (`-L` option for directory to search modules from)
* implement options for binding variables from arguments (`--arg`, `--argjson`)
* implement options for binding variables from files (`--slurpfile`, `--rawfile`)
* implement an option for indentation count (`--indent`)
* fix `isnan` for `null`
* fix path analyzer
* fix error after optional operator (`1? | .x`)
* add `$ENV` variable
* add zsh completion file

## [v0.7.0](https://github.com/itchyny/gojq/compare/v0.6.0..v0.7.0) (2019-12-22)
* implement YAML input (`--yaml-input`) and output (`--yaml-output`)
* fix pipe in object value
* fix precedence of `if`, `try`, `reduce` and `foreach` expressions
* release from GitHub Actions

## [v0.6.0](https://github.com/itchyny/gojq/compare/v0.5.0..v0.6.0) (2019-08-26)
* implement arbitrary-precision integer calculation
* implement various functions (`repeat`, `pow10`, `nan`, `isnan`, `nearbyint`,
  `halt`, `INDEX`, `JOIN`, `IN`)
* implement long options (`--compact-output`, `--raw-output`, `--join-output`,
  `--color-output`, `--monochrome-output`, `--null-input`, `--raw-input`,
  `--slurp`, `--from-file`, `--version`)
* implement join output options (`-j`, `--join-output`)
* implement color/monochrome output options (`-C`, `--color-output`,
  `-M`, `--monochrome-output`)
* refactor builtin functions

## [v0.5.0](https://github.com/itchyny/gojq/compare/v0.4.0..v0.5.0) (2019-08-03)
* implement various functions (`with_entries`, `from_entries`, `leaf_paths`,
  `contains`, `inside`, `split`, `stream`, `fromstream`, `truncate_stream`,
  `bsearch`, `path`, `paths`, `map_values`, `del`, `delpaths`, `getpath`,
  `gmtime`, `localtime`, `mktime`, `strftime`, `strflocaltime`, `strptime`,
  `todate`, `fromdate`, `now`, `match`, `test`, `capture`, `scan`, `splits`,
  `sub`, `gsub`, `debug`, `stderr`)
* implement assignment operator (`=`)
* implement modify operator (`|=`)
* implement update operators (`+=`, `-=`, `*=`, `/=`, `%=`, `//=`)
* implement destructuring alternative operator (`?//`)
* allow function declaration inside query
* implement `-f` flag for loading query from file
* improve error message for parsing multiple line query

## [v0.4.0](https://github.com/itchyny/gojq/compare/v0.3.0..v0.4.0) (2019-07-20)
* improve performance significantly
* rewrite from recursive interpreter to stack machine based interpreter
* allow debugging with `make install-debug` and `export GOJQ_DEBUG=1`
* parse built-in functions and generate syntax trees before compilation
* optimize tail recursion
* fix behavior of optional operator
* fix scopes of arguments of recursive function call
* fix duplicate function argument names
* implement `setpath` function

## [v0.3.0](https://github.com/itchyny/gojq/compare/v0.2.0..v0.3.0) (2019-06-05)
* implement `reduce`, `foreach`, `label`, `break` syntax
* improve binding variable syntax to bind to an object or an array
* implement string interpolation
* implement object index by string (`."example"`)
* implement various functions (`add`, `flatten`, `min`, `min_by`, `max`,
  `max_by`, `sort`, `sort_by`, `group_by`, `unique`, `unique_by`, `tostring`,
  `indices`, `index`, `rindex`, `walk`, `transpose`, `first`, `last`, `nth`,
  `limit`, `all`, `any`, `isempty`, `error`, `builtins`, `env`)
* implement math functions (`sin`, `cos`, `tan`, `asin`, `acos`, `atan`,
  `sinh`, `cosh`, `tanh`, `asinh`, `acosh`, `atanh`, `floor`, `round`,
  `rint`, `ceil`, `trunc`, `fabs`, `sqrt`, `cbrt`, `exp`, `exp10`, `exp2`,
  `expm1`, `frexp`, `modf`, `log`, `log10`, `log1p`, `log2`, `logb`,
  `gamma`, `tgamma`, `lgamma`, `erf`, `erfc`, `j0`, `j1`, `y0`, `y1`,
  `atan2/2`, `copysign/2`, `drem/2`, `fdim/2`, `fmax/2`, `fmin/2`, `fmod/2`,
  `hypot/2`, `jn/2`, `ldexp/2`, `nextafter/2`, `nexttoward/2`, `remainder/2`,
  `scalb/2`, `scalbln/2`, `pow/2`, `yn/2`, `fma/3`)
* support object construction with variables
* support indexing against strings
* fix function evaluation for recursive call
* fix error handling of `//` operator
* fix string representation of NaN and Inf
* implement `-R` flag for reading input as raw strings
* implement `-c` flag for compact output
* implement `-n` flag for using null as input value
* implement `-r` flag for outputting raw string
* implement `-s` flag for reading all inputs into an array

## [v0.2.0](https://github.com/itchyny/gojq/compare/v0.1.0..v0.2.0) (2019-05-06)
* implement binding variable syntax (`... as $var`)
* implement `try` `catch` syntax
* implement alternative operator (`//`)
* implement various functions (`in`, `to_entries`, `startswith`, `endswith`,
  `ltrimstr`, `rtrimstr`, `combinations`, `ascii_downcase`, `ascii_upcase`,
  `tojson`, `fromjson`)
* support query for object indexing
* support object construction with variables
* support indexing against strings

## [v0.1.0](https://github.com/itchyny/gojq/compare/v0.0.1..v0.1.0) (2019-05-02)
* implement binary operators (`+`, `-`, `*`, `/`, `%`, `==`, `!=`, `>`, `<`,
  `>=`, `<=`, `and`, `or`)
* implement unary operators (`+`, `-`)
* implement booleans (`false`, `true`), `null`, number and string constant
  values
* implement `empty` value
* implement conditional syntax (`if` `then` `elif` `else` `end`)
* implement various functions (`length`, `utf8bytelength`, `not`, `keys`,
  `has`, `map`, `select`, `recurse`, `while`, `until`, `range`, `tonumber`,
  `type`, `arrays`, `objects`, `iterables`, `booleans`, `numbers`, `strings`,
  `nulls`, `values`, `scalars`, `reverse`, `explode`, `implode`, `join`)
* support function declaration
* support iterators in object keys
* support object construction shortcut
* support query in array indices
* support negative number indexing against arrays
* support json file name arguments

## [v0.0.1](https://github.com/itchyny/gojq/compare/0fa3241..v0.0.1) (2019-04-14)
* initial implementation
