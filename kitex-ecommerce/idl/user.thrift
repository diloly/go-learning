namespace go user

struct RegisterReq {
    1: string username
    2: string password
    3: string nickname
}

struct RegisterResp {
    1: i64 user_id
}

struct LoginReq {
    1: string username
    2: string password
}

struct LoginResp {
    1: i64 user_id
    2: string token
}

struct GetUserReq {
    1: i64 user_id
}

struct GetUserResp {
    1: i64 user_id
    2: string username
    3: string nickname
}

service UserService {
    RegisterResp Register(1: RegisterReq req)
    LoginResp Login(1: LoginReq req)
    GetUserResp GetUser(1: GetUserReq req)
}
