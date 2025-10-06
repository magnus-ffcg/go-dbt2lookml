# ADR 0014: Support dbt Semantic Models for Measure Generation

**Status:** Accepted

**Date:** 2025

## Context

dbt 1.6+ introduced the Semantic Layer with semantic models that define:
- **Measures** - aggregations (sum, average, count_distinct, min, max, median, etc.)
- **Dimensions** - categorical and time dimensions
- **Entities** - join keys between models

Currently, users must define measures in two places:
1. In dbt semantic models (for the Semantic Layer)
2. In dbt meta tags (for go-dbt2lookml to generate LookML measures)

This creates duplication and maintenance overhead.

We need to decide:
- Continue requiring manual measure definition in meta tags
- Automatically parse semantic models and generate measures
- Support both approaches
- Make semantic model support optional or required

## Decision

We will add **optional support** for parsing dbt semantic models and automatically generating LookML measures from semantic model measure definitions. This will be opt-in via a `--use-semantic-models` flag or config option.

## Rationale

**Pros:**
- **Eliminates duplication** - Define measures once in semantic models
- **Single source of truth** - Semantic models drive both dbt metrics and LookML
- **Consistency** - Measures are identical in dbt Semantic Layer and Looker
- **Less configuration** - Users don't need to duplicate in meta tags
- **Future-proof** - Aligns with dbt's direction toward Semantic Layer
- **Richer metadata** - Semantic models have more detailed measure definitions

**Cons:**
- **Additional complexity** - Need to parse and handle semantic models
- **dbt version dependency** - Requires dbt 1.6+ for semantic models
- **Breaking change risk** - If made required, breaks existing workflows
- **Limited adoption** - Not all users use semantic models yet

**Alternatives considered:**
- **Require semantic models** - Too disruptive for existing users
- **Ignore semantic models** - Misses opportunity to reduce duplication
- **Parse dimensions too** - Adds complexity, dimensions already handled well
- **Parse all semantic layer features** - Too ambitious, includes metrics

## Consequences

**Positive:**
- Users with semantic models save configuration time
- Measures stay in sync between dbt and LookML
- Encourages adoption of dbt Semantic Layer
- Provides migration path from meta-only approach
- No breaking changes (opt-in)

**Negative:**
- More code to maintain (semantic model parser)
- Need to handle semantic model + meta tag conflicts
- Documentation needs to cover both approaches
- Testing requires semantic model fixtures

## Implementation Details

### Opt-in Approach

Users enable via flag:
```bash
dbt2lookml --use-semantic-models
```

Or config file:
```yaml
use_semantic_models: true
```

Default: `false` (backward compatible)

### Measure Type Mapping

| dbt Semantic Model | LookML Measure |
|-------------------|----------------|
| `agg: sum` | `type: sum` |
| `agg: average` | `type: average` |
| `agg: min` | `type: min` |
| `agg: max` | `type: max` |
| `agg: median` | `type: median` |
| `agg: count_distinct` | `type: count_distinct` |
| `agg: percentile` | Custom SQL with percentile |
| `agg: sum_boolean` | `type: sum` with boolean cast |

### Conflict Resolution

When both semantic model and meta tag define measures with the same name:

**Decision:** Semantic model measures **always** take precedence.

**Rationale:**
- Semantic models are dbt's future direction for metric/measure definition
- Simpler mental model - no configuration needed
- Encourages migration to semantic models
- Aligns with dbt's vision of semantic models as the single source of truth

**Behavior:**
- If a semantic model defines a measure named `total_revenue`
- And a meta tag also defines a measure named `total_revenue`
- The semantic model version is used
- The meta tag version is silently skipped (with debug log)
- Meta measures with unique names are still included

**Example:**
```yaml
# Semantic model (USED)
measures:
  - name: total_revenue
    agg: sum
    expr: amount

# Meta tag (SKIPPED - same name)
meta:
  looker:
    measures:
      - name: total_revenue
        type: sum
        
      - name: custom_kpi  # USED - unique name
        type: number
```

### Parsing Strategy

1. Check if `--use-semantic-models` enabled
2. Parse semantic models from `manifest.json`
3. Match semantic models to dbt models by `ref()`
4. Extract measures from matched semantic models
5. Generate LookML measures
6. Merge with any meta-defined measures (per conflict resolution)

### Example

**dbt semantic model:**
```yaml
semantic_models:
  - name: orders
    model: ref('fact_orders')
    measures:
      - name: total_revenue
        description: "Total order revenue"
        agg: sum
        expr: amount
      
      - name: avg_order_value
        description: "Average order value"
        agg: average
        expr: amount
```

**Generated LookML:**
```lkml
view: fact_orders {
  measure: total_revenue {
    type: sum
    description: "Total order revenue"
    sql: ${amount} ;;
  }

  measure: avg_order_value {
    type: average
    description: "Average order value"
    sql: ${amount} ;;
  }
}
```

## Migration Path

### Phase 1: Opt-in (Current Decision)
- Feature available but disabled by default
- Users explicitly enable via flag/config
- Both approaches supported

### Phase 2: Encourage Adoption (Future)
- Documentation emphasizes semantic models
- Deprecation notice for meta-only approach
- Migration guide provided

### Phase 3: Default On (Future)
- Flip default to `use_semantic_models: true`
- Meta tags still work but deprecated
- Users opt-out if needed

### Phase 4: Semantic Model Preferred (Future)
- Semantic models are primary approach
- Meta tags only for edge cases
- Full feature parity achieved

## Backward Compatibility

**Guaranteed:**
- Existing configs work without changes
- Meta-based measure generation unchanged
- No breaking changes to CLI or config

**When enabled:**
- Semantic models add measures automatically
- Meta tags still work (conflict resolution applies)
- Users control precedence

## Future Enhancements

1. **Parse dimensions from semantic models** - Use semantic dimension metadata
2. **Parse entities for explores** - Better join generation from entities
3. **Support dbt metrics** - Generate derived tables or LookML metrics
4. **Validate semantic models** - Check for LookML compatibility before generation

## Success Criteria

- [ ] Semantic models parsed correctly from manifest.json
- [ ] All measure aggregation types supported
- [ ] Opt-in works (default disabled)
- [ ] Conflict resolution works as specified
- [ ] No breaking changes for existing users
- [ ] Documentation covers both approaches
- [ ] Test coverage >85%

## Notes

This decision balances innovation (supporting dbt's Semantic Layer) with stability (backward compatibility). The opt-in approach allows users to adopt at their own pace while we validate the implementation and gather feedback.

The semantic model approach represents dbt's future direction, so supporting it positions go-dbt2lookml as a forward-thinking tool that integrates well with modern dbt best practices.

**Important Clarification:** If semantic models prove to be comprehensive and reliable for measure definition, the meta-based measure approach (ADR 0004) may become **deprecated**. Semantic models provide:
- More structured metadata
- Better integration with dbt's Semantic Layer
- Single source of truth
- Less duplication

Once semantic model support is validated, we may transition to **semantic models as the primary/only way** to define measures, with meta tags only needed for edge cases or custom LookML-specific configurations that semantic models don't cover (e.g., drill_fields, LookML-specific parameters).

References:
- MetricFlow - https://github.com/dbt-labs/metricflow
- Samples - https://github.com/dbt-labs/metricflow/tree/main/dbt-metricflow/dbt_metricflow/cli/sample_dbt_models/sample_models
- https://docs.getdbt.com/docs/build/about-metricflow