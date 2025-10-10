# Example Command Template

This is a sample slash command demonstrating the command system architecture and best practices.

## Purpose

This template shows you how to create effective custom slash commands with proper structure, documentation, and error handling.

## Command Structure

A well-designed slash command should include:

1. **Clear Purpose Statement**: One sentence describing what the command does
2. **Context and Role**: What persona or expertise the AI should adopt
3. **Process Steps**: Numbered, actionable steps to execute
4. **Input Handling**: How to parse and validate command arguments
5. **Output Format**: Expected deliverables and format
6. **Error Handling**: How to handle invalid inputs or edge cases

## Usage Example

```
/project:example [argument1] [argument2]
```

### Arguments

- `argument1` (required): Description of first argument
- `argument2` (optional): Description of second argument, defaults to "default value"

## Implementation Process

When this command is invoked:

1. **Parse Arguments**
   - Extract command arguments from user input
   - Validate required arguments are present
   - Apply default values for optional arguments
   - Return clear error if validation fails

2. **Execute Task**
   - Perform the main task using provided arguments
   - Follow best practices for the task type
   - Handle errors gracefully
   - Provide progress updates for long-running tasks

3. **Format Output**
   - Return results in specified format
   - Include summary of what was done
   - Provide next steps or recommendations
   - Offer relevant follow-up commands

## Best Practices for Command Design

### Do:
- ✅ Write clear, concise command descriptions
- ✅ Provide examples of usage
- ✅ Validate inputs before processing
- ✅ Give helpful error messages
- ✅ Document expected behavior
- ✅ Include edge case handling

### Don't:
- ❌ Make commands too complex
- ❌ Use ambiguous argument names
- ❌ Forget to validate inputs
- ❌ Silently fail on errors
- ❌ Assume default values without documenting

## Error Handling

Handle common error cases:

```
if (!argument1) {
  return "Error: argument1 is required. Usage: /project:example <argument1> [argument2]"
}

if (typeof argument1 !== 'expected_type') {
  return "Error: argument1 must be a <expected_type>"
}
```

## Output Format

Return structured output:

```
## Command Results

**Task**: Example command execution
**Status**: ✅ Success

### Details
- Argument 1: <value>
- Argument 2: <value>

### Actions Taken
1. Step one completed
2. Step two completed
3. Step three completed

### Next Steps
- Suggested follow-up action 1
- Suggested follow-up action 2
```

## Creating Your Own Commands

To create a new command:

1. Copy this template to `assets/templates/your-command.md`
2. Update the command description and purpose
3. Define your process steps
4. Specify input arguments
5. Document output format
6. Add command module in `assets/modules/commands/your-command.md`
7. Set `asset_paths: ["templates/your-command.md"]` in the module

## Tips

- Keep commands focused on a single task
- Use clear, action-oriented names
- Provide examples in the documentation
- Test with various inputs
- Consider command composability
