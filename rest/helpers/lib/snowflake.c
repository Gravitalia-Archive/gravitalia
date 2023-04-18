#include "snowflake.h"

long int get_id() {
    struct timeval tp;
    gettimeofday(&tp, NULL);
    long int millisecs = tp.tv_sec * 1000 + tp.tv_usec / 1000;
    long int id;
    if ((seq > seq_max) || cur_timestamp > millisecs) {
        while (cur_timestamp >= millisecs) {
            gettimeofday(&tp, NULL);
            millisecs = tp.tv_sec * 1000 + tp.tv_usec / 1000;
        }
    }
    if (cur_timestamp < millisecs) {
        cur_timestamp = millisecs;
        seq = 0L;
    }
    id = ((millisecs - SNOWFLAKE_EPOCH) << time_shift_bits)
         | (region_id << region_shift_bits)
         | (worker_id << worker_shift_bits)
         | (seq++);

    return id;
}

struct snowflake reverse(long int snowflake) {
    struct snowflake value;
    value.timestamp = ((snowflake >> (time_shift_bits)) + SNOWFLAKE_EPOCH) / 1000;
    value.region_id = (snowflake & 0x3E0000) >> (region_shift_bits);
    value.worker_id = (snowflake & 0x1F000) >> (worker_shift_bits);
    value.increment = snowflake & 0xFFF;
    return value;
}

int init(int new_region_id, int new_worker_id) {
    int max_region_id = (1 << SNOWFLAKE_REGIONID_BITS) - 1;
    if (new_region_id < 0 || new_region_id > max_region_id) {
        printf("Region ID must be in the range : 0-%d\n", max_region_id);
        return -1;
    }
    int max_worker_id = (1 << SNOWFLAKE_WORKERID_BITS) - 1;
    if (new_worker_id < 0 || new_worker_id > max_worker_id) {
        printf("Worker ID must be in the range: 0-%d\n", max_worker_id);
        return -1;
    }
    time_shift_bits = SNOWFLAKE_REGIONID_BITS + SNOWFLAKE_WORKERID_BITS + SNOWFLAKE_SEQUENCE_BITS;
    region_shift_bits = SNOWFLAKE_WORKERID_BITS + SNOWFLAKE_SEQUENCE_BITS;
    worker_shift_bits = SNOWFLAKE_SEQUENCE_BITS;

    worker_id = new_worker_id;
    region_id = new_region_id;
    seq_max = (1L << SNOWFLAKE_SEQUENCE_BITS) - 1;
    seq = 0L;
    cur_timestamp = 0L;
    return 0;
}