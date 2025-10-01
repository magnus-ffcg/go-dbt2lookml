package models

import (
	"testing"

	"github.com/magnus-ffcg/go-dbt2lookml/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestColumnClassifier_SimpleColumn(t *testing.T) {
	columns := map[string]DbtModelColumn{
		"id": {Name: "id", DataType: utils.StringPtr("INT64")},
	}

	hierarchy := NewColumnHierarchy(columns)
	arrayColumns := map[string]bool{}
	classifier := NewColumnClassifier(hierarchy, arrayColumns)

	category := classifier.Classify("id", columns["id"])
	assert.Equal(t, CategoryMainView, category)
}

func TestColumnClassifier_ArrayColumn(t *testing.T) {
	columns := map[string]DbtModelColumn{
		"tags": {Name: "tags", DataType: utils.StringPtr("ARRAY<STRING>")},
	}

	hierarchy := NewColumnHierarchy(columns)
	arrayColumns := map[string]bool{"tags": true}
	classifier := NewColumnClassifier(hierarchy, arrayColumns)

	category := classifier.Classify("tags", columns["tags"])
	assert.Equal(t, CategoryNestedView, category)
}

func TestColumnClassifier_NestedColumn(t *testing.T) {
	columns := map[string]DbtModelColumn{
		"items":     {Name: "items", DataType: utils.StringPtr("ARRAY<STRUCT>")},
		"items.id":  {Name: "items.id", DataType: utils.StringPtr("INT64")},
		"items.sku": {Name: "items.sku", DataType: utils.StringPtr("STRING")},
	}

	hierarchy := NewColumnHierarchy(columns)
	arrayColumns := map[string]bool{"items": true}
	classifier := NewColumnClassifier(hierarchy, arrayColumns)

	// Array itself
	category := classifier.Classify("items", columns["items"])
	assert.Equal(t, CategoryNestedView, category)

	// Children of array
	category = classifier.Classify("items.id", columns["items.id"])
	assert.Equal(t, CategoryNestedView, category)

	category = classifier.Classify("items.sku", columns["items.sku"])
	assert.Equal(t, CategoryNestedView, category)
}

func TestColumnClassifier_ExcludeStruct(t *testing.T) {
	columns := map[string]DbtModelColumn{
		"address":       {Name: "address", DataType: utils.StringPtr("STRUCT")},
		"address.city":  {Name: "address.city", DataType: utils.StringPtr("STRING")},
		"address.state": {Name: "address.state", DataType: utils.StringPtr("STRING")},
	}

	hierarchy := NewColumnHierarchy(columns)
	arrayColumns := map[string]bool{}
	classifier := NewColumnClassifier(hierarchy, arrayColumns)

	// STRUCT parent should be excluded (has children)
	category := classifier.Classify("address", columns["address"])
	assert.Equal(t, CategoryExcluded, category)

	// Children should be in main view
	category = classifier.Classify("address.city", columns["address.city"])
	assert.Equal(t, CategoryMainView, category)
}

func TestColumnClassifier_ArrayOfStructsNotExcluded(t *testing.T) {
	columns := map[string]DbtModelColumn{
		"items":      {Name: "items", DataType: utils.StringPtr("ARRAY<STRUCT>")},
		"items.id":   {Name: "items.id", DataType: utils.StringPtr("INT64")},
		"items.name": {Name: "items.name", DataType: utils.StringPtr("STRING")},
	}

	hierarchy := NewColumnHierarchy(columns)
	arrayColumns := map[string]bool{"items": true}
	classifier := NewColumnClassifier(hierarchy, arrayColumns)

	// ARRAY<STRUCT> should NOT be excluded, even though it has children
	category := classifier.Classify("items", columns["items"])
	assert.Equal(t, CategoryNestedView, category, "ARRAY<STRUCT> should be in nested view, not excluded")
}

func TestColumnClassifier_GetArrayParent(t *testing.T) {
	columns := map[string]DbtModelColumn{
		"items":     {Name: "items", DataType: utils.StringPtr("ARRAY<STRUCT>")},
		"items.id":  {Name: "items.id", DataType: utils.StringPtr("INT64")},
		"items.sku": {Name: "items.sku", DataType: utils.StringPtr("STRING")},
	}

	hierarchy := NewColumnHierarchy(columns)
	arrayColumns := map[string]bool{"items": true}
	classifier := NewColumnClassifier(hierarchy, arrayColumns)

	// Array itself has no parent
	parent := classifier.GetArrayParent("items")
	assert.Equal(t, "", parent)

	// Children have array parent
	parent = classifier.GetArrayParent("items.id")
	assert.Equal(t, "items", parent)

	parent = classifier.GetArrayParent("items.sku")
	assert.Equal(t, "items", parent)
}

