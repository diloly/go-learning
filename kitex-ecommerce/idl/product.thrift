namespace go product

struct CreateProductReq {
    1: string name
    2: string description
    3: double price
    4: i64 stock
}

struct ProductResp {
    1: i64 id
    2: string name
    3: string description
    4: double price
    5: i64 stock
}

struct GetProductReq {
    1: i64 product_id
}

struct ListProductsReq {
    1: i32 page
    2: i32 page_size
}

struct ListProductsResp {
    1: list<ProductResp> products
    2: i64 total
}

struct DeductStockReq {
    1: i64 product_id
    2: i64 quantity
}

struct DeductStockResp {
    1: bool success
}

service ProductService {
    ProductResp CreateProduct(1: CreateProductReq req)
    ProductResp GetProduct(1: GetProductReq req)
    ListProductsResp ListProducts(1: ListProductsReq req)
    DeductStockResp DeductStock(1: DeductStockReq req)
}
