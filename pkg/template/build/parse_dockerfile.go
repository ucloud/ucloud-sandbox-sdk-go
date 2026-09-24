package build

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/ucloud/ucloud-sandbox-sdk-go/pkg/api"
)

// FromDockerfile builds a template from a Dockerfile on disk. The file's
// directory becomes the file context, so COPY sources resolve relative to it.
func FromDockerfile(path string) (*Builder, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("template: read Dockerfile %q: %w", path, err)
	}

	builder, err := parseDockerfile(filepath.Dir(path), bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("template: parse Dockerfile %q: %w", path, err)
	}

	return builder, nil
}

func parseDockerfile(contextPath string, r io.Reader) (*Builder, error) {
	result, err := parser.Parse(r)
	if err != nil {
		return nil, err
	}

	var builder *Builder
	for _, node := range result.AST.Children {
		name := strings.ToUpper(node.Value)
		lineNo := node.StartLine
		args := dockerfileNodeArgs(node)

		if StepType(name) != StepFrom && builder == nil {
			return nil, fmt.Errorf("line %d: FROM should be the first instruction", lineNo)
		}

		switch StepType(name) {
		case StepFrom:
			if len(args) < 1 || (len(args) != 1 && !(len(args) == 3 && strings.EqualFold(args[1], "AS"))) {
				return nil, fmt.Errorf("line %d: FROM requires exactly one image", lineNo)
			}
			if builder != nil {
				return nil, fmt.Errorf("line %d: multiple FROM instructions are not supported", lineNo)
			}
			builder = FromImage(args[0])
			builder.SetFileContextPath(contextPath)

		case StepRun:
			if len(args) == 0 {
				return nil, fmt.Errorf("line %d: RUN requires a command", lineNo)
			}
			builder.RunCmd(dockerfileCommand(args))

		case StepCopy, StepAdd:
			if len(args) < 2 {
				return nil, fmt.Errorf("line %d: %s requires source and destination", lineNo, name)
			}
			// One COPY may name several sources; the build API takes one per
			// step, so they are split out here.
			for _, src := range args[:len(args)-1] {
				dest := args[len(args)-1]
				builder.Copy(src, dest)
			}

		case StepEnv:
			if len(args) < 2 {
				return nil, fmt.Errorf("line %d: invalid ENV instruction", lineNo)
			}
			envs, err := dockerfileEnvArgs(node, args)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", lineNo, err)
			}
			builder.SetEnvs(envs)

		case StepArg:
			if len(args) == 0 {
				return nil, fmt.Errorf("line %d: invalid ARG instruction", lineNo)
			}
			for _, arg := range args {
				key, value, hasValue := strings.Cut(arg, "=")
				if key == "" || strings.ContainsAny(key, " \t") {
					return nil, fmt.Errorf("line %d: invalid ARG assignment", lineNo)
				}
				instArgs := []string{key}
				if hasValue {
					instArgs = append(instArgs, value)
				}
				builder.AddStep(api.TemplateStep{
					Type: string(StepArg),
					Args: &instArgs,
				})
			}

		case StepWorkdir:
			if len(args) == 0 {
				return nil, fmt.Errorf("line %d: WORKDIR requires a path", lineNo)
			}
			builder.SetWorkdir(args[0])

		case StepUser:
			if len(args) == 0 {
				return nil, fmt.Errorf("line %d: USER requires a user", lineNo)
			}
			builder.SetUser(args[0])

		case StepEntrypoint, StepCmd:
			if len(args) == 0 {
				return nil, fmt.Errorf("line %d: %s requires a command", lineNo, name)
			}
			builder.SetStartCmd(dockerfileCommand(args))
		}
	}

	if builder == nil {
		return nil, fmt.Errorf("FROM instruction is required")
	}

	return builder, nil
}

// dockerfileCommand joins a node's arguments back into a shell command.
func dockerfileCommand(args []string) string {
	return strings.Join(args, " ")
}

// dockerfileEnvArgs reads the key/value pairs out of an ENV or ARG node.
func dockerfileEnvArgs(node *parser.Node, args []string) (map[string]string, error) {
	envs := make(map[string]string)

	if strings.EqualFold(node.Value, "ARG") {
		for _, arg := range args {
			key, value, _ := strings.Cut(arg, "=")
			if key == "" || strings.ContainsAny(key, " \t") {
				return nil, fmt.Errorf("invalid ARG assignment")
			}
			envs[key] = value
		}
		return envs, nil
	}

	// BuildKit reports ENV key=value as a key, value, "=" triple, and the
	// older "ENV key value" form as a pair.
	switch {
	case len(args) == 3 && args[2] == "":
		envs[args[0]] = args[1]
		return envs, nil
	case len(args) == 2:
		envs[args[0]] = args[1]
		return envs, nil
	case len(args)%3 != 0:
		return nil, fmt.Errorf("invalid ENV assignment")
	}

	for i := 0; i < len(args); i += 3 {
		if args[i] == "" || args[i+2] != "=" {
			return nil, fmt.Errorf("invalid ENV assignment")
		}
		envs[args[i]] = args[i+1]
	}
	return envs, nil
}

// dockerfileNodeArgs flattens a parsed node's argument list.
func dockerfileNodeArgs(node *parser.Node) []string {
	var args []string
	for arg := node.Next; arg != nil; arg = arg.Next {
		args = append(args, arg.Value)
	}
	return args
}

// dockerfileChown reads the --chown flag off a COPY or ADD node.
func dockerfileChown(node *parser.Node) string {
	for _, flag := range node.Flags {
		if strings.HasPrefix(strings.ToLower(flag), "--chown=") {
			return flag[len("--chown="):]
		}
	}
	return ""
}
