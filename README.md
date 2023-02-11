# Easy Scan
Easy Scan is a code generator that implements functions on structures to map the Rows of an SQL query to the fields of an object or a slice of objects with no reflections.

```
go install github.com/Soreing/easyscan/...
```

## Generation

Define a structure to generate scan functions for, then run the generator. The default behavior is that fields are scanned in the order they are defined in, but it can be changed with the `-any_order` flag.

`./models/book.go`
```golang
type Book struct {
	Id          string
	Title       string
	Description string
	Author      string
	Year        int
	Version     int
}
```
You can either target a single file or an entire directory by providing the relative path from the module root. When targeting a single file, the result is written into `[filename]_easyscan.go`. When targeting a directory, the result is written into `[pkgname]_easyscan.go`. The output file name can also be custom set with flags. 
```
easyscan -all ./models/book.go
```

The `-all` flag processes all defined types, but you can explicitly include or exclude types from the generation. The following example includes `Movie` without the use of the all flag, and excludes `Shape` even with the all flag
```golang
//easyscan:explicit
type Movie struct {
    Title string
    Time  time.Time
}

//easyscan:skip
type Shape struct {
    Sides int
    Color string
}
```

The `easyscan` field tag lets you customize how specific fields behave. The first value specifies an alias for the field for parsing rows. You can also use `omit` to exclude the field from scanning.
```golang
type User struct {
    Id       string `easyscan:"_Id"`
    Name     string
    Password string `easyscan:",omit"`
}
```

To generate scan functions for a slice of objects, you need to define and use a new type that wraps the slice.
```golang
type BookList []Book
```

## Usage

Use the generated method `ScanRow` to scan a single row into an object. You need to provide sql.Rows structure even if there is only one row.
```golang 
// Query the database
rows, err := db.Query("SELECT * FROM books WHERE id=$1", 0)
if err != nil {
    panic(err)
}

// Scan the row into a Book object
bk := models.Book{}
if err := bk.ScanRow(rows); err != nil {
    fmt.Println("failed to scan")
}
fmt.Println("Book: ", bk)
```

When you have multiple rows to scan, you can iterate over the sql.Rows structure and append into the slice with the `ScanAppendRow` generated method.
```golang 
// Query the database
rows, err := db.Query("SELECT * FROM books WHERE id=$1", 0)
if err != nil {
    panic(err)
}

// Scan the rows into a Book List object
bkl := models.BookList{}
for rows.Next() {
    if err := bkl.ScanAppendRow(rows); err != nil {
        fmt.Println("failed to scan")
        break
    }
}
fmt.Println("Book List: ", bkl)
```

## Options
| Option | Description |
|--------|-------------|
| -all             | Processes all defined types in the target path        |
| -any_order       | Allows scanning fields in any order from rows         |
| -output_filename | Specify the output file name for the code             |
| -lower_case      | Transform field names to lower case for parsing rows  |
| -camel_case      | Transform field names to camel case for parsing rows  |
| -kebab_case      | Transform field names to kebab case for parsing rows  |
| -snake_case      | Transform field names to snake case for parsing rows  |
| -pascal_case     | Transform field names to pascal case for parsing rows |