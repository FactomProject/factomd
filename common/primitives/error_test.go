// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives_test

import (
	"testing"

	. "github.com/PaulSnow/factom2d/common/primitives"
)

func TestErrors(t *testing.T) {
	errs := []Error{
		Error{ErrorInternal, 500, "Internal", "An internal error occurred", "", ""},
		Error{ErrorJSONMarshal, 500, "JSON Marshal", "An error occurred marshalling into JSON", "", ""},
		Error{ErrorXMLMarshal, 500, "XML Marshal", "An error occurred marshalling into XML", "", ""},
		Error{ErrorUnsupportedMarshal, 500, "Unsupported Marshal", "The server attempted to marshal the data into an unsupported format", "", ""},
		Error{ErrorBadMethod, 405, "Bad Method", "The specified method cannot be used on the specified resource", "", ""},
		Error{ErrorNotAcceptable, 406, "Not Acceptable", "The resource cannot be retreived as any of the acceptable types", "", ""},
		Error{ErrorMissingVersionSpec, 400, "Missing Version Spec", "The API version specifier is missing from the request URL", "", ""},
		Error{ErrorMalformedVersionSpec, 400, "Malformed Version Spec", "The API version specifier is malformed", "", ""},
		Error{ErrorBadVersionSpec, 400, "Bad Version Spec", "The API version specifier specifies a bad version", "", ""},
		Error{ErrorEmptyRequest, 200, "Empty Request", "The request is empty", "", ""},
		Error{ErrorBadElementSpec, 400, "Bad Element Spec", "The element specifier is bad", "", ""},
		Error{ErrorBadIdentifier, 400, "Bad Identifier", "The element identifier was malformed", "", ""},
		Error{ErrorBlockNotFound, 404, "Block Not Found", "The specified block cannot be found", "", ""},
		Error{ErrorEntryNotFound, 404, "Entry Not Found", "The specified entry cannot be found", "", ""},
		Error{ErrorJSONUnmarshal, 400, "JSON Unmarshal", "An error occurred while unmarshalling from JSON", "", ""},
		Error{ErrorXMLUnmarshal, 400, "XML Unmarshal", "An error occurred while unmarshalling from XML", "", ""},
		Error{ErrorUnsupportedUnmarshal, 400, "Unsupported Unmarshal", "The data was specified to be in an unsupported format", "", ""},
		Error{ErrorBadPOSTData, 400, "Bad POST Data", "The body of the POST request is malformed", "", ""},
		Error{ErrorTemplateError, 500, "Template Error", "A template error occurred", "", ""},
		Error{ErrorHTTPNewRequestFailure, 500, "HTTP Request Failure", "Failed to create an HTTP request", "", ""},
		Error{ErrorHTTPDoRequestFailure, 500, "HTTP Request Failure", "Error while executing an HTTP request", "", ""},
		Error{ErrorHTMLMarshal, 500, "HTML Marshal", "An error occurred marshalling into HTML", "", ""},
		Error{9999, 500, "Unknown Error", "An unknown error occurred", "", ""},
	}

	for _, e := range errs {
		e2 := CreateError(uint(e.APICode), "message")

		if e.APICode != e2.APICode {
			t.Errorf("Invalid APICode - %v vs %v", e.APICode, e2.APICode)
		}
		if e.HTTPCode != e2.HTTPCode {
			t.Errorf("Invalid HTTPCode - %v vs %v", e.HTTPCode, e2.HTTPCode)
		}
		if e.Name != e2.Name {
			t.Errorf("Invalid Name - %v vs %v", e.Name, e2.Name)
		}
		if e.Description != e2.Description {
			t.Errorf("Invalid Description - %v vs %v", e.Description, e2.Description)
		}
		if e.SupportURL != e2.SupportURL {
			t.Errorf("Invalid SupportURL - %v vs %v", e.SupportURL, e2.SupportURL)
		}

		if e2.Message != "message" {
			t.Errorf("Invalid Message - %v", e.Message)
		}
	}
}
