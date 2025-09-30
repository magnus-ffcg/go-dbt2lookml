# Un-hide and use this explore, or copy the joins into another explore, to get all the fully nested relationships from this view
explore: d_item_v3 {
  hidden: yes
    join: d_item_v3__packaging_information__packaging_material_composition {
      view_label: "D Item V3: Packaging Information Packaging Material Composition"
      sql: LEFT JOIN UNNEST(${d_item_v3.packaging_information__packaging_material_composition}) as d_item_v3__packaging_information__packaging_material_composition ;;
      relationship: one_to_many
    }
    join: d_item_v3__ica_environmental_accreditation {
      view_label: "D Item V3: Ica Environmental Accreditation"
      sql: LEFT JOIN UNNEST(${d_item_v3.ica_environmental_accreditation}) as d_item_v3__ica_environmental_accreditation ;;
      relationship: one_to_many
    }
    join: d_item_v3__ica_ethical_accreditation {
      view_label: "D Item V3: Ica Ethical Accreditation"
      sql: LEFT JOIN UNNEST(${d_item_v3.ica_ethical_accreditation}) as d_item_v3__ica_ethical_accreditation ;;
      relationship: one_to_many
    }
    join: d_item_v3__item_information_claim_detail {
      view_label: "D Item V3: Item Information Claim Detail"
      sql: LEFT JOIN UNNEST(${d_item_v3.item_information_claim_detail}) as d_item_v3__item_information_claim_detail ;;
      relationship: one_to_many
    }
    join: d_item_v3__ica_ecological_accreditation {
      view_label: "D Item V3: Ica Ecological Accreditation"
      sql: LEFT JOIN UNNEST(${d_item_v3.ica_ecological_accreditation}) as d_item_v3__ica_ecological_accreditation ;;
      relationship: one_to_many
    }
    join: d_item_v3__ica_non_ecological_accreditation {
      view_label: "D Item V3: Ica Non Ecological Accreditation"
      sql: LEFT JOIN UNNEST(${d_item_v3.ica_non_ecological_accreditation}) as d_item_v3__ica_non_ecological_accreditation ;;
      relationship: one_to_many
    }
    join: d_item_v3__packaging_information__packaging_material_composition__packaging_material_composition_quantity {
      view_label: "D Item V3: Packaging Information Packaging Material Composition Packaging Material Composition Quantity"
      sql: LEFT JOIN UNNEST(${d_item_v3.packaging_information__packaging_material_composition__packaging_material_composition_quantity}) as d_item_v3__packaging_information__packaging_material_composition__packaging_material_composition_quantity ;;
      relationship: one_to_many
    }
    join: d_item_v3__central_department {
      view_label: "D Item V3: Central Department"
      sql: LEFT JOIN UNNEST(${d_item_v3.central_department}) as d_item_v3__central_department ;;
      relationship: one_to_many
    }
    join: d_item_v3__country_of_origin {
      view_label: "D Item V3: Country Of Origin"
      sql: LEFT JOIN UNNEST(${d_item_v3.country_of_origin}) as d_item_v3__country_of_origin ;;
      relationship: one_to_many
    }
    join: d_item_v3__load_carrier_deposit {
      view_label: "D Item V3: Load Carrier Deposit"
      sql: LEFT JOIN UNNEST(${d_item_v3.load_carrier_deposit}) as d_item_v3__load_carrier_deposit ;;
      relationship: one_to_many
    }
    join: d_item_v3__ica_swedish_accreditation {
      view_label: "D Item V3: Ica Swedish Accreditation"
      sql: LEFT JOIN UNNEST(${d_item_v3.ica_swedish_accreditation}) as d_item_v3__ica_swedish_accreditation ;;
      relationship: one_to_many
    }
    join: d_item_v3__accreditation {
      view_label: "D Item V3: Accreditation"
      sql: LEFT JOIN UNNEST(${d_item_v3.accreditation}) as d_item_v3__accreditation ;;
      relationship: one_to_many
    }
}
view: d_item_v3 {
  sql_table_name: `ac16-p-conlaybi-prd-4257.item_versioned.d_item_v3` ;;
  description: "Dimension for items identified by global trade item number (SCD1)"
  dimension: primary_soi_supplier_reference__supplier_id {
    type: string
    sql: ${TABLE}.primary_soi_supplier_reference.supplier_id ;;
    group_label: "Primary_soi_supplier_reference"
    group_item_label: "Supplier_id"
    description: "Supplier number in the Fusion Cloud application"
  }
  dimension: season__code_name {
    type: string
    sql: ${TABLE}.season.code_name ;;
    group_label: "Season"
    group_item_label: "Code_name"
  }
  dimension: brand__code_value {
    type: string
    sql: ${TABLE}.brand.code_value ;;
    group_label: "Brand"
    group_item_label: "Code_value"
  }
  dimension: assortment_attributes__swedish__code_value {
    type: string
    sql: ${TABLE}.assortment_attributes.swedish.code_value ;;
    group_label: "Assortment_attributes Swedish"
    group_item_label: "Code_value"
  }
  dimension: lifecycle__ica_discontinue_reason {
    type: string
    sql: ${TABLE}.lifecycle.ica_discontinue_reason ;;
    group_label: "Lifecycle"
    group_item_label: "Ica_discontinue_reason"
    description: "Reason for discontinueing the Item, either ICA or supplier. IF ItemEBO/PackStructure/Item/TradeItem/TradeItemSynchronisationDates/DiscontinuedDateTime is null THEN ItemEBO/PackStructure/Item/ItemStatuses/ICADiscontinueReason = 'ICA' ELSE 'SUPPLIER'"
  }
  dimension: primary_soi_supplier_reference__is_primary_consumer_item_for_soi {
    type: yesno
    sql: ${TABLE}.primary_soi_supplier_reference.is_primary_consumer_item_for_soi ;;
    group_label: "Primary_soi_supplier_reference"
    group_item_label: "Is_primary_consumer_item_for_soi"
    description: "Primary supplier and item used for purchasing"
  }
  dimension: ecr_category__code_name {
    type: string
    sql: ${TABLE}.ecr_category.code_name ;;
    group_label: "Ecr_category"
    group_item_label: "Code_name"
  }
  dimension: lifecycle__introduction_status {
    type: string
    sql: ${TABLE}.lifecycle.introduction_status ;;
    group_label: "Lifecycle"
    group_item_label: "Introduction_status"
    description: "Mapping to the ICA End to End Process Status is for information purposes only. Only those for ItemEBO/PackStructure/Item/ItemStatuses/Status = 'DRAFT' will be modelled as a secondary status in FPH as an attribute called 'Item Introduction Status'"
  }
  dimension: primary_soi_supplier_reference__supplychain_supplier_long_name {
    type: string
    sql: ${TABLE}.primary_soi_supplier_reference.supplychain_supplier_long_name ;;
    group_label: "Primary_soi_supplier_reference"
    group_item_label: "Supplychain_supplier_long_name"
    description: "Long name of supplychain supplier"
  }
  dimension: alcohol_percentage_by_volume {
    type: number
    sql: ${TABLE}.alcohol_percentage_by_volume ;;
    description: "(T2208) Percentage of alcohol contained in the base unit trade item"
  }
  dimension: category_specific_attributes__origin__code_name {
    type: string
    sql: ${TABLE}.category_specific_attributes.origin.code_name ;;
    group_label: "Category_specific_attributes Origin"
    group_item_label: "Code_name"
  }
  dimension: category_specific_attributes__raw_material__code_value {
    type: string
    sql: ${TABLE}.category_specific_attributes.raw_material.code_value ;;
    group_label: "Category_specific_attributes Raw_material"
    group_item_label: "Code_value"
  }
  dimension: information_providing_supplier {
    type: string
    sql: ${TABLE}.information_providing_supplier ;;
    description: "(Record)  Supplier that has been associated with the Item NOTE! It is the Information provider that will be used for the item information in FPH"
  }
  dimension: assortment_attributes__multicultural__code_name {
    type: string
    sql: ${TABLE}.assortment_attributes.multicultural.code_name ;;
    group_label: "Assortment_attributes Multicultural"
    group_item_label: "Code_name"
  }
  dimension: category_specific_attributes__flavour__code_value {
    type: string
    sql: ${TABLE}.category_specific_attributes.flavour.code_value ;;
    group_label: "Category_specific_attributes Flavour"
    group_item_label: "Code_value"
  }
  dimension: is_private_label {
    type: yesno
    sql: ${TABLE}.is_private_label ;;
    description: "Attribute indicating if the item is a Private Label, that is an ICA branded product (aka EMV)."
  }
  dimension: division_name {
    type: string
    sql: ${TABLE}.division_name ;;
    description: "Merchandise hierarchy node category name; e.g Asiatiska köket"
  }
  dimension: category_specific_attributes__execution1__code_value {
    type: string
    sql: ${TABLE}.category_specific_attributes.execution1.code_value ;;
    group_label: "Category_specific_attributes Execution1"
    group_item_label: "Code_value"
  }
  dimension: net_weight {
    type: number
    sql: ${TABLE}.net_weight ;;
    description: "The net weight in GRAM of the trade item. Autocalculated from GS1 attributes; 'Gross Weight' - 'Packaging weight'."
  }
  dimension: packaging_information__packaging_weight {
    type: number
    sql: ${TABLE}.packaging_information.packaging_weight ;;
    group_label: "Packaging_information"
    group_item_label: "Packaging_weight"
    description: "Used to identify the measurement of the packaging weight of the trade item."
  }
  dimension: category_specific_attributes__execution1__code_description {
    type: string
    sql: ${TABLE}.category_specific_attributes.execution1.code_description ;;
    group_label: "Category_specific_attributes Execution1"
    group_item_label: "Code_description"
  }
  dimension: d_item_key {
    type: number
    sql: ${TABLE}.d_item_key ;;
    description: "Technical key for d_item, derived from GTIN"
  }
  dimension: assortment_attributes__swedish__code_description {
    type: string
    sql: ${TABLE}.assortment_attributes.swedish.code_description ;;
    group_label: "Assortment_attributes Swedish"
    group_item_label: "Code_description"
  }
  dimension: measurements__depth_unit_of_measure {
    type: string
    sql: ${TABLE}.measurements.depth_unit_of_measure ;;
    group_label: "Measurements"
    group_item_label: "Depth_unit_of_measure"
    description: "(T3780) unit of measure value associated to depth value"
  }
  dimension: measurements__net_content_per_piece {
    type: number
    sql: ${TABLE}.measurements.net_content_per_piece ;;
    group_label: "Measurements"
    group_item_label: "Net_content_per_piece"
    description: "(T0082) The amount of the trade item contained by a package, usually as claimed on the label. For example, Water 750ml - net content = 750 MLT ; 20 count pack of diapers, net content = 20 ea.. In case of multi-pack, indicates the net content of the total trade item. For fixed value trade items use the value claimed on the package, to avoid variable fill rate issue that arises with some trade item which are sold by volume or weight, and whose actual content may vary slightly from batch to batch. In case of variable quantity trade items, indicates the average quantity. Allows for the representation of the same value in different units of measure but not multiple values. Only values having UOM = PIECE"
  }
  dimension: returnable_asset_deposit_name {
    type: string
    sql: ${TABLE}.returnable_asset_deposit_name ;;
    description: "(T0148) Depositname e.g. Engångs Pet över 1000 ml"
  }
  dimension: category_specific_attributes__specific_content__code_value {
    type: string
    sql: ${TABLE}.category_specific_attributes.specific_content.code_value ;;
    group_label: "Category_specific_attributes Specific_content"
    group_item_label: "Code_value"
  }
  dimension: is_despatch_unit {
    type: yesno
    sql: ${TABLE}.is_despatch_unit ;;
    description: "(T4038) An indicator identifying that the information providerconsiders the trade item as a despatch (shipping) unit. Thismay be relationship dependent based on channel of tradeor other point to point agreement."
  }
  dimension: price_comparison {
    type: number
    sql: ${TABLE}.price_comparison ;;
    description: "The quantity of the product at usage. Applicable for concentrated products and products where the comparison price is calculated based on a measurement other than netContent. This field is dependent on the population of priceComparisonContentType and is required when priceComparisonContentType is used. Allows for the representation of the same value in different units of measure but not multiple values."
  }
  dimension: catchweight_type_cd {
    type: string
    sql: ${TABLE}.catchweight_type_cd ;;
    description: "(Record) Possibillity to flag items as solid weight even if the GS1 information says it's not. It can both be items with variable weight or not. 'Solid' or 'Exact"
  }
  dimension: segment_name {
    type: string
    sql: ${TABLE}.segment_name ;;
    description: "Merchandise hierarchy node category name; e.g Asiatiska köket"
  }
  dimension: vat_percent {
    type: number
    sql: ${TABLE}.vat_percent ;;
    description: "(T0195) The current tax or duty rate percentage applicable to the trade item."
  }
  dimension: category_specific_attributes__colour__code_description {
    type: string
    sql: ${TABLE}.category_specific_attributes.colour.code_description ;;
    group_label: "Category_specific_attributes Colour"
    group_item_label: "Code_description"
  }
  dimension: sub_category_description {
    type: string
    sql: ${TABLE}.sub_category_description ;;
    description: "Merchandise hierarchy node category description; concatenation of id and name; e.g 7101 - Asiatiska köket"
  }
  dimension: assortment_attributes__pack_variant__code_description {
    type: string
    sql: ${TABLE}.assortment_attributes.pack_variant.code_description ;;
    group_label: "Assortment_attributes Pack_variant"
    group_item_label: "Code_description"
  }
  dimension: bica_calculated_fields__bica_improved_weight_volume_uom {
    type: string
    sql: ${TABLE}.bica_calculated_fields.bica_improved_weight_volume_uom ;;
    group_label: "Bica_calculated_fields"
    group_item_label: "Bica_improved_weight_volume_uom"
    description: "Calculated field - Unit of measure value associated to bica_improved_weight_volume"
  }
  dimension: assortment_attributes__gdpr_sensitive__code_description {
    type: string
    sql: ${TABLE}.assortment_attributes.gdpr_sensitive.code_description ;;
    group_label: "Assortment_attributes Gdpr_sensitive"
    group_item_label: "Code_description"
  }
  dimension: measurements__net_content_others_unit_of_measure {
    type: string
    sql: ${TABLE}.measurements.net_content_others_unit_of_measure ;;
    group_label: "Measurements"
    group_item_label: "Net_content_others_unit_of_measure"
    description: "(T3780) unit of measure value associated to net content others value"
  }
  dimension: assortment_attributes__quality__code_description {
    type: string
    sql: ${TABLE}.assortment_attributes.quality.code_description ;;
    group_label: "Assortment_attributes Quality"
    group_item_label: "Code_description"
  }
  dimension: category_specific_attributes__preparation__code_name {
    type: string
    sql: ${TABLE}.category_specific_attributes.preparation.code_name ;;
    group_label: "Category_specific_attributes Preparation"
    group_item_label: "Code_name"
  }
  dimension: assortment_attributes__price_range__code_value {
    type: string
    sql: ${TABLE}.assortment_attributes.price_range.code_value ;;
    group_label: "Assortment_attributes Price_range"
    group_item_label: "Code_value"
  }
  dimension: returnable_asset_deposit_type {
    type: string
    sql: ${TABLE}.returnable_asset_deposit_type ;;
    description: "(T0148) Type of deposit item (Container,Crate,LoadCarrier)"
  }
  dimension: gpc_category_code {
    type: string
    sql: ${TABLE}.gpc_category_code ;;
    description: "(T0280) Code specifying a product category according to the GS1 Global Product Classification (GPC) standard."
  }
  dimension: core_input_reason__code_value {
    type: string
    sql: ${TABLE}.core_input_reason.code_value ;;
    group_label: "Core_input_reason"
    group_item_label: "Code_value"
  }
  dimension: measurements__net_content_in_millilitre {
    type: number
    sql: ${TABLE}.measurements.net_content_in_millilitre ;;
    group_label: "Measurements"
    group_item_label: "Net_content_in_millilitre"
    description: "(T0082) The amount of the trade item contained by a package, usually as claimed on the label. For example, Water 750ml - net content = 750 MLT ; 20 count pack of diapers, net content = 20 ea.. In case of multi-pack, indicates the net content of the total trade item. For fixed value trade items use the value claimed on the package, to avoid variable fill rate issue that arises with some trade item which are sold by volume or weight, and whose actual content may vary slightly from batch to batch. In case of variable quantity trade items, indicates the average quantity. Allows for the representation of the same value in different units of measure but not multiple values. Only values having UOM = MILLIITRE"
  }
  dimension: bica_calculated_fields__bica_improved_ecological_markup {
    type: string
    sql: ${TABLE}.bica_calculated_fields.bica_improved_ecological_markup ;;
    group_label: "Bica_calculated_fields"
    group_item_label: "Bica_improved_ecological_markup"
    description: "Using ecological_markup and additional information from item_description to retrive if it's an Eco / Krav product"
  }
  dimension: is_scale_plu {
    type: yesno
    sql: ${TABLE}.is_scale_plu ;;
    description: "If an item is a scale-PLU item"
  }
  dimension: category_specific_attributes__preparation__code_description {
    type: string
    sql: ${TABLE}.category_specific_attributes.preparation.code_description ;;
    group_label: "Category_specific_attributes Preparation"
    group_item_label: "Code_description"
  }
  dimension: is_ica_external_sourcing {
    type: yesno
    sql: ${TABLE}.is_ica_external_sourcing ;;
    description: "Attribute that indicates wether the specified item is a Central or External item. True = 'External' False = 'Central'"
  }
  dimension: is_base_unit {
    type: yesno
    sql: ${TABLE}.is_base_unit ;;
    description: "(T4012) An indicator identifying the trade item as the base unit level of the trade item hierarchy."
  }
  dimension: category_specific_attributes__execution3__code_description {
    type: string
    sql: ${TABLE}.category_specific_attributes.execution3.code_description ;;
    group_label: "Category_specific_attributes Execution3"
    group_item_label: "Code_description"
  }
  dimension: lifecycle__novelty_type {
    type: string
    sql: ${TABLE}.lifecycle.novelty_type ;;
    group_label: "Lifecycle"
    group_item_label: "Novelty_type"
    description: "Type of novelty; e-g- New , Changed"
  }
  dimension: category_specific_attributes__specific_content__code_description {
    type: string
    sql: ${TABLE}.category_specific_attributes.specific_content.code_description ;;
    group_label: "Category_specific_attributes Specific_content"
    group_item_label: "Code_description"
  }
  dimension: category_specific_attributes__execution4__code_value {
    type: string
    sql: ${TABLE}.category_specific_attributes.execution4.code_value ;;
    group_label: "Category_specific_attributes Execution4"
    group_item_label: "Code_value"
  }
  dimension: season__code_description {
    type: string
    sql: ${TABLE}.season.code_description ;;
    group_label: "Season"
    group_item_label: "Code_description"
  }
  dimension: assortment_attributes__gdpr_sensitive__code_name {
    type: string
    sql: ${TABLE}.assortment_attributes.gdpr_sensitive.code_name ;;
    group_label: "Assortment_attributes Gdpr_sensitive"
    group_item_label: "Code_name"
  }
  dimension: assortment_attributes__multicultural__code_description {
    type: string
    sql: ${TABLE}.assortment_attributes.multicultural.code_description ;;
    group_label: "Assortment_attributes Multicultural"
    group_item_label: "Code_description"
  }
  dimension: category_specific_attributes__execution4__code_description {
    type: string
    sql: ${TABLE}.category_specific_attributes.execution4.code_description ;;
    group_label: "Category_specific_attributes Execution4"
    group_item_label: "Code_description"
  }
  dimension: primary_soi_supplier_reference__supplier_site_description {
    type: string
    sql: ${TABLE}.primary_soi_supplier_reference.supplier_site_description ;;
    group_label: "Primary_soi_supplier_reference"
    group_item_label: "Supplier_site_description"
    description: "Description of supplier site"
  }
  dimension: is_bonus_item {
    type: yesno
    sql: ${TABLE}.is_bonus_item ;;
    description: "If the item will give ICA bonus to end customer or not"
  }
  dimension: assortment_attributes__ethical {
    type: string
    sql: ${TABLE}.assortment_attributes.ethical ;;
    group_label: "Assortment_attributes"
    group_item_label: "Ethical"
    description: "Indicates if item has any markings that is considerad as Ethical; Etisk märkning / Saknar etisk märkning"
  }
  dimension: measurements__net_content_others {
    type: number
    sql: ${TABLE}.measurements.net_content_others ;;
    group_label: "Measurements"
    group_item_label: "Net_content_others"
    description: "(T0082) The amount of the trade item contained by a package, usually as claimed on the label. For example, Water 750ml - net content = 750 MLT ; 20 count pack of diapers, net content = 20 ea.. In case of multi-pack, indicates the net content of the total trade item. For fixed value trade items use the value claimed on the package, to avoid variable fill rate issue that arises with some trade item which are sold by volume or weight, and whose actual content may vary slightly from batch to batch. In case of variable quantity trade items, indicates the average quantity. Allows for the representation of the same value in different units of measure but not multiple values. Only values having UOM not matching any of the other"
  }
  dimension: measurements__net_content_in_gram {
    type: number
    sql: ${TABLE}.measurements.net_content_in_gram ;;
    group_label: "Measurements"
    group_item_label: "Net_content_in_gram"
    description: "(T0082) The amount of the trade item contained by a package, usually as claimed on the label. For example, Water 750ml - net content = 750 MLT ; 20 count pack of diapers, net content = 20 ea.. In case of multi-pack, indicates the net content of the total trade item. For fixed value trade items use the value claimed on the package, to avoid variable fill rate issue that arises with some trade item which are sold by volume or weight, and whose actual content may vary slightly from batch to batch. In case of variable quantity trade items, indicates the average quantity. Allows for the representation of the same value in different units of measure but not multiple values. Only values having UOM = GRAM"
  }
  dimension: measurements__width {
    type: number
    sql: ${TABLE}.measurements.width ;;
    group_label: "Measurements"
    group_item_label: "Width"
    description: "(T4017) The width of the unit load, as measured according to the GS1 Package Measurement Rules, including the shipping platform unless it is excluded according to the Pallet Type Code chosen."
  }
  dimension: segment_description {
    type: string
    sql: ${TABLE}.segment_description ;;
    description: "Merchandise hierarchy node category description; concatenation of id and name; e.g 7101 - Asiatiska köket"
  }
  dimension: assortment_attributes__health {
    type: string
    sql: ${TABLE}.assortment_attributes.health ;;
    group_label: "Assortment_attributes"
    group_item_label: "Health"
    description: "Indicates if item is "healthy" or not; Yes/No"
  }
  dimension: category_specific_attributes__product_group__code_name {
    type: string
    sql: ${TABLE}.category_specific_attributes.product_group.code_name ;;
    group_label: "Category_specific_attributes Product_group"
    group_item_label: "Code_name"
  }
  dimension: measurements__net_content_in_millimeter {
    type: number
    sql: ${TABLE}.measurements.net_content_in_millimeter ;;
    group_label: "Measurements"
    group_item_label: "Net_content_in_millimeter"
    description: "(T0082) The amount of the trade item contained by a package, usually as claimed on the label. For example, Water 750ml - net content = 750 MLT ; 20 count pack of diapers, net content = 20 ea.. In case of multi-pack, indicates the net content of the total trade item. For fixed value trade items use the value claimed on the package, to avoid variable fill rate issue that arises with some trade item which are sold by volume or weight, and whose actual content may vary slightly from batch to batch. In case of variable quantity trade items, indicates the average quantity. Allows for the representation of the same value in different units of measure but not multiple values. Only values having UOM = MILLIMETER"
  }
  dimension: category_specific_attributes__origin__code_value {
    type: string
    sql: ${TABLE}.category_specific_attributes.origin.code_value ;;
    group_label: "Category_specific_attributes Origin"
    group_item_label: "Code_value"
  }
  dimension: assortment_attributes__plantbased__code_description {
    type: string
    sql: ${TABLE}.assortment_attributes.plantbased.code_description ;;
    group_label: "Assortment_attributes Plantbased"
    group_item_label: "Code_description"
  }
  dimension: assortment_attributes__ecological {
    type: string
    sql: ${TABLE}.assortment_attributes.ecological ;;
    group_label: "Assortment_attributes"
    group_item_label: "Ecological"
    description: "Indicates if item has any markings that is considerad as ecological/organic; Ekologisk märkning / Saknar Ekologisk märkning"
  }
  dimension: assortment_attributes__price_range__code_name {
    type: string
    sql: ${TABLE}.assortment_attributes.price_range.code_name ;;
    group_label: "Assortment_attributes Price_range"
    group_item_label: "Code_name"
  }
  dimension: consumer_item_reference__plu_number {
    type: string
    sql: ${TABLE}.consumer_item_reference.plu_number ;;
    group_label: "Consumer_item_reference"
    group_item_label: "Plu_number"
    description: "PriceLookUp; an ICA internal number that the cashier in store can use for alternative way of sales"
  }
  dimension: item_description {
    type: string
    sql: ${TABLE}.item_description ;;
    description: "The ItemDescription consists of the combination of the GS1 attributes Brand Name, Item Name and Article Size"
  }
  dimension: ecr_category__code_value {
    type: string
    sql: ${TABLE}.ecr_category.code_value ;;
    group_label: "Ecr_category"
    group_item_label: "Code_value"
  }
  dimension: primary_soi_supplier_reference__soi_status {
    type: string
    sql: ${TABLE}.primary_soi_supplier_reference.soi_status ;;
    group_label: "Primary_soi_supplier_reference"
    group_item_label: "Soi_status"
    description: "A current status for a SOI"
  }
  dimension: category_specific_attributes__product_group__code_value {
    type: string
    sql: ${TABLE}.category_specific_attributes.product_group.code_value ;;
    group_label: "Category_specific_attributes Product_group"
    group_item_label: "Code_value"
  }
  dimension: category_specific_attributes__execution2__code_value {
    type: string
    sql: ${TABLE}.category_specific_attributes.execution2.code_value ;;
    group_label: "Category_specific_attributes Execution2"
    group_item_label: "Code_value"
  }
  dimension: descriptive_size {
    type: string
    sql: ${TABLE}.descriptive_size ;;
    description: "Descriptive size information."
  }
  dimension: supply_chain_orderable_status {
    type: string
    sql: ${TABLE}.supply_chain_orderable_status ;;
    description: "Used to know if ICA will order on this item record's level. Also used to keep track of Add Item process (will be ticked when agreement is set in BasICA)."
  }
  dimension: category_specific_attributes__consumer_group__code_name {
    type: string
    sql: ${TABLE}.category_specific_attributes.consumer_group.code_name ;;
    group_label: "Category_specific_attributes Consumer_group"
    group_item_label: "Code_name"
  }
  dimension: assortment_attributes__packing_size__code_description {
    type: string
    sql: ${TABLE}.assortment_attributes.packing_size.code_description ;;
    group_label: "Assortment_attributes Packing_size"
    group_item_label: "Code_description"
  }
  dimension: measurements__width_unit_of_measure {
    type: string
    sql: ${TABLE}.measurements.width_unit_of_measure ;;
    group_label: "Measurements"
    group_item_label: "Width_unit_of_measure"
    description: "(T3780) unit of measure value associated to width value"
  }
  dimension: measurements__depth {
    type: number
    sql: ${TABLE}.measurements.depth ;;
    group_label: "Measurements"
    group_item_label: "Depth"
    description: "(T4018) The depth of the unit load, as measured according to the GS1 Package Measurement Rules, including the shipping platform unless it is excluded according to the Pallet Type Code chosen."
  }
  dimension: primary_soi_supplier_reference__supplier_site_id {
    type: string
    sql: ${TABLE}.primary_soi_supplier_reference.supplier_site_id ;;
    group_label: "Primary_soi_supplier_reference"
    group_item_label: "Supplier_site_id"
    description: "This is the end-user facing, unique Supplier Site number in the Fusion Cloud application"
  }
  dimension: assortment_attributes__quality__code_name {
    type: string
    sql: ${TABLE}.assortment_attributes.quality.code_name ;;
    group_label: "Assortment_attributes Quality"
    group_item_label: "Code_name"
  }
  dimension: is_catchweight_item {
    type: yesno
    sql: ${TABLE}.is_catchweight_item ;;
    description: "This attribute determine if ICA sees an item as a catch weight item."
  }
  dimension: category_specific_attributes__execution2__code_description {
    type: string
    sql: ${TABLE}.category_specific_attributes.execution2.code_description ;;
    group_label: "Category_specific_attributes Execution2"
    group_item_label: "Code_description"
  }
  dimension: measurements__gross_weight_in_gram {
    type: number
    sql: ${TABLE}.measurements.gross_weight_in_gram ;;
    group_label: "Measurements"
    group_item_label: "Gross_weight_in_gram"
    description: "(T4020) Used to identify the gross weight of the trade item. The gross weight includes all packaging materials of the trade item. At pallet level the trade item, grossWeight includes the weight of the pallet itself. For example, 200 GRM, value - total pounds, total grams, etc. Has to be associated with a valid UOM."
  }
  dimension: is_corporate_brand {
    type: yesno
    sql: ${TABLE}.is_corporate_brand ;;
    description: "Attribute indicating if the item is Coperate Brand based on rule (list of Coperate Brands)"
  }
  dimension: bica_calculated_fields__bica_improved_weight_volume {
    type: number
    sql: ${TABLE}.bica_calculated_fields.bica_improved_weight_volume ;;
    group_label: "Bica_calculated_fields"
    group_item_label: "Bica_improved_weight_volume"
    description: "Calculated field - Parsed weight or volume from field descriptive_size if net_content_in_gram or net_content_in_miligram is null"
  }
  dimension: category_specific_attributes__specific_content__code_name {
    type: string
    sql: ${TABLE}.category_specific_attributes.specific_content.code_name ;;
    group_label: "Category_specific_attributes Specific_content"
    group_item_label: "Code_name"
  }
  dimension: core_input_reason__code_description {
    type: string
    sql: ${TABLE}.core_input_reason.code_description ;;
    group_label: "Core_input_reason"
    group_item_label: "Code_description"
  }
  dimension: category_specific_attributes__execution1__code_name {
    type: string
    sql: ${TABLE}.category_specific_attributes.execution1.code_name ;;
    group_label: "Category_specific_attributes Execution1"
    group_item_label: "Code_name"
  }
  dimension: segment_id {
    type: string
    sql: ${TABLE}.segment_id ;;
    description: "Merchandise hierarchy node sub category id; e.g 7101.5.3 (prefixed by category_id and subcategory_id)"
  }
  dimension: is_orderable_unit {
    type: yesno
    sql: ${TABLE}.is_orderable_unit ;;
    description: "(T0017) An indicator identifying that the information provider considers this trade item to be at a hierarchy level wherethey will accept orders from customers. This may bedifferent from what the information provider identifies as adespatch unit. This may be a relationship dependent basedon channel of trade or other point to point agreement"
  }
  dimension: assortment_attributes__ica_swedish__code_description {
    type: string
    sql: ${TABLE}.assortment_attributes.ica_swedish.code_description ;;
    group_label: "Assortment_attributes Ica_swedish"
    group_item_label: "Code_description"
  }
  dimension: assortment_attributes__pack_variant__code_name {
    type: string
    sql: ${TABLE}.assortment_attributes.pack_variant.code_name ;;
    group_label: "Assortment_attributes Pack_variant"
    group_item_label: "Code_name"
  }
  dimension: functional_name {
    type: string
    sql: ${TABLE}.functional_name ;;
    description: "Item friendly name"
  }
  dimension: assortment_attributes__quality__code_value {
    type: string
    sql: ${TABLE}.assortment_attributes.quality.code_value ;;
    group_label: "Assortment_attributes Quality"
    group_item_label: "Code_value"
  }
  dimension: category_specific_attributes__consumer_group__code_description {
    type: string
    sql: ${TABLE}.category_specific_attributes.consumer_group.code_description ;;
    group_label: "Category_specific_attributes Consumer_group"
    group_item_label: "Code_description"
  }
  dimension: css_main_category_group_description {
    type: string
    sql: ${TABLE}.css_main_category_group_description ;;
    description: "CSS (Central sortimentstruktur) main category group"
  }
  dimension: category_specific_attributes__origin__code_description {
    type: string
    sql: ${TABLE}.category_specific_attributes.origin.code_description ;;
    group_label: "Category_specific_attributes Origin"
    group_item_label: "Code_description"
  }
  dimension: core_input_reason__code_name {
    type: string
    sql: ${TABLE}.core_input_reason.code_name ;;
    group_label: "Core_input_reason"
    group_item_label: "Code_name"
  }
  dimension: category_specific_attributes__preparation__code_value {
    type: string
    sql: ${TABLE}.category_specific_attributes.preparation.code_value ;;
    group_label: "Category_specific_attributes Preparation"
    group_item_label: "Code_value"
  }
  dimension: assortment_attributes__sustainable {
    type: string
    sql: ${TABLE}.assortment_attributes.sustainable ;;
    group_label: "Assortment_attributes"
    group_item_label: "Sustainable"
    description: "Indicates if item has any markings that is considerad as Sustainable; Hållbar / Ej hållbar"
  }
  dimension: category_specific_attributes__colour__code_name {
    type: string
    sql: ${TABLE}.category_specific_attributes.colour.code_name ;;
    group_label: "Category_specific_attributes Colour"
    group_item_label: "Code_name"
  }
  dimension: category_specific_attributes__colour__code_value {
    type: string
    sql: ${TABLE}.category_specific_attributes.colour.code_value ;;
    group_label: "Category_specific_attributes Colour"
    group_item_label: "Code_value"
  }
  dimension: is_invoice_unit {
    type: yesno
    sql: ${TABLE}.is_invoice_unit ;;
    description: "(T4014) An indicator identifying that the information provider willinclude this trade item on their billing or invoice. This maybe relationship dependent based on channel of trade orother point to point agreement."
  }
  dimension: season__code_value {
    type: string
    sql: ${TABLE}.season.code_value ;;
    group_label: "Season"
    group_item_label: "Code_value"
  }
  dimension: category_id {
    type: string
    sql: ${TABLE}.category_id ;;
    description: "Merchandise hierarchy node category id; e.g 7101"
  }
  dimension: primary_soi_supplier_reference__supplier_organization_name {
    type: string
    sql: ${TABLE}.primary_soi_supplier_reference.supplier_organization_name ;;
    group_label: "Primary_soi_supplier_reference"
    group_item_label: "Supplier_organization_name"
    description: "The name of the Supplier"
  }
  dimension: ecr_category__code_description {
    type: string
    sql: ${TABLE}.ecr_category.code_description ;;
    group_label: "Ecr_category"
    group_item_label: "Code_description"
  }
  dimension: category_specific_attributes__raw_material__code_name {
    type: string
    sql: ${TABLE}.category_specific_attributes.raw_material.code_name ;;
    group_label: "Category_specific_attributes Raw_material"
    group_item_label: "Code_name"
  }
  dimension: category_specific_attributes__execution2__code_name {
    type: string
    sql: ${TABLE}.category_specific_attributes.execution2.code_name ;;
    group_label: "Category_specific_attributes Execution2"
    group_item_label: "Code_name"
  }
  dimension: aggregated_base_item_quantity {
    type: number
    sql: ${TABLE}.aggregated_base_item_quantity ;;
    description: "total quantity of base items in this GTIN , based on packstucture information"
  }
  dimension: category_specific_attributes__raw_material__code_description {
    type: string
    sql: ${TABLE}.category_specific_attributes.raw_material.code_description ;;
    group_label: "Category_specific_attributes Raw_material"
    group_item_label: "Code_description"
  }
  dimension: is_seasonal {
    type: yesno
    sql: ${TABLE}.is_seasonal ;;
    description: "Shows if an item is seasonal."
  }
  dimension: primary_soi_supplier_reference__supplychain_supplier_id {
    type: string
    sql: ${TABLE}.primary_soi_supplier_reference.supplychain_supplier_id ;;
    group_label: "Primary_soi_supplier_reference"
    group_item_label: "Supplychain_supplier_id"
    description: "supplier identification used in supplychain aka MAS-leverantör"
  }
  dimension: assortment_attributes__plantbased__code_name {
    type: string
    sql: ${TABLE}.assortment_attributes.plantbased.code_name ;;
    group_label: "Assortment_attributes Plantbased"
    group_item_label: "Code_name"
  }
  dimension: item_pack_type {
    type: string
    sql: ${TABLE}.item_pack_type ;;
    description: "The pack type of the item; Pallet, Case, "Base Unit or Each" or empty."
  }
  dimension: category_specific_attributes__execution3__code_value {
    type: string
    sql: ${TABLE}.category_specific_attributes.execution3.code_value ;;
    group_label: "Category_specific_attributes Execution3"
    group_item_label: "Code_value"
  }
  dimension: division_id {
    type: string
    sql: ${TABLE}.division_id ;;
    description: "Merchandise hierarchy node Division id; e.g 01"
  }
  dimension: category_specific_attributes__flavour__code_description {
    type: string
    sql: ${TABLE}.category_specific_attributes.flavour.code_description ;;
    group_label: "Category_specific_attributes Flavour"
    group_item_label: "Code_description"
  }
  dimension: brand__code_name {
    type: string
    sql: ${TABLE}.brand.code_name ;;
    group_label: "Brand"
    group_item_label: "Code_name"
  }
  dimension: category_name {
    type: string
    sql: ${TABLE}.category_name ;;
    description: "Merchandise hierarchy node category name; e.g Asiatiska köket"
  }
  dimension: category_specific_attributes__execution3__code_name {
    type: string
    sql: ${TABLE}.category_specific_attributes.execution3.code_name ;;
    group_label: "Category_specific_attributes Execution3"
    group_item_label: "Code_name"
  }
  dimension: item_reporting_description {
    type: string
    sql: ${TABLE}.item_reporting_description ;;
    description: "The item_reporting_description consists of either description, short_description or item_description from s_consumer_item_main or s_item_main"
  }
  dimension: assortment_attributes__gdpr_sensitive__code_value {
    type: string
    sql: ${TABLE}.assortment_attributes.gdpr_sensitive.code_value ;;
    group_label: "Assortment_attributes Gdpr_sensitive"
    group_item_label: "Code_value"
  }
  dimension: brand__code_description {
    type: string
    sql: ${TABLE}.brand.code_description ;;
    group_label: "Brand"
    group_item_label: "Code_description"
  }
  dimension: packaging_information__packaging_weight_uom {
    type: string
    sql: ${TABLE}.packaging_information.packaging_weight_uom ;;
    group_label: "Packaging_information"
    group_item_label: "Packaging_weight_uom"
    description: "The Unit Of Measure for attribute PackagingWeight"
  }
  dimension: css_main_category_group_id {
    type: string
    sql: ${TABLE}.css_main_category_group_id ;;
    description: "CSS (Central sortimentstruktur) main category group"
  }
  dimension: category_description {
    type: string
    sql: ${TABLE}.category_description ;;
    description: "Merchandise hierarchy node category description; concatenation of id and name; e.g 7101 - Asiatiska köket"
  }
  dimension: assortment_attributes__price_range__code_description {
    type: string
    sql: ${TABLE}.assortment_attributes.price_range.code_description ;;
    group_label: "Assortment_attributes Price_range"
    group_item_label: "Code_description"
  }
  dimension: main_category_id {
    type: string
    sql: ${TABLE}.main_category_id ;;
    description: "Merchandise hierarchy node main category id; e.g 101"
  }
  dimension: lifecycle__on_hold_reason {
    type: string
    sql: ${TABLE}.lifecycle.on_hold_reason ;;
    group_label: "Lifecycle"
    group_item_label: "On_hold_reason"
    description: "Reason for onhold status; eg. ICA Delist, Supplier conflict, Supplier out of stock, Seasonal hold"
  }
  dimension: gpc_category_name {
    type: string
    sql: ${TABLE}.gpc_category_name ;;
    description: "Name associated with the specified Global Product Classification (GPC) category code."
  }
  dimension: primary_soi_supplier_reference__soi_description {
    type: string
    sql: ${TABLE}.primary_soi_supplier_reference.soi_description ;;
    group_label: "Primary_soi_supplier_reference"
    group_item_label: "Soi_description"
    description: "Description of SOI"
  }
  dimension: assortment_attributes__multicultural__code_value {
    type: string
    sql: ${TABLE}.assortment_attributes.multicultural.code_value ;;
    group_label: "Assortment_attributes Multicultural"
    group_item_label: "Code_value"
  }
  dimension: assortment_attributes__pack_variant__code_value {
    type: string
    sql: ${TABLE}.assortment_attributes.pack_variant.code_value ;;
    group_label: "Assortment_attributes Pack_variant"
    group_item_label: "Code_value"
  }
  dimension: item_id {
    type: string
    sql: ${TABLE}.item_id ;;
    description: "(T0154) GTIN (Global Trade Item Number, GS1-artikelnummer)"
  }
  dimension: assortment_attributes__packing_size__code_value {
    type: string
    sql: ${TABLE}.assortment_attributes.packing_size.code_value ;;
    group_label: "Assortment_attributes Packing_size"
    group_item_label: "Code_value"
  }
  dimension: category_specific_attributes__consumer_group__code_value {
    type: string
    sql: ${TABLE}.category_specific_attributes.consumer_group.code_value ;;
    group_label: "Category_specific_attributes Consumer_group"
    group_item_label: "Code_value"
  }
  dimension: md_audit_seq {
    type: string
    sql: ${TABLE}.md_audit_seq ;;
    description: "Technical field for specific dbt run"
  }
  dimension: category_specific_attributes__execution4__code_name {
    type: string
    sql: ${TABLE}.category_specific_attributes.execution4.code_name ;;
    group_label: "Category_specific_attributes Execution4"
    group_item_label: "Code_name"
  }
  dimension: md_row_hash {
    type: number
    sql: ${TABLE}.md_row_hash ;;
    description: "Technical field for comparison of attributes"
  }
  dimension: assortment_attributes__ica_swedish__code_name {
    type: string
    sql: ${TABLE}.assortment_attributes.ica_swedish.code_name ;;
    group_label: "Assortment_attributes Ica_swedish"
    group_item_label: "Code_name"
  }
  dimension: measurements__net_content_in_litre {
    type: number
    sql: ${TABLE}.measurements.net_content_in_litre ;;
    group_label: "Measurements"
    group_item_label: "Net_content_in_litre"
    description: "(T0082) The amount of the trade item contained by a package, usually as claimed on the label. For example, Water 750ml - net content = 750 MLT ; 20 count pack of diapers, net content = 20 ea.. In case of multi-pack, indicates the net content of the total trade item. For fixed value trade items use the value claimed on the package, to avoid variable fill rate issue that arises with some trade item which are sold by volume or weight, and whose actual content may vary slightly from batch to batch. In case of variable quantity trade items, indicates the average quantity. Allows for the representation of the same value in different units of measure but not multiple values."
  }
  dimension: lifecycle__central_status {
    type: string
    sql: ${TABLE}.lifecycle.central_status ;;
    group_label: "Lifecycle"
    group_item_label: "Central_status"
    description: "This is the current status of the item Possible values: Draft - This is the item status throughout the Proposed to Accepted process. Once the Proposed to Accepted attribute is set to accepted the item status can be updated to be 'New' New - The Item is now approved for assortment in ICA but some final enrichment is still needed. Active - The item can be set to 'Active' once all criteria is met Phase-Out - An item is set to 'Phase-Out' when the delist date field is populated. On-Hold - The on-hold status simply stops all sell/purchasing of the item. Inactive - Item is made 'Inactive' it has now been either delisted/discontinued or ICA want to remove item Obsolete - Once in this status the item can be purged. It will need to be in the 'Obsolete' status for 24 months prior purging."
  }
  dimension: global_trade_item_number {
    type: string
    sql: ${TABLE}.global_trade_item_number ;;
    description: "(T0154) GTIN (Global Trade Item Number, GS1-artikelnummer)"
  }
  dimension: main_category_name {
    type: string
    sql: ${TABLE}.main_category_name ;;
    description: "Merchandise hierarchy node category name; e.g Asiatiska köket"
  }
  dimension: main_category_description {
    type: string
    sql: ${TABLE}.main_category_description ;;
    description: "Merchandise hierarchy node category description; concatenation of id and name; e.g 7101 - Asiatiska köket"
  }
  dimension: assortment_attributes__swedish__code_name {
    type: string
    sql: ${TABLE}.assortment_attributes.swedish.code_name ;;
    group_label: "Assortment_attributes Swedish"
    group_item_label: "Code_name"
  }
  dimension: price_comparison_unit_of_measure {
    type: string
    sql: ${TABLE}.price_comparison_unit_of_measure ;;
    description: "The Unit Of Measure for the PriceComparisonMeasurement attribute"
  }
  dimension: primary_soi_supplier_reference__supplychain_supplier_short_name {
    type: string
    sql: ${TABLE}.primary_soi_supplier_reference.supplychain_supplier_short_name ;;
    group_label: "Primary_soi_supplier_reference"
    group_item_label: "Supplychain_supplier_short_name"
    description: "Short name of supplychain supplier"
  }
  dimension: standard_unit_of_measure {
    type: string
    sql: ${TABLE}.standard_unit_of_measure ;;
    description: "Automatically default to EACH for all catch weight orderable trade Items (not part of GS1 attribute)."
  }
  dimension: category_specific_attributes__flavour__code_name {
    type: string
    sql: ${TABLE}.category_specific_attributes.flavour.code_name ;;
    group_label: "Category_specific_attributes Flavour"
    group_item_label: "Code_name"
  }
  dimension: measurements__net_content_in_kilogram {
    type: number
    sql: ${TABLE}.measurements.net_content_in_kilogram ;;
    group_label: "Measurements"
    group_item_label: "Net_content_in_kilogram"
    description: "(T0082) The amount of the trade item contained by a package, usually as claimed on the label. For example, Water 750ml - net content = 750 MLT ; 20 count pack of diapers, net content = 20 ea.. In case of multi-pack, indicates the net content of the total trade item. For fixed value trade items use the value claimed on the package, to avoid variable fill rate issue that arises with some trade item which are sold by volume or weight, and whose actual content may vary slightly from batch to batch. In case of variable quantity trade items, indicates the average quantity. Allows for the representation of the same value in different units of measure but not multiple values.Only values having UOM = KILOGRAM"
  }
  dimension: category_specific_attributes__product_group__code_description {
    type: string
    sql: ${TABLE}.category_specific_attributes.product_group.code_description ;;
    group_label: "Category_specific_attributes Product_group"
    group_item_label: "Code_description"
  }
  dimension: sub_category_id {
    type: string
    sql: ${TABLE}.sub_category_id ;;
    description: "Merchandise hierarchy node sub category id; e.g 7101.5 (prefixed by category_id)"
  }
  dimension: item_reporting_id {
    type: string
    sql: ${TABLE}.item_reporting_id ;;
    description: "The item_reporting_id is for display purpose, consists of either global trade item number or item part  where item id represents store unique items (item_id contains |##|)"
  }
  dimension: aggregated_deposit_amount {
    type: number
    sql: ${TABLE}.aggregated_deposit_amount ;;
    description: "total deposit amount including VAT (based aggregated_base_item_quantity and specified amount for each base item)"
  }
  dimension: measurements__height {
    type: number
    sql: ${TABLE}.measurements.height ;;
    group_label: "Measurements"
    group_item_label: "Height"
    description: "(T4019) The height of the unit load, as measured according to the GS1 Package Measurement Rules, including the shipping platform unless it is excluded according to the Pallet Type Code chosen."
  }
  dimension: assortment_attributes__plantbased__code_value {
    type: string
    sql: ${TABLE}.assortment_attributes.plantbased.code_value ;;
    group_label: "Assortment_attributes Plantbased"
    group_item_label: "Code_value"
  }
  dimension: is_consumer_unit {
    type: yesno
    sql: ${TABLE}.is_consumer_unit ;;
    description: "(T4037) Identifies whether the trade item to be taken possession of ,or to be consumed or used by an end user or both, as determined by the manufacturer. The end user could be, but is not limited to, a consumer as in items sold at retail, or a patient/clinician/technician in a healthcare setting, or an operator for foodservice such as restaurants, airlines, cafeterias, etc."
  }
  dimension: assortment_attributes__environmental {
    type: string
    sql: ${TABLE}.assortment_attributes.environmental ;;
    group_label: "Assortment_attributes"
    group_item_label: "Environmental"
    description: "Indicates if item has any markings that is considerad as environmentally good; Miljömärkt / Saknar miljömärkning"
  }
  dimension: primary_soi_supplier_reference__store_orderable_item_id {
    type: string
    sql: ${TABLE}.primary_soi_supplier_reference.store_orderable_item_id ;;
    group_label: "Primary_soi_supplier_reference"
    group_item_label: "Store_orderable_item_id"
    description: "This number is a unique identifier and represents the ICA SOI number aka the MAS artikelnummer"
  }
  dimension: sub_category_name {
    type: string
    sql: ${TABLE}.sub_category_name ;;
    description: "Merchandise hierarchy node category name; e.g Asiatiska köket"
  }
  dimension: assortment_attributes__environmental_non_ecological {
    type: string
    sql: ${TABLE}.assortment_attributes.environmental_non_ecological ;;
    group_label: "Assortment_attributes"
    group_item_label: "Environmental_non_ecological"
    description: "Indicates if item has any markings that is considerad as environmentally good; Miljömärkt / Saknar miljömärkning"
  }
  dimension: gpc_category_definition {
    type: string
    sql: ${TABLE}.gpc_category_definition ;;
    description: "A GS1 supplied definition associated with the specified Global Product Classification (GPC) category code."
  }
  dimension: assortment_attributes__ica_swedish__code_value {
    type: string
    sql: ${TABLE}.assortment_attributes.ica_swedish.code_value ;;
    group_label: "Assortment_attributes Ica_swedish"
    group_item_label: "Code_value"
  }
  dimension: measurements__height_unit_of_measure {
    type: string
    sql: ${TABLE}.measurements.height_unit_of_measure ;;
    group_label: "Measurements"
    group_item_label: "Height_unit_of_measure"
    description: "(T3780) unit of measure value associated to height value"
  }
  dimension: consumer_item_reference__consumer_item_id {
    type: string
    sql: ${TABLE}.consumer_item_reference.consumer_item_id ;;
    group_label: "Consumer_item_reference"
    group_item_label: "Consumer_item_id"
    description: "Identifier for the Consumer item, equivalent to EMS Store Item number"
  }
  dimension: division_description {
    type: string
    sql: ${TABLE}.division_description ;;
    description: "Merchandise hierarchy node category description; concatenation of id and name; e.g 7101 - Asiatiska köket"
  }
  dimension: css_main_category_group_name {
    type: string
    sql: ${TABLE}.css_main_category_group_name ;;
    description: "CSS (Central sortimentstruktur) main category group"
  }
  dimension: assortment_attributes__packing_size__code_name {
    type: string
    sql: ${TABLE}.assortment_attributes.packing_size.code_name ;;
    group_label: "Assortment_attributes Packing_size"
    group_item_label: "Code_name"
  }
  dimension: country_of_origin {
    type: string
    sql: d_item_v3__country_of_origin ;;
    hidden: yes
  }
  dimension: ica_ecological_accreditation {
    type: string
    sql: d_item_v3__ica_ecological_accreditation ;;
    hidden: yes
  }
  dimension: central_department {
    type: string
    sql: d_item_v3__central_department ;;
    hidden: yes
  }
  dimension: item_information_claim_detail {
    type: string
    sql: d_item_v3__item_information_claim_detail ;;
    hidden: yes
  }
  dimension: ica_swedish_accreditation {
    type: string
    sql: d_item_v3__ica_swedish_accreditation ;;
    hidden: yes
  }
  dimension: load_carrier_deposit {
    type: string
    sql: d_item_v3__load_carrier_deposit ;;
    hidden: yes
  }
  dimension: accreditation {
    type: string
    sql: d_item_v3__accreditation ;;
    hidden: yes
  }
  dimension: ica_environmental_accreditation {
    type: string
    sql: d_item_v3__ica_environmental_accreditation ;;
    hidden: yes
  }
  dimension: packaging_information__packaging_material_composition {
    type: string
    sql: d_item_v3__packaging_information__packaging_material_composition ;;
    hidden: yes
  }
  dimension: ica_non_ecological_accreditation {
    type: string
    sql: d_item_v3__ica_non_ecological_accreditation ;;
    hidden: yes
  }
  dimension: ica_ethical_accreditation {
    type: string
    sql: d_item_v3__ica_ethical_accreditation ;;
    hidden: yes
  }
  dimension_group: season_start {
    type: time
    sql: ${TABLE}.season_start_date ;;
    timeframes: [raw, date, week, month, quarter, year]
  }

  dimension_group: ecr_revision {
    type: time
    sql: ${TABLE}.ecr_revision_date ;;
    timeframes: [raw, date, week, month, quarter, year]
  }

  dimension_group: lifecycle__reactivation {
    type: time
    sql: ${TABLE}.lifecycle.reactivation_date ;;
    timeframes: [raw, date, week, month, quarter, year]
  }

  dimension_group: lifecycle__obsolete {
    type: time
    sql: ${TABLE}.lifecycle.obsolete_date ;;
    timeframes: [raw, date, week, month, quarter, year]
  }

  dimension_group: primary_soi_supplier_reference__orderability_start {
    type: time
    sql: ${TABLE}.primary_soi_supplier_reference.orderability_start_date ;;
    timeframes: [raw, date, week, month, quarter, year]
  }

  dimension_group: lifecycle__novelty_end {
    type: time
    sql: ${TABLE}.lifecycle.novelty_end_date ;;
    timeframes: [raw, date, week, month, quarter, year]
  }

  dimension_group: lifecycle__creation {
    type: time
    sql: ${TABLE}.lifecycle.creation_datetime ;;
    timeframes: [raw, time, date, week, month, quarter, year]
  }

  dimension_group: lifecycle__purge {
    type: time
    sql: ${TABLE}.lifecycle.purge_date ;;
    timeframes: [raw, date, week, month, quarter, year]
  }

  dimension_group: season_end {
    type: time
    sql: ${TABLE}.season_end_date ;;
    timeframes: [raw, date, week, month, quarter, year]
  }

  dimension_group: md_insert_dttm {
    type: time
    sql: ${TABLE}.md_insert_dttm ;;
    timeframes: [raw, time, date, week, month, quarter, year]
  }

  dimension_group: primary_soi_supplier_reference__delivery_start {
    type: time
    sql: ${TABLE}.primary_soi_supplier_reference.delivery_start_date ;;
    timeframes: [raw, date, week, month, quarter, year]
  }

  dimension_group: lifecycle__on_hold_start {
    type: time
    sql: ${TABLE}.lifecycle.on_hold_start_date ;;
    timeframes: [raw, date, week, month, quarter, year]
  }

  dimension_group: lifecycle__novelty_start {
    type: time
    sql: ${TABLE}.lifecycle.novelty_start_date ;;
    timeframes: [raw, date, week, month, quarter, year]
  }

  dimension_group: lifecycle__ica_discontinue {
    type: time
    sql: ${TABLE}.lifecycle.ica_discontinue_date ;;
    timeframes: [raw, date, week, month, quarter, year]
  }

  dimension_group: primary_soi_supplier_reference__orderability_end {
    type: time
    sql: ${TABLE}.primary_soi_supplier_reference.orderability_end_date ;;
    timeframes: [raw, date, week, month, quarter, year]
  }

  measure: count {
    type: count
  }

}

