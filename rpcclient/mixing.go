package rpcclient

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/mixing"
	"github.com/btcsuite/btcd/wire"
)

// FutureGetMixPairRequestsResult is a future promise to deliver the result of a
// GetMixPairRequestsAsync RPC invocation (or an applicable error).
type FutureGetMixPairRequestsResult chan *Response

// Receive waits for the Response promised by the future and returns current set
// of mixing pair request messages from mixpool.
func (r FutureGetMixPairRequestsResult) Receive() ([]*wire.MsgMixPairReq, error) {
	res, err := ReceiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a list of strings.
	var result []string
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}

	// Convert each block hash to a wire.MsgMixPairReq and store a pointer to
	// each.
	convertedResult := make([]*wire.MsgMixPairReq, len(result))
	for i, msgHex := range result {
		var msg wire.MsgMixPairReq
		err = msg.BtcDecode(hex.NewDecoder(strings.NewReader(msgHex)), wire.MixVersion, wire.BaseEncoding)
		if err != nil {
			return nil, err
		}
		convertedResult[i] = &msg
	}

	return convertedResult, nil
}

// GetMixPairRequestsAsync returns an instance of a type that can be used to get
// the result of the RPC at some future time by invoking the Receive function on
// the returned instance.
//
// See GetMixPairRequests for the blocking version and more details.
func (c *Client) GetMixPairRequestsAsync() FutureGetMixPairRequestsResult {
	cmd := btcjson.NewGetMixPairRequestsCmd()
	return c.SendCmd(cmd)
}

// GetMixPairRequests returns the current set of mixing pair request messages
// from mixpool.
func (c *Client) GetMixPairRequests() ([]*wire.MsgMixPairReq, error) {
	return c.GetMixPairRequestsAsync().Receive()
}

// FutureGetMixMessageResult is a future promise to deliver the result of a
// GetMixMessageAsync RPC invocation (or an applicable error).
type FutureGetMixMessageResult chan *Response

// Receive waits for the Response promised by the future and returns data about
// the current network.
func (r FutureGetMixMessageResult) Receive() (mixing.Message, error) {
	res, err := ReceiveFuture(r)
	if err != nil {
		return nil, err
	}

	var result btcjson.GetMixMessageResult
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}

	return mixMessage(result.Type, result.Message)
}

// GetMixMessageAsync returns an instance of a type that can be used to get the
// result of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See GetMixMessage for the blocking version and more details.
func (c *Client) GetMixMessageAsync(hash string) FutureGetMixMessageResult {
	cmd := btcjson.NewGetMixMessageCmd(hash)
	return c.SendCmd(cmd)
}

// GetMixMessage returns data about the current network.
func (c *Client) GetMixMessage(hash string) (mixing.Message, error) {
	return c.GetMixMessageAsync(hash).Receive()
}

func mixMessage(command, msgHex string) (mixing.Message, error) {
	var msg mixing.Message
	switch command {
	case wire.CmdMixPairReq:
		msg = new(wire.MsgMixPairReq)
	case wire.CmdMixKeyExchange:
		msg = new(wire.MsgMixKeyExchange)
	case wire.CmdMixCiphertexts:
		msg = new(wire.MsgMixCiphertexts)
	case wire.CmdMixSlotReserve:
		msg = new(wire.MsgMixSlotReserve)
	case wire.CmdMixDCNet:
		msg = new(wire.MsgMixDCNet)
	case wire.CmdMixConfirm:
		msg = new(wire.MsgMixConfirm)
	case wire.CmdMixFactoredPoly:
		msg = new(wire.MsgMixFactoredPoly)
	case wire.CmdMixSecrets:
		msg = new(wire.MsgMixSecrets)
	default:
		return nil, fmt.Errorf("unrecognized mixing message command string: %v", command)
	}

	err := msg.BtcDecode(hex.NewDecoder(strings.NewReader(msgHex)), wire.MixVersion, wire.BaseEncoding)
	return msg, err
}

// FutureSendRawMixMessageResult is a future promise to deliver the result
// of a SendRawMixMessageAsync RPC invocation (or an applicable error).
type FutureSendRawMixMessageResult chan *Response

// Receive waits for the Response promised by the future and returns an error if
// submitting a mixing message to the mixpool was not successful.
func (r FutureSendRawMixMessageResult) Receive() error {
	_, err := ReceiveFuture(r)
	return err
}

// SendRawMixMessageAsync returns an instance of a type that can be used to get
// the result of the RPC at some future time by invoking the Receive function on
// the returned instance.
//
// See SendRawMixMessage for the blocking version and more details.
func (c *Client) SendRawMixMessageAsync(msg mixing.Message) FutureSendRawMixMessageResult {
	var b strings.Builder
	err := msg.BtcEncode(hex.NewEncoder(&b), wire.MixVersion, wire.BaseEncoding)
	if err != nil {
		return newFutureError(err)
	}

	cmd := btcjson.NewSendRawMixMessageCmd(msg.Command(), b.String())
	return c.SendCmd(cmd)
}

// SendRawMixMessage submits the encoded transaction to the server which will
// then relay it to the network.
func (c *Client) SendRawMixMessage(msg mixing.Message) error {
	return c.SendRawMixMessageAsync(msg).Receive()
}
