// +build all

package wsapi_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/wsapi"
)

func getResp(err *primitives.JSONError) string {
	resp := primitives.NewJSON2Response()
	resp.ID = nil
	resp.Error = err

	js := resp.String()

	var buf bytes.Buffer
	json.Indent(&buf, []byte(js), "", "\t")
	return string(buf.Bytes())
}

func TestErrors(t *testing.T) {
	je := NewParseError()
	if je.Code != -32700 || je.Message != "Parse error" {
		t.Error("Code or message is wrong for parse error")
	}

	je = NewInvalidRequestError()
	if je.Code != -32600 || je.Message != "Invalid Request" {
		t.Error("Code or message is wrong for NewInvalidRequestError")
	}

	je = NewMethodNotFoundError()
	if je.Code != -32601 || je.Message != "Method not found" {
		t.Error("Code or message is wrong for NewMethodNotFoundError")
	}

	je = NewInvalidParamsError()
	if je.Code != -32602 || je.Message != "Invalid params" {
		t.Error("Code or message is wrong for NewInvalidParamsError")
	}

	je = NewInternalError()
	if je.Code != -32603 || je.Message != "Internal error" {
		t.Error("Code or message is wrong for NewInternalError")
	}

	je = NewCustomInternalError(nil)
	if je.Code != -32603 || je.Message != "Internal error" {
		t.Error("Code or message is wrong for NewCustomInternalError")
	}

	je = NewCustomInvalidParamsError(nil)
	if je.Code != -32602 || je.Message != "Invalid params" {
		t.Error("Code or message is wrong for NewCustomInvalidParamsError")
	}

	je = NewInvalidAddressError()
	if je.Code != -32602 || je.Message != "Invalid params" {
		t.Error("Code or message is wrong for NewInvalidAddressError")
	}

	je = NewUnableToDecodeTransactionError()
	if je.Code != -32602 || je.Message != "Invalid params" {
		t.Error("Code or message is wrong for NewUnableToDecodeTransactionError")
	}

	je = NewInvalidTransactionError()
	if je.Code != -32602 || je.Message != "Invalid params" {
		t.Error("Code or message is wrong for NewInvalidTransactionError")
	}

	je = NewInvalidHashError()
	if je.Code != -32602 || je.Message != "Invalid params" {
		t.Error("Code or message is wrong for NewInvalidHashError")
	}

	je = NewInvalidEntryError()
	if je.Code != -32602 || je.Message != "Invalid params" {
		t.Error("Code or message is wrong for NewInvalidEntryError")
	}

	je = NewInvalidCommitChainError()
	if je.Code != -32602 || je.Message != "Invalid params" {
		t.Error("Code or message is wrong for NewInvalidCommitChainError")
	}

	je = NewInvalidCommitEntryError()
	if je.Code != -32602 || je.Message != "Invalid params" {
		t.Error("Code or message is wrong for NewInvalidCommitEntryError")
	}

	je = NewInvalidDataPassedError()
	if je.Code != -32602 || je.Message != "Invalid params" {
		t.Error("Code or message is wrong for NewInvalidDataPassedError")
	}

	je = NewInternalDatabaseError()
	if je.Code != -32603 || je.Message != "Internal error" {
		t.Error("Code or message is wrong for NewInternalDatabaseError")
	}

	je = NewBlockNotFoundError()
	if je.Code != -32008 || je.Message != "Block not found" {
		t.Error("Code or message is wrong for NewBlockNotFoundError")
	}

	je = NewEntryNotFoundError()
	if je.Code != -32008 || je.Message != "Entry not found" {
		t.Error("Code or message is wrong for NewEntryNotFoundError")
	}

	je = NewObjectNotFoundError()
	if je.Code != -32008 || je.Message != "Object not found" {
		t.Error("Code or message is wrong for NewObjectNotFoundError")
	}

	je = NewMissingChainHeadError()
	if je.Code != -32009 || je.Message != "Missing Chain Head" {
		t.Error("Code or message is wrong for NewMissingChainHeadError")
	}

	je = NewReceiptError()
	if je.Code != -32010 || je.Message != "Receipt creation error" {
		t.Error("Code or message is wrong for NewReceiptError")
	}

	je = NewRepeatCommitError("")
	if je.Code != -32011 || je.Message != "Repeated Commit" {
		t.Error("Code or message is wrong for NewReceiptError")
	}

	fmt.Println(getResp(je))

}