view: d_item_v3__item_information_claim_detail {
  sql_table_name:  ;;
  dimension: claim_element__claim_element_code_name {
    type: string
    sql: ${TABLE}.claim_element.claim_element_code_name ;;
  }
  dimension: claim_element__claim_element_code_description {
    type: string
    sql: ${TABLE}.claim_element.claim_element_code_description ;;
  }
  dimension: claim_type__claim_type_code_name {
    type: string
    sql: ${TABLE}.claim_type.claim_type_code_name ;;
  }
  dimension: claim_element__claim_element_code_value {
    type: string
    sql: ${TABLE}.claim_element.claim_element_code_value ;;
  }
  dimension: item_information_claim_detail_code_value {
    type: string
    sql: ${TABLE}.item_information_claim_detail_code_value ;;
    description: "(T4358, T4359) Combination of code_values for claim_type and claim_element, e.g. FREE_FROM GLUTEN, LOW_ON LACTOSE"
  }
  dimension: claim_type__claim_type_code_description {
    type: string
    sql: ${TABLE}.claim_type.claim_type_code_description ;;
  }
  dimension: is_item_information_claim_marked_on_package {
    type: yesno
    sql: ${TABLE}.is_item_information_claim_marked_on_package ;;
    description: "(T4357) Item information claim details is marked on packaage (true/false)"
  }
  dimension: claim_type__claim_type_code_value {
    type: string
    sql: ${TABLE}.claim_type.claim_type_code_value ;;
  }
  dimension: d_item_v3__item_information_claim_detail {
    type: string
    sql: ${TABLE}.item_information_claim_detail ;;
    description: "(T4357, T4358, T4359) Item information claim details"
    hidden: yes
  }
  dimension: item_information_claim_detail_code_name {
    type: string
    sql: ${TABLE}.item_information_claim_detail_code_name ;;
    description: "(T4358, T4359) Combination of code_names for claim_type and claim_element, e.g. Fri från Gluten, Låg Laktos"
  }
}

