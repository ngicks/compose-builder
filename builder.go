package composebuilder

import (
	"context"
	"os"
	"sort"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
)

type ServiceMutator func(sc types.ServiceConfig, p *types.Project, meta map[string]any) (types.ServiceConfig, error)

type BuildEntry struct {
	Config types.ServiceConfig
	Meta   map[string]any
}

func Build(
	base *types.Project,
	defaultService types.ServiceConfig,
	buildInfo map[string]BuildEntry,
	mutator []ServiceMutator,
) error {
	for k, v := range buildInfo {
		merged := mergeService(defaultService, v.Config)
		merged.Name = k

		for _, m := range mutator {
			var err error
			merged, err = m(merged, base, v.Meta)
			if err != nil {
				return err
			}
		}

		base.Services = append(base.Services, merged)
	}

	sort.Slice(base.Services, func(i, j int) bool {
		return base.Services[i].Name > base.Services[j].Name
	})

	return nil
}

func mergeService(service types.ServiceConfig, override types.ServiceConfig) types.ServiceConfig {
	service.Name = "tmp"
	override.Name = "tmp"

	p := types.Project{
		Services: []types.ServiceConfig{service},
	}
	o := types.Project{
		Services: []types.ServiceConfig{override},
	}

	bin, _ := p.MarshalYAML()
	left, _ := loader.ParseYAML(bin)
	binOverride, _ := o.MarshalYAML()
	right, _ := loader.ParseYAML(binOverride)

	project, err := loader.LoadWithContext(
		context.Background(),
		types.ConfigDetails{
			ConfigFiles: []types.ConfigFile{
				{Config: left},
				{Config: right},
			},
			Environment: types.NewMapping(os.Environ()),
		},
		func(o *loader.Options) {
			o.SetProjectName("temp", true)
			o.SkipValidation = true
			o.SkipInterpolation = true
			o.SkipNormalization = true
			o.ResolvePaths = false
			o.SkipConsistencyCheck = true
			o.SkipExtends = true
			o.SkipResolveEnvironment = true
			o.Profiles = []string{"*"}
		},
	)
	if err != nil {
		panic(err)
	}

	return project.AllServices()[0]
}
