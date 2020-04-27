package proto

type Message struct {
	data []byte
}

// proto.Message
// "github.com/golang/protobuf/proto"
func (m *Message) ProtoMessage() {}

func (m *Message) Reset() {
	*m = Message{}
}

func (m *Message) String() string {
	return string(m.data)
}

// json.Marshaler
// "encoding/json"
func (m *Message) MarshalJSON() ([]byte, error) {
	return m.data, nil
}

// json.Unmarshaler
// "encoding/json"
func (m *Message) UnmarshalJSON(data []byte) error {
	m.data = data
	return nil
}

func NewMessage(data []byte) *Message {
	return &Message{data}
}