view: d_item_v3__ica_ecological_accreditation {
  sql_table_name:  ;;
  dimension: ica_ecological_accreditation_name {
    type: string
    sql: ${TABLE}.ica_ecological_accreditation_name ;;
  }
  dimension: d_item_v3__ica_ecological_accreditation {
    type: string
    sql: ${TABLE}.ica_ecological_accreditation ;;
    description: "(T3777) Item accreditations considered as environmental and ecological by ICA, see detail on BICA wiki, subset of accredition-attribute"
    hidden: yes
  }
  dimension: ica_ecological_accreditation_code {
    type: string
    sql: ${TABLE}.ica_ecological_accreditation_code ;;
  }
  dimension: ica_ecological_accreditation_description {
    type: string
    sql: ${TABLE}.ica_ecological_accreditation_description ;;
  }
}

view: d_item_v3__packaging_information__packaging_material_composition__packaging_material_composition_quantity {
  sql_table_name:  ;;
  dimension: d_item_v3__packaging_information__packaging_material_composition__packaging_material_composition_quantity {
    type: string
    sql: ${TABLE}.packaging_information.packaging_material_composition.packaging_material_composition_quantity ;;
    hidden: yes
  }
  dimension: quantity_unit_of_measure {
    type: string
    sql: ${TABLE}.quantity_unit_of_measure ;;
    description: "The Unit Of Measure for the PackagingMaterialCompositionQuantity attribute."
  }
  dimension: quantity_value {
    type: number
    sql: ${TABLE}.quantity_value ;;
    description: "The quantity of the packaging material of the trade item. Can be weight, volume or surface, can vary by country."
  }
}

