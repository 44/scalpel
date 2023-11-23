package odl

type EventHeaderV2 struct {
	Magic uint32
	Type uint32
	Timestamp uint64
	ProcessID uint32
	ThreadID uint32
	Schema EventSchema
	PayloadSize uint32
	TraceID uint32
}

const EventType_LogEvent = 0
const EventType_LogEvent = 1

type EventHeaderV3 struct {
	Magic uint32
	Signature uint32
	Timestamp uint64
	ProcessID uint32
	ThreadID uint32
	PayloadSize uint32
	TraceID uint32
}

const EventMagicValue = 0xffeeddcc

type EventSchema struct {
	ProviderID [16]byte
	EventID uint32
	EventVersion uint32
}

const DataType_Unknown = 0
const DataType_UnicodeString = 1
const DataType_AnsiString = 2
const DataType_Int32 = 3
const DataType_Int64 = 4
const DataType_UInt32 = 5
const DataType_UInt64 = 6
const DataType_Boolean = 7
const DataType_Float = 8
const DataType_Hash = 9
const DataType_CobaltHash = 10
const DataType_PrivateInfo = 11

