package utility

import (
	"encoding/binary"
	"encoding/json"

	"server-prototype/protocol"
)

// packet:
// 0~4 tag
// 5~8 length
// 9~  value

func EncodeStruct(protoStruct interface{}, messageID protocol.MsgID) ([]byte, error) {
	protoStructByteData, protoStructMarshalError := json.Marshal(protoStruct)
	if protoStructMarshalError != nil {
		return []byte{}, protoStructMarshalError
	}

	tagByteDataInBigEndian := make([]byte, 4)
	binary.BigEndian.PutUint32(tagByteDataInBigEndian, uint32(messageID))

	lengthByteDataInBigEndian := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthByteDataInBigEndian, uint32(len(protoStructByteData)))

	encodeMessageByteData := make([]byte, 0, 4+4+len(protoStructByteData))
	encodeMessageByteData = append(encodeMessageByteData, tagByteDataInBigEndian...)
	encodeMessageByteData = append(encodeMessageByteData, lengthByteDataInBigEndian...)
	encodeMessageByteData = append(encodeMessageByteData, protoStructByteData...)

	return encodeMessageByteData, nil
}
