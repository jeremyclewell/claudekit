package templates

import (
	"bytes"
	"embed"
	"text/template"
	"time"

	"jeremyclewell.com/claudekit/internal/config"
	"jeremyclewell.com/claudekit/internal/util"
)

// RenderClaudeMD generates the CLAUDE.md content from the template
func RenderClaudeMD(cfg config.Config, assetsFS embed.FS) string {
	tmplContent, err := assetsFS.ReadFile("assets/templates/CLAUDE.md.tmpl")
	if err != nil {
		panic(err)
	}

	tmpl, err := template.New("claude").Funcs(template.FuncMap{
		"or": util.Or,
	}).Parse(string(tmplContent))
	if err != nil {
		panic(err)
	}

	data := struct {
		config.Config
		HasGo         bool
		HasTypeScript bool
		HasPython     bool
		HasRust       bool
		HasCpp        bool
		HasJava       bool
		HasCsharp     bool
		HasPhp        bool
		HasRuby       bool
		HasSwift      bool
		HasDart       bool
		HasShell      bool
		HasLua        bool
		HasElixir     bool
		HasHaskell    bool
		HasElm        bool
		HasJulia      bool
		HasSql        bool
		Date          string
	}{
		Config:        cfg,
		HasGo:         util.Includes(cfg.Language, "Go"),
		HasTypeScript: util.Includes(cfg.Language, "TypeScript"),
		HasPython:     util.Includes(cfg.Language, "Python"),
		HasRust:       util.Includes(cfg.Language, "Rust"),
		HasCpp:        util.Includes(cfg.Language, "C++"),
		HasJava:       util.Includes(cfg.Language, "Java") || util.Includes(cfg.Language, "Kotlin"),
		HasCsharp:     util.Includes(cfg.Language, "C#"),
		HasPhp:        util.Includes(cfg.Language, "PHP"),
		HasRuby:       util.Includes(cfg.Language, "Ruby"),
		HasSwift:      util.Includes(cfg.Language, "Swift"),
		HasDart:       util.Includes(cfg.Language, "Dart"),
		HasShell:      util.Includes(cfg.Language, "Shell"),
		HasLua:        util.Includes(cfg.Language, "Lua"),
		HasElixir:     util.Includes(cfg.Language, "Elixir"),
		HasHaskell:    util.Includes(cfg.Language, "Haskell"),
		HasElm:        util.Includes(cfg.Language, "Elm"),
		HasJulia:      util.Includes(cfg.Language, "Julia"),
		HasSql:        util.Includes(cfg.Language, "SQL"),
		Date:          time.Now().Format("2006-01-02"),
	}

	var b bytes.Buffer
	if err := tmpl.Execute(&b, data); err != nil {
		panic(err)
	}
	return b.String()
}

// RenderAgent returns the content for a given agent name from assets
func RenderAgent(name string, assetsFS embed.FS) string {
	content, err := assetsFS.ReadFile("assets/agents/" + name + ".md")
	if err != nil {
		return `---
name: ` + name + `
description: Custom subagent
---
Provide a focused role and steps.`
	}
	return string(content)
}
