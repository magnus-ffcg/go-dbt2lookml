{{
    config(
        materialized='table'
    )
}}

with date_spine as (
    select
        date '2024-01-01' + interval (row_number() over (order by 1) - 1) day as date_day
    from generate_series(1, 365)
)

select
    date_day as date_day
from date_spine
