package model

type Snowflake struct {
	Timestamp int64
	Worker_id int
	Region_id int
	Increment int
}