view: d_item_v3__packaging_information__packaging_material_composition {
  sql_table_name:  ;;
  dimension: packaging_material_composition_quantity {
    type: string
    sql: ${TABLE}.packaging_material_composition_quantity ;;
  }
  dimension: packaging_material_type__code_description {
    type: string
    sql: ${TABLE}.packaging_material_type.code_description ;;
    description: "The materials used for the packaging of the trade item for example glass or plastic. This material information can be used by data recipients for; o Tax calculations/fees/duties calculation o Carbon footprint calculations/estimations (resource optimisation) o to determine the material used."
  }
  dimension: packaging_material_type__code_name {
    type: string
    sql: ${TABLE}.packaging_material_type.code_name ;;
    description: "The materials used for the packaging of the trade item for example glass or plastic. This material information can be used by data recipients for; o Tax calculations/fees/duties calculation o Carbon footprint calculations/estimations (resource optimisation) o to determine the material used."
  }
  dimension: packaging_material_type__code_value {
    type: string
    sql: ${TABLE}.packaging_material_type.code_value ;;
    description: "The materials used for the packaging of the trade item for example glass or plastic. This material information can be used by data recipients for; o Tax calculations/fees/duties calculation o Carbon footprint calculations/estimations (resource optimisation) o to determine the material used."
  }
  dimension: d_item_v3__packaging_information__packaging_material_composition {
    type: string
    sql: ${TABLE}.packaging_information.packaging_material_composition ;;
    hidden: yes
  }
}

