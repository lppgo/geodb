package helpers

import (
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/jsonpb"
)

var jpb = jsonpb.Marshaler{}

func PrettyJson(msg proto.Message) string {
	str, _ := jpb.MarshalToString(msg)
	return fmt.Sprintln(str)
}
