package command

import "strings"

type tagArgs struct {
    Add    []string
    Remove []string
    Others []string
}

func extractTagArgs(args []string) tagArgs {
	var result tagArgs
	for _, arg := range args {
		if tag, ok := parseAddTag(arg); ok {
			if tag != "" {
				result.Add = append(result.Add, tag)
			}
			continue
		}
		if tag, ok := parseRemoveTag(arg); ok {
			if tag != "" {
				result.Remove = append(result.Remove, tag)
			}
			continue
		}
		result.Others = append(result.Others, arg)
	}
	return result
}

func parseAddTag(arg string) (string, bool) {
	if len(arg) < 2 {
		return "", false
	}
	switch arg[0] {
	case '@', '#':
		return strings.TrimSpace(arg[1:]), true
	case '+':
		if len(arg) > 1 && (arg[1] == '@' || arg[1] == '#') {
			return strings.TrimSpace(arg[2:]), true
		}
	}
	return "", false
}

func parseRemoveTag(arg string) (string, bool) {
    if len(arg) < 2 {
        return "", false
    }
    if arg[0] == '-' && (arg[1] == '@' || arg[1] == '#') {
        return strings.TrimSpace(arg[2:]), true
    }
    return "", false
}

func splitTags(raw string) []string {
    if strings.TrimSpace(raw) == "" {
        return nil
    }
    parts := strings.Split(raw, ",")
    var tags []string
    seen := make(map[string]struct{})
    for _, p := range parts {
        tag := strings.TrimSpace(strings.ToLower(p))
        if tag == "" {
            continue
        }
        if _, ok := seen[tag]; ok {
            continue
        }
        seen[tag] = struct{}{}
        tags = append(tags, tag)
    }
    return tags
}
