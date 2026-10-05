package genid

import (
	"github.com/bwmarrin/snowflake"
)

type SnowflakeGenerator struct {
	Snowflake *snowflake.Node
}

var Snowflake *SnowflakeGenerator

func New(node int64) {
	snowflakeNode, _ := snowflake.NewNode(node)
	Snowflake = &SnowflakeGenerator{
		Snowflake: snowflakeNode,
	}
}

func (s *SnowflakeGenerator) GenerateId() int64 {
	return s.Snowflake.Generate().Int64()
}