view: d_item_v3__central_department {
  sql_table_name:  ;;
  dimension: profile_id {
    type: string
    sql: ${TABLE}.profile_id ;;
    description: "Store profile GLN thats connected to current central department"
  }
  dimension: d_item_v3__central_department {
    type: string
    sql: ${TABLE}.central_department ;;
    description: "Department (used for central analysis close to store, maintained by Store and Marketing sponsor area)"
    hidden: yes
  }
  dimension: central_department_name {
    type: string
    sql: ${TABLE}.central_department_name ;;
    description: "department (used for central analysis close to store , maintained by Store and Marketing sponsor area)"
  }
  dimension: central_department_description {
    type: string
    sql: ${TABLE}.central_department_description ;;
    description: "department (used for central analysis close to store , maintained by Store and Marketing sponsor area)"
  }
  dimension: profile_name {
    type: string
    sql: ${TABLE}.profile_name ;;
    description: "Store profile name thats connected to current central department"
  }
  dimension: central_department_code {
    type: string
    sql: ${TABLE}.central_department_code ;;
    description: "department (used for central analysis close to store , maintained by Store and Marketing sponsor area)"
  }
}

view: d_item_v3__ica_swedish_accreditation {
  sql_table_name:  ;;
  dimension: ica_swedish_accreditation_code {
    type: string
    sql: ${TABLE}.ica_swedish_accreditation_code ;;
  }
  dimension: d_item_v3__ica_swedish_accreditation {
    type: string
    sql: ${TABLE}.ica_swedish_accreditation ;;
    description: "(T3777) Item accreditations considered as swedish by ICA, see detail on BICA wiki, subset of accredition-attribute"
    hidden: yes
  }
  dimension: ica_swedish_accreditation_name {
    type: string
    sql: ${TABLE}.ica_swedish_accreditation_name ;;
  }
  dimension: ica_swedish_accreditation_description {
    type: string
    sql: ${TABLE}.ica_swedish_accreditation_description ;;
  }
}

