#include <stdlib.h>
#include <stdio.h>
#include <time.h>
#include <unistd.h>

#include <sys/time.h>

static unsigned int worst_latency = 0;
static unsigned int best_latency = 1000000L;
static unsigned int sum_latency = 0;

static int idle_thread(int cycles, int period)
{
	struct timespec sleep, t0, t1;
	int i, total_t0, total_t1, latency;

	sleep.tv_sec = 0;
	sleep.tv_nsec = period * 1000000L;

	for (i = 0; i < cycles; i++) {

		clock_gettime(CLOCK_REALTIME, &t0);
		if(nanosleep(&sleep , NULL) < 0 )  {
			printf("Nano sleep system call failed \n");
			return -1;
		}
		clock_gettime(CLOCK_REALTIME, &t1);

		total_t0 = t0.tv_sec * 1000000000L + t0.tv_nsec;
		total_t1 = t1.tv_sec * 1000000000L + t1.tv_nsec;
		latency = ((total_t1 - total_t0) - (period * 1000000L))/1000L;

		if (latency > worst_latency) {
			worst_latency = latency;
		}

		if (latency < best_latency) {
			best_latency = latency;
		}

		sum_latency += latency;
	}

	return 0;
}

int main(int argc, char *argv[])
{
	int cycles = 500, period = 100, opt = 0;

	while ((opt = getopt(argc, argv, "c:p:")) != -1) {
		switch (opt) {
		case 'c':
			cycles = atoi(optarg);
			break;
		case 'p':
			period = atoi(optarg);
			break;
		default:
			fprintf(stderr, "Usage: %s [-cp]\n", argv[0]);
			exit(EXIT_FAILURE);
		}
	}

	printf("%ld cycles, %ld ms sleep period\n", cycles, period);

	if (idle_thread(cycles, period) < 0) {
		exit(EXIT_FAILURE);
	}

	printf("Latency: [Avg %ld µs, Best %ld µs, Worst %ld µs]\n", sum_latency/cycles, best_latency, worst_latency);
}
