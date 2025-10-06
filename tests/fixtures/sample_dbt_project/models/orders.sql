{{
    config(
        materialized='table'
    )
}}

select
    1 as order_id,
    101 as customer_id,
    150.00 as amount,
    '2024-01-15'::date as order_date,
    'completed' as status

union all

select
    2 as order_id,
    102 as customer_id,
    275.50 as amount,
    '2024-01-16'::date as order_date,
    'completed' as status

union all

select
    3 as order_id,
    101 as customer_id,
    89.99 as amount,
    '2024-01-17'::date as order_date,
    'pending' as status