view: d_item_v3__load_carrier_deposit {
  sql_table_name:  ;;
  dimension: base_item_quantity {
    type: number
    sql: ${TABLE}.base_item_quantity ;;
    description: "quantity of base items in this GTIN , based on packstucture information"
  }
  dimension: returnable_asset_contained_quantity {
    type: number
    sql: ${TABLE}.returnable_asset_contained_quantity ;;
    description: "(T4125) Number of deposit items per item"
  }
  dimension: returnable_asset_deposit_type {
    type: string
    sql: ${TABLE}.returnable_asset_deposit_type ;;
    description: "(T0148) Type of deposit item (Container,Crate,LoadCarrier)"
  }
  dimension: deposit_amount {
    type: number
    sql: ${TABLE}.deposit_amount ;;
    description: "deposit amount (returnable_asset_contained_quantity*returnable_package_deposit_amount)"
  }
  dimension: returnable_package_deposit_amount {
    type: number
    sql: ${TABLE}.returnable_package_deposit_amount ;;
    description: "(T0148) Deposit value per deposit asset incluiding VAT"
  }
  dimension: returnable_asset_deposit_name {
    type: string
    sql: ${TABLE}.returnable_asset_deposit_name ;;
    description: "(T0148) Depositname e.g. Engångs Pet över 1000 ml"
  }
  dimension: d_item_v3__load_carrier_deposit {
    type: string
    sql: ${TABLE}.load_carrier_deposit ;;
    description: "(Record) returnable asset details"
    hidden: yes
  }
}

