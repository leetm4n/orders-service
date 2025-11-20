-- migrate:up
create type order_status as enum (
    'pending',
    'shipped',
    'delivered',
    'canceled'
);

create table if not exists orders (
    id uuid primary key default gen_random_uuid(),
    quantity int not null,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now(),
    status order_status not null default 'pending',
    idempotency_key uuid unique,
    shipping_address text not null,
    sku uuid not null
);

create index if not exists idx_orders_idempotency_key on orders (idempotency_key);

-- migrate:down
drop table if exists orders;
drop type if exists order_status;