func TestColumnClassifier_MostSpecificArrayParent(t *testing.T) {
	columns := map[string]DbtModelColumn{
		"orders":           {Name: "orders", DataType: utils.StringPtr("ARRAY<STRUCT>")},
		"orders.items":     {Name: "orders.items", DataType: utils.StringPtr("ARRAY<STRUCT>")},
		"orders.items.sku": {Name: "orders.items.sku", DataType: utils.StringPtr("STRING")},
	}

	hierarchy := NewColumnHierarchy(columns)
	arrayColumns := map[string]bool{
		"orders":       true,
		"orders.items": true,
	}
	classifier := NewColumnClassifier(hierarchy, arrayColumns)

	// Should return most specific (longest) parent
	parent := classifier.GetArrayParent("orders.items.sku")
	assert.Equal(t, "orders.items", parent, "Should return most specific array parent")
}

func TestColumnClassifier_HasChildren(t *testing.T) {
	columns := map[string]DbtModelColumn{
		"items":     {Name: "items", DataType: utils.StringPtr("ARRAY<STRUCT>")},
		"items.id":  {Name: "items.id", DataType: utils.StringPtr("INT64")},
		"items.sku": {Name: "items.sku", DataType: utils.StringPtr("STRING")},
	}

	hierarchy := NewColumnHierarchy(columns)
	arrayColumns := map[string]bool{"items": true}
	classifier := NewColumnClassifier(hierarchy, arrayColumns)

	assert.True(t, classifier.HasChildren("items"))
	assert.False(t, classifier.HasChildren("items.id"))
	assert.False(t, classifier.HasChildren("items.sku"))
}

func TestColumnClassifier_IsNestedArray(t *testing.T) {
	columns := map[string]DbtModelColumn{
		"orders":       {Name: "orders", DataType: utils.StringPtr("ARRAY<STRUCT>")},
		"orders.items": {Name: "orders.items", DataType: utils.StringPtr("ARRAY<STRUCT>")},
		"tags":         {Name: "tags", DataType: utils.StringPtr("ARRAY<STRING>")},
	}

	hierarchy := NewColumnHierarchy(columns)
	arrayColumns := map[string]bool{
		"orders":       true,
		"orders.items": true,
		"tags":         true,
	}
	classifier := NewColumnClassifier(hierarchy, arrayColumns)

	// orders.items is nested under orders (both are arrays)
	assert.True(t, classifier.IsNestedArray("orders.items"))

	// orders is top-level array (not nested)
	assert.False(t, classifier.IsNestedArray("orders"))

	// tags is top-level array (not nested)
	assert.False(t, classifier.IsNestedArray("tags"))
}

func TestColumnClassifier_ComplexNesting(t *testing.T) {
	columns := map[string]DbtModelColumn{
		"customer":                  {Name: "customer", DataType: utils.StringPtr("STRUCT")},
		"customer.name":             {Name: "customer.name", DataType: utils.StringPtr("STRING")},
		"customer.orders":           {Name: "customer.orders", DataType: utils.StringPtr("ARRAY<STRUCT>")},
		"customer.orders.id":        {Name: "customer.orders.id", DataType: utils.StringPtr("INT64")},
		"customer.orders.items":     {Name: "customer.orders.items", DataType: utils.StringPtr("ARRAY<STRUCT>")},
		"customer.orders.items.sku": {Name: "customer.orders.items.sku", DataType: utils.StringPtr("STRING")},
	}

	hierarchy := NewColumnHierarchy(columns)
	arrayColumns := map[string]bool{
		"customer.orders":       true,
		"customer.orders.items": true,
	}
	classifier := NewColumnClassifier(hierarchy, arrayColumns)

	// customer STRUCT should be excluded (has children)
	assert.Equal(t, CategoryExcluded, classifier.Classify("customer", columns["customer"]))

	// customer.name is child of STRUCT (not array), goes to main view
	assert.Equal(t, CategoryMainView, classifier.Classify("customer.name", columns["customer.name"]))

	// customer.orders is an array
	assert.Equal(t, CategoryNestedView, classifier.Classify("customer.orders", columns["customer.orders"]))

	// customer.orders.id is child of array
	assert.Equal(t, CategoryNestedView, classifier.Classify("customer.orders.id", columns["customer.orders.id"]))

	// customer.orders.items is nested array
	assert.Equal(t, CategoryNestedView, classifier.Classify("customer.orders.items", columns["customer.orders.items"]))
	assert.True(t, classifier.IsNestedArray("customer.orders.items"))

	// customer.orders.items.sku is child of nested array
	assert.Equal(t, CategoryNestedView, classifier.Classify("customer.orders.items.sku", columns["customer.orders.items.sku"]))
	parent := classifier.GetArrayParent("customer.orders.items.sku")
	assert.Equal(t, "customer.orders.items", parent)
}

func TestColumnClassifier_EmptyArrayColumns(t *testing.T) {
	columns := map[string]DbtModelColumn{
		"id":   {Name: "id", DataType: utils.StringPtr("INT64")},
		"name": {Name: "name", DataType: utils.StringPtr("STRING")},
	}

	hierarchy := NewColumnHierarchy(columns)
	arrayColumns := map[string]bool{}
	classifier := NewColumnClassifier(hierarchy, arrayColumns)

	// All columns should go to main view when no arrays
	assert.Equal(t, CategoryMainView, classifier.Classify("id", columns["id"]))
	assert.Equal(t, CategoryMainView, classifier.Classify("name", columns["name"]))
}
