use actix_web::{web, App, HttpResponse, Responder, HttpServer};
use serde::{Deserialize, Serialize};

// Define a struct for your JSON response
#[derive(Serialize)]
struct MyResponse {
    message: String,
}

// Handler function returning JSON response
async fn index() -> impl Responder {
    let response_data = MyResponse {
        message: "Hello, RUST!".to_string(),
    };
    HttpResponse::Ok().json(response_data)
}

async fn greet() -> impl Responder {
    HttpResponse::Ok().body("Hello, RUST!")
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    HttpServer::new(|| {
        App::new()
            .route("/", web::get().to(index))
            .route("/greet", web::get().to(greet))
    })
    .bind(("0.0.0.0", 8090))?
    .run()
    .await
}
