package composebuilder

import (
	"context"
	_ "embed"
	"os"
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

func TestBuilder(t *testing.T) {
	base, err := loader.LoadWithContext(
		context.Background(),
		types.ConfigDetails{
			ConfigFiles: []types.ConfigFile{{
				Content: baseBin,
			}},
			Environment: types.NewMapping(os.Environ()),
		},
		func(o *loader.Options) {
			o.SetProjectName("yay", true)
		},
	)
	assert.NilError(t, err)

	var p types.Project
	err = Build(
		&p,
		types.ServiceConfig{
			Labels: types.Labels{"baz": "qux"},
		},
		map[string]BuildEntry{
			"foo": {
				Config: base.AllServices()[0],
				Meta: map[string]any{
					"add": "init",
				},
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

	bin, err := p.MarshalYAML()
	assert.NilError(t, err)
	t.Logf("%s", bin)

	cmp.DeepEqual(bin, mutatedBin)
}