view: d_item_v3__accreditation {
  sql_table_name:  ;;
  dimension: accreditation_description {
    type: string
    sql: ${TABLE}.accreditation_description ;;
  }
  dimension: accreditation_code {
    type: string
    sql: ${TABLE}.accreditation_code ;;
  }
  dimension: d_item_v3__accreditation {
    type: string
    sql: ${TABLE}.accreditation ;;
    description: "(T3777) All item acceditations (GS1 CodeList PackagingMarkedLabelAccreditationCode)"
    hidden: yes
  }
  dimension: accreditation_name {
    type: string
    sql: ${TABLE}.accreditation_name ;;
  }
}

view: d_item_v3__ica_non_ecological_accreditation {
  sql_table_name:  ;;
  dimension: ica_non_ecological_accreditation_description {
    type: string
    sql: ${TABLE}.ica_non_ecological_accreditation_description ;;
  }
  dimension: ica_non_ecological_accreditation_code {
    type: string
    sql: ${TABLE}.ica_non_ecological_accreditation_code ;;
  }
  dimension: d_item_v3__ica_non_ecological_accreditation {
    type: string
    sql: ${TABLE}.ica_non_ecological_accreditation ;;
    description: "(T3777) Item accreditations considered as environmental and non-ecological by ICA, see detail on BICA wiki, subset of accredition-attribute"
    hidden: yes
  }
  dimension: ica_non_ecological_accreditation_name {
    type: string
    sql: ${TABLE}.ica_non_ecological_accreditation_name ;;
  }
}

