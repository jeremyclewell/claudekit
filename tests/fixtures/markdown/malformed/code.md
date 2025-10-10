# Code Block Issues

Indented code block (should be fenced):

    function hello() {
        console.log("Hello, world!");
    }

More text here.

Tildes instead of backticks:

~~~javascript
function goodbye() {
    console.log("Goodbye!");
}
~~~

Missing language identifier:

```
def python_function():
    return "No language tag"
```

Correct fenced code:

```go
func correct() {
    fmt.Println("This is correct")
}
```
