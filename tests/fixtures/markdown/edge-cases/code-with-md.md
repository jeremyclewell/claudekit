# Code Blocks Containing Markdown

This tests that markdown inside code blocks is NOT formatted.

## Example 1

```markdown
##This should NOT be fixed
Because it's inside a code fence.

* Mixed
- Markers
+ Should stay as-is

    Indented code in markdown example
```

## Example 2

```python
# This is a Python comment, not a heading
def bad_markdown_example():
    """
    ##Missing space - but it's in a docstring
    * Mixed list markers in docs
    """
    return "Don't format code"
```

## Example 3

The formatter must preserve code block contents byte-for-byte:

```text
Intentional     extra    spaces
Trailing spaces should stay
Mixed			tabs and spaces
```
