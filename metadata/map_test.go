package metadata_test

import (
	"testing"

	"github.com/castmetal/cliquefarma-analize-redirect-csv/metadata"
	"github.com/stretchr/testify/require"
)

func TestMetadataAsInt(t *testing.T) {
	testCases := []struct {
		desc string

		view     metadata.Map
		validate func(t *testing.T, m metadata.Map)
	}{
		{
			desc: "value not existing, using default",
			view: metadata.Map{},
			validate: func(t *testing.T, m metadata.Map) {
				got := m.AsInt("value", 1)
				require.Equal(t, 1, got)
			},
		},
		{
			desc: "from string, as int",
			view: metadata.Map{"value": "88"},
			validate: func(t *testing.T, m metadata.Map) {
				got := m.AsInt("value", 0)
				require.Equal(t, 88, got)
			},
		},
		{
			desc: "from float, as int",
			view: metadata.Map{"value": 1.0},
			validate: func(t *testing.T, m metadata.Map) {
				got := m.AsInt("value", 0)
				require.Equal(t, 1, got)
			},
		},
		{
			desc: "from float64, as int",
			view: metadata.Map{"value": float64(2.0)},
			validate: func(t *testing.T, m metadata.Map) {
				got := m.AsInt("value", 0)
				require.Equal(t, 2, got)
			},
		},
		{
			desc: "from int64, as int",
			view: metadata.Map{"value": int64(5)},
			validate: func(t *testing.T, m metadata.Map) {
				got := m.AsInt("value", 0)
				require.Equal(t, 5, got)
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			tC.validate(t, tC.view)
		})
	}
}

func TestMetadataAsString(t *testing.T) {
	testCases := []struct {
		desc string

		view     metadata.Map
		validate func(t *testing.T, m metadata.Map)
	}{
		{
			desc: "value not existing, using default",
			view: metadata.Map{},
			validate: func(t *testing.T, m metadata.Map) {
				got := m.AsString("value", "hello, world")
				require.Equal(t, "hello, world", got)
			},
		},
		{
			desc: "value coming as integer",
			view: metadata.Map{"value": 33},
			validate: func(t *testing.T, m metadata.Map) {
				got := m.AsString("value", "30")
				require.Equal(t, "33", got)
			},
		},
		{
			desc: "value coming as int64",
			view: metadata.Map{"value": int64(33)},
			validate: func(t *testing.T, m metadata.Map) {
				got := m.AsString("value", "30")
				require.Equal(t, "33", got)
			},
		},
		{
			desc: "value coming as float64",
			view: metadata.Map{"value": float64(33.0)},
			validate: func(t *testing.T, m metadata.Map) {
				got := m.AsString("value", "30")
				require.Equal(t, "33.000000", got)
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			tC.validate(t, tC.view)
		})
	}
}

func TestMetadataAsMap(t *testing.T) {
	testCases := []struct {
		desc string

		view     metadata.Map
		validate func(t *testing.T, m metadata.Map)
	}{
		{
			desc: "getting child map from metadata",
			view: metadata.Map{"inner": metadata.Map{"hello": "world"}},
			validate: func(t *testing.T, m metadata.Map) {
				got := m.AsMap("inner")
				gotStrFromInner := got.AsString("hello", "")
				require.Equal(t, "world", gotStrFromInner)
			},
		},
		{
			desc: "getting child map[string]interface{} from metadata",
			view: metadata.Map{"inner": map[string]interface{}{"hello": "world"}},
			validate: func(t *testing.T, m metadata.Map) {
				got := m.AsMap("inner")
				gotStrFromInner := got.AsString("hello", "")
				require.Equal(t, "world", gotStrFromInner)
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			tC.validate(t, tC.view)
		})
	}
}

func TestMetadataAsFloat64(t *testing.T) {
	testCases := []struct {
		desc string

		view     metadata.Map
		validate func(t *testing.T, m metadata.Map)
	}{
		{
			desc: "value not existing, using default",
			view: metadata.Map{},
			validate: func(t *testing.T, m metadata.Map) {
				got := m.AsFloat64("value", 1.00)
				require.Equal(t, 1.0, got)
			},
		},
		{
			desc: "value as string, converted to float",
			view: metadata.Map{"value": "3.00"},
			validate: func(t *testing.T, m metadata.Map) {
				got := m.AsFloat64("value", 1.00)
				require.Equal(t, 3.0, got)
			},
		},
		{
			desc: "value as int, converted to float",
			view: metadata.Map{"value": 3},
			validate: func(t *testing.T, m metadata.Map) {
				got := m.AsFloat64("value", 1.00)
				require.Equal(t, 3.0, got)
			},
		},
		{
			desc: "value as int64, converted to float",
			view: metadata.Map{"value": int64(3)},
			validate: func(t *testing.T, m metadata.Map) {
				got := m.AsFloat64("value", 1.00)
				require.Equal(t, 3.0, got)
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			tC.validate(t, tC.view)
		})
	}
}
