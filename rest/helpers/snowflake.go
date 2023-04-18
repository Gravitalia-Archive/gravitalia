package helpers

/*

#include "lib/snowflake.c"

*/
import "C"
import (
	"strconv"

	"github.com/Gravitalia/gravitalia/model"
)

func Init() bool {
	C.init(1, 1)
	return true
}

func Generate() string {
	return strconv.Itoa(int(C.get_id()))
}

func Reverse(id C.long) *model.Snowflake {
	rev := C.reverse(id)
	return &model.Snowflake{Timestamp: int64(rev.timestamp), Worker_id: int(rev.worker_id), Region_id: int(rev.region_id), Increment: int(rev.increment)}
}
