package selector

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"go.starlark.net/starlark"
)

// ScriptSelector executes a user-provided Starlark script to select artifacts.
type ScriptSelector struct {
	ScriptPaths []string
}

// NewScriptSelector is the constructor for ScriptSelector.
func NewScriptSelector(paths []string) *ScriptSelector {
	return &ScriptSelector{ScriptPaths: paths}
}

// Select executes the Starlark script and returns the artifacts it defines.
func (s *ScriptSelector) Select(ctx Context) ([]Artifact, error) {
	var allArtifacts []Artifact

	for _, scriptPath := range s.ScriptPaths {
		resolvedPath, err := resolvePath(ctx.Root, scriptPath)
		if err != nil {
			return nil, fmt.Errorf("resolve selector script %q: %w", scriptPath, err)
		}

		thread := &starlark.Thread{Name: "selector-script"}

		// 🛡️ Define a safe, sandboxed API for the script.
		// These are the only functions and data the user's script can access.
		predeclared := starlark.StringDict{
			"walk":      starlark.NewBuiltin("walk", goWalk(ctx)),
			"read_file": starlark.NewBuiltin("read_file", goReadFile(ctx)),
			"env":       buildEnvDict(ctx),
			"consts":    buildConstsDict(ctx),
			"root":      starlark.String(ctx.Root),
		}

		globals, err := starlark.ExecFile(thread, resolvedPath, nil, predeclared)
		if err != nil {
			return nil, fmt.Errorf("error executing selection script %s: %w", resolvedPath, err)
		}

		resultsVal := globals["artifacts"]
		if resultsVal == nil {
			return nil, fmt.Errorf("script %s must define a global 'artifacts' variable", resolvedPath)
		}

		artifacts, err := parseStarlarkResults(resultsVal)
		if err != nil {
			return nil, fmt.Errorf("failed to parse artifacts from script %s: %w", resolvedPath, err)
		}

		allArtifacts = append(allArtifacts, artifacts...)
	}

	return allArtifacts, nil
}
// --- Starlark Builtin Implementations ---

// goWalk exposes a filesystem walk to Starlark.
func goWalk(ctx Context) func(*starlark.Thread, *starlark.Builtin, starlark.Tuple, []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		root := "."
		if err := starlark.UnpackArgs("walk", args, kwargs, "root?", &root); err != nil {
			return nil, err
		}

		walkRoot, err := resolvePath(ctx.Root, root)
		if err != nil {
			return nil, err
		}

		var paths []starlark.Value
		err = filepath.WalkDir(walkRoot, func(path string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() {
				paths = append(paths, starlark.String(path))
			}
			return err
		})
		if err != nil {
			return nil, err
		}
		return starlark.NewList(paths), nil
	}
}

// goReadFile exposes a file reader to Starlark.
func goReadFile(ctx Context) func(*starlark.Thread, *starlark.Builtin, starlark.Tuple, []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var path string
		if err := starlark.UnpackArgs("read_file", args, kwargs, "path", &path); err != nil {
			return nil, err
		}

		resolvedPath, err := resolvePath(ctx.Root, path)
		if err != nil {
			return nil, err
		}

		content, err := os.ReadFile(resolvedPath)
		if err != nil {
			return nil, err
		}
		return starlark.String(string(content)), nil
	}
}


// --- Helper Functions ---

// buildEnvDict converts the context's allowed environment variables into a Starlark dictionary.
func buildEnvDict(ctx Context) *starlark.Dict {
	dict := starlark.NewDict(len(ctx.EnvVars))
	for k, v := range ctx.EnvVars {
		dict.SetKey(starlark.String(k), starlark.String(v))
	}
	return dict
}

func buildConstsDict(ctx Context) *starlark.Dict {
	dict := starlark.NewDict(len(ctx.Consts))
	for k, v := range ctx.Consts {
		dict.SetKey(starlark.String(k), starlark.String(v))
	}
	return dict
}

// resolvePath ensures the candidate path stays inside the selector root.
func resolvePath(root, candidate string) (string, error) {
	cleanRoot := filepath.Clean(root)
	if !filepath.IsAbs(cleanRoot) {
		absRoot, err := filepath.Abs(cleanRoot)
		if err != nil {
			return "", err
		}
		cleanRoot = absRoot
	}

	cleanCandidate := filepath.Clean(candidate)
	var resolved string
	switch {
	case cleanCandidate == "":
		resolved = cleanRoot
	case filepath.IsAbs(cleanCandidate):
		resolved = cleanCandidate
	default:
		resolved = filepath.Join(cleanRoot, cleanCandidate)
	}

	resolved = filepath.Clean(resolved)
	rel, err := filepath.Rel(cleanRoot, resolved)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("path %q is outside selector root", candidate)
	}

	return resolved, nil
}

// parseStarlarkResults converts the Starlark return value into a Go slice of Artifacts.
func parseStarlarkResults(val starlark.Value) ([]Artifact, error) {
	list, ok := val.(*starlark.List)
	if !ok {
		return nil, fmt.Errorf("expected 'artifacts' to be a list, but got %s", val.Type())
	}

	var artifacts []Artifact
	iter := list.Iterate()
	defer iter.Done()

	var item starlark.Value
	for iter.Next(&item) {
		dict, ok := item.(*starlark.Dict)
		if !ok {
			return nil, fmt.Errorf("expected artifact list item to be a dict, but got %s", item.Type())
		}

		pathVal, _, _ := dict.Get(starlark.String("path"))
		typeVal, _, _ := dict.Get(starlark.String("type"))

		pathStr, ok := pathVal.(starlark.String)
		if !ok {
			return nil, fmt.Errorf("artifact 'path' must be a string")
		}
		typeStr, ok := typeVal.(starlark.String)
		if !ok {
			return nil, fmt.Errorf("artifact 'type' must be a string")
		}

		artifacts = append(artifacts, Artifact{
			Path: string(pathStr),
			Type: string(typeStr),
		})
	}

	return artifacts, nil
}
