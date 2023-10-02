package composebuilder

import (
	"context"
	_ "embed"
	"sort"
	"testing"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

//go:embed testdata/base.yml
var baseBin []byte

//go:embed testdata/mutated.yml
var mutatedBin []byte

func load(content []byte) *types.Project {
	p, err := loader.LoadWithContext(
		context.Background(),
		types.ConfigDetails{
			ConfigFiles: []types.ConfigFile{
				{Content: content},
			},
			Environment: types.NewMapping(nil),
		},
		func(o *loader.Options) {
			o.SetProjectName("yay", true)
			o.ResolvePaths = false
			o.SkipResolveEnvironment = true
			o.Profiles = []string{"*"}
		},
	)
	if err != nil {
		panic(err)
	}
	return p
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func TestBuilder(t *testing.T) {
	base := load(baseBin)
	mutated := load(mutatedBin)

	var p types.Project = types.Project{
		Name: "yay",
		Networks: types.Networks{
			"default": {Name: "yay_default"},
			"sample_network": {
				Name:     "sample_network",
				External: types.External{External: true},
			},
		},
		Volumes: types.Volumes{
			"sample_volume": {
				Name:     "sample_volume",
				External: types.External{External: true},
			},
		},
		Secrets: types.Secrets{
			"server-certificate": {
				Name:     "server-certificate",
				External: types.External{External: true},
			},
		},
		Configs: types.Configs{
			"http_config": {
				Name: "yay_http_config",
				File: "./httpd.conf",
			},
		},
		Environment: types.Mapping{"COMPOSE_PROJECT_NAME": "yay"},
	}

	err := Build(
		&p,
		types.ServiceConfig{
			Labels: types.Labels{"baz": "qux"},
		},
		map[string]BuildEntry{
			"foo": {
				Config: must(base.GetService("foo")),
				Meta: map[string]any{
					"add": "init",
				},
			},
			"baz": {
				Config: must(base.GetService("baz")),
			},
		},
		[]ServiceMutator{
			func(sc types.ServiceConfig, p *types.Project, meta map[string]any) (types.ServiceConfig, error) {
				cmp.DeepEqual(map[string]any{"add": "init"}, meta)
				if _, ok := meta["add"]; ok {
					t := true
					sc.Init = &t
				}
				return sc, nil
			},
		},
	)
	assert.NilError(t, err)

	sort.Slice(mutated.Services, func(i, j int) bool {
		return mutated.Services[i].Name > mutated.Services[j].Name
	})

	comparison := cmp.DeepEqual(&p, mutated)
	if cmp := comparison(); !cmp.Success() {
		t.Errorf("cmp: %v\n", cmp)
	}
}
