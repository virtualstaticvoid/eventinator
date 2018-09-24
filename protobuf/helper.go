package protobuf

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"reflect"
	"strings"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

// Message interface for Descriptor() method
type Message interface {
	proto.Message

	Descriptor() ([]byte, []int)
}

// extracts the FileDescriptor for the given message
func getFileDescriptor(msg Message) (*descriptor.FileDescriptorProto, []int, error) {
	gzdata, path := msg.Descriptor()
	r, e := gzip.NewReader(bytes.NewReader(gzdata))
	if e != nil {
		return nil, nil, e
	}
	defer r.Close()
	b, e := ioutil.ReadAll(r)
	if e != nil {
		return nil, nil, e
	}

	fd := new(descriptor.FileDescriptorProto)
	e = proto.Unmarshal(b, fd)
	if e != nil {
		return nil, nil, e
	}
	return fd, path, nil
}

type MessageMetadata struct {
	Topic   string
	Version string
}

var (
	metaDataCache = make(map[reflect.Type]*MessageMetadata)
	lock          = sync.Mutex{}
)

func GetMessageMetadata(payload Message) (*MessageMetadata, error) {

	lock.Lock()
	defer lock.Unlock()

	key := reflect.TypeOf(payload)
	mmd := metaDataCache[key]
	if mmd != nil {
		return mmd, nil
	}

	fd, path, e := getFileDescriptor(payload)
	if e != nil {
		return nil, e
	}

	md := fd.MessageType[path[0]]
	var options = md.GetOptions()

	var topic string
	t, e := proto.GetExtension(options, E_Topic)
	if e != nil || t == nil {
		topic = strings.ToLower(reflect.TypeOf(payload).String())
	} else {
		topic = *(t.(*string))
	}

	var version string
	v, e := proto.GetExtension(options, E_Version)
	if e != nil || v == nil {
		version = "1.0"
	} else {
		version = *(v.(*string))
	}

	mmd = &MessageMetadata{
		Topic:   topic,
		Version: version,
	}

	metaDataCache[key] = mmd

	return mmd, nil

}
