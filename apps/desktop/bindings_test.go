package main

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

type bindingsBase struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type bindingsSample struct {
	bindingsBase
	Name     string         `json:"title"`
	Optional *int           `json:"optional,omitempty"`
	Nullable *int           `json:"nullable"`
	Created  time.Time      `json:"created"`
	Tags     []string       `json:"tags"`
	Counts   map[string]int `json:"counts"`
	Raw      []byte         `json:"raw"`
	Hidden   string         `json:"-"`
	Untagged bool
	Children []bindingsSample  `json:"children"`
	Labels   map[string]string `json:"labels,omitempty"`
	internal string            //nolint:unused // checks that unexported fields are left out
}

type bindingsService struct{}

func (bindingsService) Load(_ string, _ int) (*bindingsSample, error) { return nil, nil }
func (bindingsService) Save(_ bindingsSample) error                   { return nil }

func TestInterfaceFollowsEncodingJSON(t *testing.T) {
	generator := newTypeScriptGenerator()
	got := generator.interfaceOf(reflect.TypeFor[bindingsSample]())
	want := strings.Join([]string{
		"\texport interface bindingsSample {",
		"\t\tid: string;",
		"\t\tname: string;",
		"\t\ttitle: string;",
		"\t\toptional?: number;",
		"\t\tnullable: number | null;",
		"\t\tcreated: string;",
		"\t\ttags: string[];",
		"\t\tcounts: Record<string, number>;",
		"\t\traw: string;",
		"\t\tUntagged: boolean;",
		"\t\tchildren: main.bindingsSample[];",
		"\t\tlabels?: Record<string, string>;",
		"\t}",
		"",
	}, "\n")
	if got != want {
		t.Errorf("interface:\n%s\nwant:\n%s", got, want)
	}
}

func TestClientFunctions(t *testing.T) {
	generator := newTypeScriptGenerator()
	names := map[string][]string{"Load": {"id", "limit"}}
	got := generator.client(reflect.TypeFor[bindingsService](), names)

	wantLines := []string{
		"import type { main } from '#lib/backend/generated/models.ts';",
		"export function Load(id: string, limit: number): Promise<main.bindingsSample | null> {",
		"\treturn call('Load', [id, limit]);",
		// Without source names, parameters are numbered.
		"export function Save(argument1: main.bindingsSample): Promise<void> {",
	}
	for _, line := range wantLines {
		if !strings.Contains(got, line) {
			t.Errorf("client is missing %q:\n%s", line, got)
		}
	}
}