view: d_item_v3__ica_environmental_accreditation {
  sql_table_name:  ;;
  dimension: ica_environmental_accreditation_code {
    type: string
    sql: ${TABLE}.ica_environmental_accreditation_code ;;
  }
  dimension: ica_environmental_accreditation_description {
    type: string
    sql: ${TABLE}.ica_environmental_accreditation_description ;;
  }
  dimension: ica_environmental_accreditation_name {
    type: string
    sql: ${TABLE}.ica_environmental_accreditation_name ;;
  }
  dimension: d_item_v3__ica_environmental_accreditation {
    type: string
    sql: ${TABLE}.ica_environmental_accreditation ;;
    description: "(T3777) Item accreditations considered as environmental by ICA, see detail on BICA wiki, subset of accredition-attribute"
    hidden: yes
  }
}

view: d_item_v3__country_of_origin {
  sql_table_name:  ;;
  dimension: d_item_v3__country_of_origin {
    type: string
    sql: ${TABLE}.country_of_origin ;;
    description: "The country the item may have originated from, has been processed in. Etc."
    hidden: yes
  }
}

view: d_item_v3__ica_ethical_accreditation {
  sql_table_name:  ;;
  dimension: ica_ethical_accreditation_code {
    type: string
    sql: ${TABLE}.ica_ethical_accreditation_code ;;
  }
  dimension: ica_ethical_accreditation_name {
    type: string
    sql: ${TABLE}.ica_ethical_accreditation_name ;;
  }
  dimension: d_item_v3__ica_ethical_accreditation {
    type: string
    sql: ${TABLE}.ica_ethical_accreditation ;;
    description: "(T3777) Item accreditations considered as ethical by ICA, see detail on BICA wiki, subset of accredition-attribute"
    hidden: yes
  }
  dimension: ica_ethical_accreditation_description {
    type: string
    sql: ${TABLE}.ica_ethical_accreditation_description ;;
  }
}
