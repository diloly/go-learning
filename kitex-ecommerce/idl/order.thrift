namespace go order

struct CreateOrderReq {
    1: i64 user_id
    2: i64 product_id
    3: i64 quantity
    4: string address
}

struct CreateOrderResp {
    1: string order_id
    2: string status
}

struct GetOrderReq {
    1: string order_id
}

struct OrderResp {
    1: string order_id
    2: i64 user_id
    3: i64 product_id
    4: i64 quantity
    5: double total_price
    6: string status
    7: string address
}

struct ListOrdersReq {
    1: i64 user_id
    2: i32 page
    3: i32 page_size
}

struct ListOrdersResp {
    1: list<OrderResp> orders
    2: i64 total
}

service OrderService {
    CreateOrderResp CreateOrder(1: CreateOrderReq req)
    OrderResp GetOrder(1: GetOrderReq req)
    ListOrdersResp ListOrders(1: ListOrdersReq req)
}
