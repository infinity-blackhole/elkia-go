use clap::Parser;
use elkia_gateway::run_server;
use std::error::Error;
use tokio::net::TcpListener;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    /// Address to listen on
    #[arg(long, env = "ADDR", default_value = "0.0.0.0:4000")]
    addr: String,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    env_logger::init();

    let args = Args::parse();

    let listener = TcpListener::bind(&args.addr).await?;

    run_server(listener).await
}
