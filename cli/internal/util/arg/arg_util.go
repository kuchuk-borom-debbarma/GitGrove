package arg

// parseArgs parses the raw arguments into a map of arguments so that its easy to work with
func ParseArg(rawArgs []string) map[string]any {
	args := make(map[string]any)
	for i := 0; i < len(rawArgs); i++ {
		arg := rawArgs[i]
		// Handle flags
		if len(arg) > 0 && arg[0] == '-' {
			// Remove leading dashes
			key := arg
			for len(key) > 0 && key[0] == '-' {
				key = key[1:]
			}

			// Check for value
			if i+1 < len(rawArgs) && (len(rawArgs[i+1]) == 0 || rawArgs[i+1][0] != '-') {
				args[key] = rawArgs[i+1]
				i++ // Skip next arg since we consumed it
			} else {
				args[key] = true // Boolean flag
			}
		} else {
			// Positional argument, maybe store in a special key or just ignore for now based on requirement
			// For this specific requirement, we might just need flags, but let's store them as "args" list
			if val, ok := args["args"]; ok {
				if list, ok := val.([]string); ok {
					args["args"] = append(list, arg)
				}
			} else {
				args["args"] = []string{arg}
			}
		}
	}
	return args
}
