-- name: GetOrderByID :one
select * from orders
where id = $1;

-- name: CreateOrder :one
insert into orders (quantity, shipping_address, sku, idempotency_key) values ($1, $2, $3, $4) returning *;

-- name: GetOrderByIdempotencyKey :one
select * from orders
where idempotency_key = $1;
