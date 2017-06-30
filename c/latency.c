#include <stdlib.h>
#include <stdio.h>
#include <time.h>
#include <unistd.h>

#include <sys/time.h>

static unsigned long worst_latency = 0;
static unsigned long best_latency = 1000000L;
static unsigned long sum_latency = 0;
static size_t buffer_size = 4096*10;

static int save_latencies(char *path, int cycles, int *latencies) {
	int i;
	FILE *f;

	f = fopen(path, "w+");
	if (f == NULL) {
		return -1;
	}

	for (i = 0; i < cycles; i++) {
		fprintf(f, "%d\n", latencies[i]);
	}

	fclose(f);

	return 0;
}

static void free_buffers(char **buffers, int num_buffers) {
	int i;

	for (i = 0; i < num_buffers; i++) {
		if (buffers[i] != NULL) {
			free(buffers[i]);
		}
	}

	free(buffers);
}

static int idle_thread(int cycles, int period, int num_buffers, char *path)
{
	struct timespec sleep, t0, t1;
	int i, j, k, total_t0, total_t1, latency;
	char **buffers;
	int *latencies;

	sleep.tv_sec = 0;
	sleep.tv_nsec = period * 1000000L;

	latencies = (int *)calloc(cycles, sizeof(int));

	for (i = 0; i < cycles; i++) {
		buffers = calloc(num_buffers, sizeof(char *));
		if (buffers == NULL) {
			free(latencies);
			return -1;
		}

		for (j = 0; j < num_buffers; j++) {
			buffers[j] = calloc(buffer_size, sizeof(char));
			if (buffers[j] == NULL) {
				free_buffers(buffers, num_buffers);
				free(latencies);
				return -1;
			}

			for (k = 0; k < buffer_size; k++) {
				buffers[j][k] = k;
			}
		}


		clock_gettime(CLOCK_REALTIME, &t0);
		if(nanosleep(&sleep , NULL) < 0 )  {
			printf("Nano sleep system call failed \n");
			free_buffers(buffers, num_buffers);
			free(latencies);
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

		latencies[i] = latency;
		sum_latency += latency;

		free_buffers(buffers, num_buffers);
	}

	if (path != NULL) {
		save_latencies(path, cycles, latencies);
	}

	free(latencies);

	return 0;
}

int main(int argc, char *argv[])
{
	int cycles = 500, period = 100, opt = 0, buffers = 10;
	char *file = NULL;

	while ((opt = getopt(argc, argv, "b:c:f:p:")) != -1) {
		switch (opt) {
		case 'b':
			buffers = atoi(optarg);
			break;
		case 'c':
			cycles = atoi(optarg);
			break;
		case 'p':
			period = atoi(optarg);
			break;
		case 'f':
			file = optarg;
			break;
		default:
			fprintf(stderr, "Usage: %s [-cpb]\n", argv[0]);
			exit(EXIT_FAILURE);
		}
	}

	printf("%d cycles, %d ms sleep period\n", cycles, period);

	if (idle_thread(cycles, period, buffers, file) < 0) {
		exit(EXIT_FAILURE);
	}

	printf("Latency: [Avg %ld µs, Best %ld µs, Worst %ld µs]\n", sum_latency/cycles, best_latency, worst_latency);
}
