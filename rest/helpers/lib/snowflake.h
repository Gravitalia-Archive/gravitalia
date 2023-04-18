#ifndef SNOWFLAKE_SNOWFLAKE_H
#define SNOWFLAKE_SNOWFLAKE_H

#include <pthread.h>
#include <stdio.h>
#include <sys/time.h>

#define SNOWFLAKE_EPOCH 1672600000000 // Sunday 1 January 2023 19:06:40
#define SNOWFLAKE_REGIONID_BITS 5
#define SNOWFLAKE_WORKERID_BITS 5
#define SNOWFLAKE_SEQUENCE_BITS 12

long int cur_timestamp;
long int seq_max;
long int worker_id;
long int region_id;
long int seq;
long int time_shift_bits;
long int region_shift_bits;
long int worker_shift_bits;

struct snowflake {
    long int timestamp;
    long int worker_id;
    long int region_id;
    long int increment;
};

long int get_id();
int init(int new_region_id, int new_worker_id);
struct snowflake reverse(long int snowflake);

#endif //SNOWFLAKE_SNOWFLAKE_H