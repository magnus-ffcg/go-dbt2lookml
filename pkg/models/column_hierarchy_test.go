package models

import (
	"testing"

	"github.com/magnus-ffcg/go-dbt2lookml/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestColumnHierarchy_SimpleColumns(t *testing.T) {
	columns := map[string]DbtModelColumn{
		"id":   {Name: "id", DataType: utils.StringPtr("INT64")},
		"name": {Name: "name", DataType: utils.StringPtr("STRING")},
	}

	hierarchy := NewColumnHierarchy(columns)

	assert.NotNil(t, hierarchy.Get("id"))
	assert.NotNil(t, hierarchy.Get("name"))
	assert.False(t, hierarchy.IsArray("id"))
	assert.False(t, hierarchy.HasChildren("id"))
}

func TestColumnHierarchy_NestedStruct(t *testing.T) {
	columns := map[string]DbtModelColumn{
		"id":            {Name: "id", DataType: utils.StringPtr("INT64")},
		"address":       {Name: "address", DataType: utils.StringPtr("STRUCT")},
		"address.city":  {Name: "address.city", DataType: utils.StringPtr("STRING")},
		"address.state": {Name: "address.state", DataType: utils.StringPtr("STRING")},
	}

	hierarchy := NewColumnHierarchy(columns)

	// Check parent
	assert.NotNil(t, hierarchy.Get("address"))
	assert.True(t, hierarchy.HasChildren("address"))
	assert.False(t, hierarchy.IsArray("address"))

	// Check children
	children := hierarchy.GetChildren("address")
	assert.Equal(t, 2, len(children))
	assert.Contains(t, children, "address.city")
	assert.Contains(t, children, "address.state")
}

func TestColumnHierarchy_ArrayColumn(t *testing.T) {
	columns := map[string]DbtModelColumn{
		"id":   {Name: "id", DataType: utils.StringPtr("INT64")},
		"tags": {Name: "tags", DataType: utils.StringPtr("ARRAY<STRING>")},
	}

	hierarchy := NewColumnHierarchy(columns)

	assert.NotNil(t, hierarchy.Get("tags"))
	assert.True(t, hierarchy.IsArray("tags"))
	assert.False(t, hierarchy.HasChildren("tags"))
}

func TestColumnHierarchy_ArrayOfStructs(t *testing.T) {
	columns := map[string]DbtModelColumn{
		"items":       {Name: "items", DataType: utils.StringPtr("ARRAY<STRUCT>")},
		"items.id":    {Name: "items.id", DataType: utils.StringPtr("INT64")},
		"items.name":  {Name: "items.name", DataType: utils.StringPtr("STRING")},
		"items.price": {Name: "items.price", DataType: utils.StringPtr("FLOAT64")},
	}

	hierarchy := NewColumnHierarchy(columns)

	// Check array parent
	assert.True(t, hierarchy.IsArray("items"))
	assert.True(t, hierarchy.HasChildren("items"))

	// Check children
	children := hierarchy.GetChildren("items")
	assert.Equal(t, 3, len(children))
	assert.Contains(t, children, "items.id")
	assert.Contains(t, children, "items.name")
	assert.Contains(t, children, "items.price")

	// Check child columns
	assert.False(t, hierarchy.IsArray("items.id"))
	assert.False(t, hierarchy.HasChildren("items.id"))
}

func TestColumnHierarchy_DeeplyNested(t *testing.T) {
	columns := map[string]DbtModelColumn{
		"customer":                  {Name: "customer", DataType: utils.StringPtr("STRUCT")},
		"customer.orders":           {Name: "customer.orders", DataType: utils.StringPtr("ARRAY<STRUCT>")},
		"customer.orders.items":     {Name: "customer.orders.items", DataType: utils.StringPtr("ARRAY<STRUCT>")},
		"customer.orders.items.sku": {Name: "customer.orders.items.sku", DataType: utils.StringPtr("STRING")},
	}

	hierarchy := NewColumnHierarchy(columns)

	// Check customer (STRUCT)
	assert.False(t, hierarchy.IsArray("customer"))
	assert.True(t, hierarchy.HasChildren("customer"))

	// Check customer.orders (ARRAY)
	assert.True(t, hierarchy.IsArray("customer.orders"))
	assert.True(t, hierarchy.HasChildren("customer.orders"))

	// Check customer.orders.items (nested ARRAY)
	assert.True(t, hierarchy.IsArray("customer.orders.items"))
	assert.True(t, hierarchy.HasChildren("customer.orders.items"))

	// Check deepest child
	assert.False(t, hierarchy.IsArray("customer.orders.items.sku"))
	assert.False(t, hierarchy.HasChildren("customer.orders.items.sku"))
}

func TestColumnHierarchy_NilDataType(t *testing.T) {
	columns := map[string]DbtModelColumn{
		"unknown": {Name: "unknown", DataType: nil},
	}

	hierarchy := NewColumnHierarchy(columns)

	assert.NotNil(t, hierarchy.Get("unknown"))
	assert.False(t, hierarchy.IsArray("unknown"), "nil DataType should not be treated as array")
}

func TestColumnHierarchy_NonExistentColumn(t *testing.T) {
	columns := map[string]DbtModelColumn{
		"id": {Name: "id", DataType: utils.StringPtr("INT64")},
	}

	hierarchy := NewColumnHierarchy(columns)

	assert.Nil(t, hierarchy.Get("nonexistent"))
	assert.False(t, hierarchy.IsArray("nonexistent"))
	assert.False(t, hierarchy.HasChildren("nonexistent"))
	assert.Empty(t, hierarchy.GetChildren("nonexistent"))
}

func TestColumnHierarchy_EmptyColumns(t *testing.T) {
	columns := map[string]DbtModelColumn{}

	hierarchy := NewColumnHierarchy(columns)

	assert.NotNil(t, hierarchy)
	assert.Empty(t, hierarchy.All())
}

func TestColumnHierarchy_All(t *testing.T) {
	columns := map[string]DbtModelColumn{
		"id":        {Name: "id", DataType: utils.StringPtr("INT64")},
		"items":     {Name: "items", DataType: utils.StringPtr("ARRAY<STRUCT>")},
		"items.sku": {Name: "items.sku", DataType: utils.StringPtr("STRING")},
	}

	hierarchy := NewColumnHierarchy(columns)
	all := hierarchy.All()

	// Should have all 3 columns
	require.GreaterOrEqual(t, len(all), 3)
	assert.NotNil(t, all["id"])
	assert.NotNil(t, all["items"])
	assert.NotNil(t, all["items.sku"])
}

func TestColumnHierarchy_ColumnReference(t *testing.T) {
	columns := map[string]DbtModelColumn{
		"id": {Name: "id", DataType: utils.StringPtr("INT64")},
	}

	hierarchy := NewColumnHierarchy(columns)
	info := hierarchy.Get("id")

	require.NotNil(t, info)
	require.NotNil(t, info.Column)
	assert.Equal(t, "id", info.Column.Name)
	assert.Equal(t, "INT64", *info.Column.DataType)
}
