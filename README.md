sql-parser
==========

A Go library for lexing and parsing SQL strings and streams.

sql-parser requires Go version 1.2 or greater.


Status & Roadmap
----------------
**This is an unfinished project. Estimated release date of the first version is Feb 2014.**


Limitations
-----------
The planned features for the first version include only support for MySQL DML statements (SELECT, INSERT, UPDATE, DELETE).

Other parts of SQL like DDL (CREATE, DROP, ALTER) and DCL (GRANT, REVOKE) and other SQL dialects will be implemented in future versions.


Installation
------------
```
go get github.com/na--/sql-parser
```

Usage
-----

```go
import "github.com/na--/sql-parser"
```

### Lexer example ###
```go
TODO
```

### Parser example ###
```go
TODO
```

Credits
-------
The lexer implementation is inspired by [Rob Pike's lecture "Lexical Scanning in Go"](https://www.youtube.com/watch?v=HxaD_trXwRE) and the implementation of ["text/template" package](http://golang.org/pkg/text/template/) in the Go standard library.

Used in
-------
This library will be used in the implementation of the [SQL anonymizer tool](https://github.com/na--/sql-anonymizer)

License
-------
Licensed under the [MIT License](http://opensource.org/licenses/MIT)