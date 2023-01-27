package model

import (
	"context"
	"testing"

	"github.com/evergreen-ci/evergreen/db"
	"github.com/evergreen-ci/evergreen/mock"
	"github.com/evergreen-ci/evergreen/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindExpansionsForVariant(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	env := &mock.Environment{}
	require.NoError(t, env.Configure(ctx))

	assert.NoError(t, db.ClearCollections(ParserProjectCollection))
	pp := ParserProject{
		Id: "v1",
		Axes: []matrixAxis{
			{
				Id: "version",
				Values: []axisValue{
					{
						Id:        "latest",
						Variables: util.Expansions{"VERSION": "latest"},
					},
				},
			},
			{
				Id: "os",
				Values: []axisValue{
					{
						Id:        "windows-64",
						Variables: util.Expansions{"OS": "windows-64"},
					},
				},
			},
		},
		BuildVariants: []parserBV{
			{
				Name:       "myBV",
				Expansions: util.Expansions{"hello": "world", "goodbye": "mars"},
			},
			{
				Name:       "yourBV",
				Expansions: util.Expansions{"milky": "way"},
			},
			{
				Matrix: &matrix{
					Id: "test",
					Spec: matrixDefinition{
						"os": parserStringSlice{
							"*",
						},
						"version": parserStringSlice{
							"*",
						},
					},
				},
			},
		},
	}

	v := &Version{Id: "v1"}
	ppStorage, err := GetParserProjectStorage(env.Settings(), ProjectStorageMethodDB)
	require.NoError(t, err)
	defer ppStorage.Close(ctx)
	assert.NoError(t, ppStorage.UpsertOne(ctx, &pp))
	expansions, err := FindExpansionsForVariant(ctx, env.Settings(), v, "myBV")
	assert.NoError(t, err)
	assert.Equal(t, expansions["hello"], "world")
	assert.Equal(t, expansions["goodbye"], "mars")
	assert.Empty(t, expansions["milky"])
	expansions, err = FindExpansionsForVariant(ctx, env.Settings(), v, "test__version~latest_os~windows-64")
	assert.NoError(t, err)
	assert.Equal(t, expansions["VERSION"], "latest")
	assert.Equal(t, expansions["OS"], "windows-64")
}
