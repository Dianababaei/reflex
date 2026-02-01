package reflex

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"

	"github.com/xtls/xray-core/common/errors"
	"github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/common/uuid"
)

// Account represents a Reflex user account (from proto)
type Account struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id     string
	Policy string
}

// ProtoReflect implements proto.Message.ProtoReflect
func (x *Account) ProtoReflect() protoreflect.Message {
	return nil // Minimal implementation, not used for testing
}

// Reset implements proto.Message.Reset
func (x *Account) Reset() {
	*x = Account{}
}

// String implements proto.Message.String
func (x *Account) String() string {
	return "Account{" + x.Id + "," + x.Policy + "}"
}

// AsAccount implements protocol.Account.AsAccount().
func (a *Account) AsAccount() (protocol.Account, error) {
	id, err := uuid.ParseString(a.Id)
	if err != nil {
		return nil, errors.New("failed to parse ID: ", err)
	}
	return &MemoryAccount{
		ID:     protocol.NewID(id),
		Policy: a.Policy,
	}, nil
}

// MemoryAccount is an in-memory form of Reflex account.
type MemoryAccount struct {
	ID     *protocol.ID
	Policy string
}

// Equals implements protocol.Account.Equals().
func (a *MemoryAccount) Equals(account protocol.Account) bool {
	reflexAccount, ok := account.(*MemoryAccount)
	if !ok {
		return false
	}
	return a.ID.Equals(reflexAccount.ID)
}

// ToProto converts MemoryAccount to Account (implements proto.Message)
func (a *MemoryAccount) ToProto() proto.Message {
	return &Account{
		Id:     a.ID.String(),
		Policy: a.Policy,
	}
}
