package utils

import (
	"math/rand"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	maxListCount int = 10
	maxMapCount  int = 10

	minFloatValue float64 = -10000000000.0
	maxFloatValue float64 = 10000000000.0
)

func RandKindValue(field protoreflect.FieldDescriptor) protoreflect.Value {
	kind := field.Kind()
	switch kind {
	case protoreflect.BoolKind:
		return protoreflect.ValueOfBool(RandInt(0, 1) == 1)
	case protoreflect.EnumKind:
		enumRanges := field.Enum().Values()
		idx := rand.Intn(enumRanges.Len())
		return protoreflect.ValueOfEnum(enumRanges.Get(idx).Number())
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return protoreflect.ValueOfInt32(int32(rand.Uint32()))
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return protoreflect.ValueOfUint32(rand.Uint32())
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return protoreflect.ValueOfInt64(int64(rand.Uint64()))
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return protoreflect.ValueOfUint64(rand.Uint64())
	case protoreflect.FloatKind:
		return protoreflect.ValueOfFloat32(float32(RandFloat64(minFloatValue, maxFloatValue)))
	case protoreflect.DoubleKind:
		return protoreflect.ValueOfFloat64(RandFloat64(minFloatValue, maxFloatValue))
	case protoreflect.StringKind:
		return protoreflect.ValueOfString(string(RandBytes(0)))
	case protoreflect.BytesKind:
		return protoreflect.ValueOfBytes(RandBytes(0))
	}
	return protoreflect.Value{}
}

func RandFillMessage(msg proto.Message) {
	msgReflect := msg.ProtoReflect()
	fields := msgReflect.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		kind := field.Kind()
		if field.IsList() {
			count := RandInt(0, maxListCount)
			list := msgReflect.NewField(field).List()
			for k := 0; k < count; k++ {
				if kind == protoreflect.MessageKind {
					subMsg := list.AppendMutable().Message().Interface()
					RandFillMessage(subMsg)
				} else {
					list.Append(RandKindValue(field))
				}
			}
			msgReflect.Set(field, protoreflect.ValueOf(list))
		} else if field.IsMap() {
			pmap := msgReflect.NewField(field).Map()
			keyField := field.MapKey()
			valueField := field.MapValue()
			count := RandInt(0, maxMapCount)
			for i := 0; i < count; i++ {
				// @NOTE: protobuf map.Key: cannot be float/double, bytes or message types
				key := RandKindValue(keyField)

				if valueField.Kind() == protoreflect.MessageKind {
					value := pmap.Mutable(key.MapKey())
					RandFillMessage(value.Message().Interface())
				} else {
					value := RandKindValue(valueField)
					pmap.Set(key.MapKey(), value)
				}
			}
			msgReflect.Set(field, protoreflect.ValueOf(pmap))
		} else {
			if kind == protoreflect.MessageKind {
				subMsg := msgReflect.Mutable(field).Message().Interface()
				RandFillMessage(subMsg)
			} else {
				msgReflect.Set(field, RandKindValue(field))
			}
		}
	}
}
