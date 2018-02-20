extern crate clap;

use clap::{Arg, App};
use std::{thread, time};
use std::str::FromStr;

const BUFFER_SIZE: usize = 4096*10;

fn allocate_and_sleep(allocation: bool, period_millis: u64) -> u64 {

    if allocation {
        for _i in 1..10 {
            let mut vec: Vec<u32> = Vec::with_capacity(BUFFER_SIZE);
            for _j in 0..BUFFER_SIZE {
                vec.push(_j as u32)
            }
        }
    }

    let period_duration = time::Duration::from_millis(period_millis);

    let now = time::Instant::now();
    thread::sleep(period_duration);
    let elapsed = now.elapsed();

    let t = elapsed - period_duration;

    u64::from(t.subsec_nanos())
}

fn main() {
    let matches = App::new("Scheduling latency for Rust")
        .arg(Arg::with_name("cycles")
             .short("c")
             .long("cycles")
             .value_name("number of test cycles")
             .help("Test cycles")
             .takes_value(true))
        .arg(Arg::with_name("period")
             .short("p")
             .long("period")
             .value_name("period in milliseconds")
             .help("Thread sleeping period")
             .takes_value(true))
        .arg(Arg::with_name("allocation")
             .short("a")
             .long("alloc")
             .help("Allocate memory on the heap"))
        .get_matches();

    let c = matches.value_of("cycles").unwrap();
    let cycles = u64::from_str(c).unwrap();

    let p = matches.value_of("period").unwrap();
    let period = u64::from_str(p).unwrap();

    let mut allocation = false;
    if matches.is_present("allocation") {
        allocation = true;
    }

    let mut sum_latency = 0;
    let mut worst_latency = 0;
    let mut best_latency = 1000000000;
    
    for _i in 1..cycles {
        let latency = allocate_and_sleep(allocation, period);
        if latency < best_latency {
            best_latency = latency
        }

        if latency > worst_latency {
            worst_latency = latency
        }

        sum_latency += latency
    }

    println!("Latency: [Avg {} µs, best {} µs, worst {}µs]\n", (sum_latency/cycles)/1000, best_latency/1000, worst_latency/1000);
}
