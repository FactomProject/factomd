package wsapi_test

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/testHelper"
	. "github.com/FactomProject/factomd/wsapi"
	"github.com/stretchr/testify/assert"
)

func TestGetEndpoints(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()
	Start(state)

	cases := map[string]struct {
		Method   string
		Url      string
		Expected int
		Body     io.Reader
	}{
		"baseGetUrl":       {"GET", "http://localhost:8088", http.StatusNotFound, nil},
		"basePostUrl":      {"POST", "http://localhost:8088", http.StatusNotFound, body("")},
		"trailing-slashes": {"GET", "http://localhost:8088/v2/", http.StatusNotFound, nil},
		"wrong-method":     {"GET", "http://localhost:8088/v1/factoid-submit/", http.StatusNotFound, nil},
	}
	client := &http.Client{}
	for name, testCase := range cases {
		t.Logf("test case '%s'", name)
		request, err := http.NewRequest(testCase.Method, testCase.Url, testCase.Body)
		response, err := client.Do(request)

		if err != nil {
			t.Errorf("test '%s' failed: %v \nresponse: %v", name, err, response)
		} else if response == nil {
			t.Errorf("test '%s' failed: response == nil", name)
		} else if testCase.Expected != response.StatusCode {
			t.Errorf("test '%s' failed: wrong status code expected '%d' != actual '%d'", name, testCase.Expected, response.StatusCode)
		}
	}
}

func TestAuthenticatedUnauthorizedRequest(t *testing.T) {
	username := "user"
	password := "password"

	propertiesV2Body := body(primitives.NewJSON2Request("properties", 0, ""))
	globals.Params.NetworkName = "LOCAL"

	state := testHelper.CreateAndPopulateTestState()
	state.RpcUser = username
	state.RpcPass = password
	state.SetPort(18088)
	Start(state)

	cases := map[string]struct {
		Method       string
		Url          string
		Authenticate bool
		Expected     int
		Body         io.Reader
	}{
		"v1Authorized":   {"GET", "http://localhost:18088/v1/properties/", true, http.StatusOK, nil},
		"v1Unauthorized": {"GET", "http://localhost:18088/v1/properties/", false, http.StatusUnauthorized, nil},
		"v2Authorized":   {"POST", "http://localhost:18088/v2", true, http.StatusOK, propertiesV2Body},
		"v2Unauthorized": {"POST", "http://localhost:18088/v2", false, http.StatusUnauthorized, propertiesV2Body},
	}

	client := &http.Client{}
	for name, testCase := range cases {
		t.Logf("test case '%s'", name)
		request, err := http.NewRequest(testCase.Method, testCase.Url, testCase.Body)
		if testCase.Authenticate {
			request.SetBasicAuth(username, password)
		}

		response, err := client.Do(request)

		if err != nil {
			t.Errorf("test '%s' failed: %v \nresponse: %v", name, err, response)
		} else if response == nil {
			t.Errorf("test '%s' failed: response == nil", name)
		} else if testCase.Expected != response.StatusCode {
			body, _ := ioutil.ReadAll(response.Body)
			t.Errorf("test '%s' failed: wrong status code expected '%d' != actual '%d', body: %s", name, testCase.Expected, response.StatusCode, string(body))
		}
	}
}

func body(content interface{}) io.Reader {
	payload, _ := json.Marshal(content)
	body := bytes.NewBuffer(payload)
	return body
}

// use the mock state to override the GetTlsInfo
type MockState struct {
	state.State
	mockTlsInfo func() (bool, string, string)
}

func (s *MockState) GetTlsInfo() (bool, string, string) {
	return s.mockTlsInfo()
}

func TestHTTPS(t *testing.T) {
	certFile, pkFile, cleanup := testSetupCertificateFiles(t)
	defer cleanup()

	state := testHelper.CreateAndPopulateTestState()
	state.SetPort(10443)
	mState := &MockState{
		State: *state,
		mockTlsInfo: func() (bool, string, string) {
			return true, pkFile, certFile
		},
	}

	Start(mState)

	url := "https://localhost:10443/v1/heights/"
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
	}
	client := &http.Client{Transport: transCfg}
	response, err := client.Get(url)

	assert.Nil(t, err, "%v", response)
	if response != nil {
		assert.Equal(t, http.StatusOK, response.StatusCode)
	}
}

// an arbitrary self-signed certificate, generated with
// `openssl req -x509 -nodes -days 365 -newkey rsa:1024 -keyout cert.pem -out cert.pem`
var pkey = `-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQDVBUw40q0zpF/zWzwBf0GFkXmnkw+YCNTiV8l7mso1DCv/VTYM
cqtvy0g2KNBV7SFLC+NHuxJkNOAtJ8Fxx1EpeIw5A3KeCRNb4lo6ecAkuDLiPYGO
qgAqjj8QmhmZA68qTIuWGYM1FTtUK3wO4wrHnqHEjs3cWNghmby6AgLHVQIDAQAB
AoGAcy5GJINlu4KpjwBJ1dVlLD+YtA9EY0SDN0+YVglARKasM4dzjg+CuxQDm6U9
4PgzBE0NO3/fVedxP3k7k7XeH73PosaxjWpfMawXR3wSLFKJBwxux/8gNdzeGRHN
X1sYsJ70WiZLFOAPQ9jctF1ejUP6fpLHsti6ZHQj/R1xqBECQQDrHxmpMoviQL6n
4CBR4HvlIRtd4Qr21IGEXtbjIcC5sgbkfne6qhqdv9/zxsoiPTi0859cr704Mf3y
cA8LZ8c3AkEA5+/KjSoqgzPaUnvPZ0p9TNx6odxMsd5h1AMIVIbZPT6t2vffCaZ7
R0ffim/KeWfoav8u9Cyz8eJpBG6OHROT0wJBAML54GLCCuROAozePI8JVFS3NqWM
OHZl1R27NAHYfKTBMBwNkCYYZ8gHVKUoZXktQbg1CyNmjMhsFIYWTTONFNMCQFsL
eBld2f5S1nrWex3y0ajgS4tKLRkNUJ2m6xgzLwepmRmBf54MKgxbHFb9dx+dOFD4
Bvh2q9RhqhPBSiwDyV0CQBxN3GPbaa8V7eeXBpBYO5Evy4VxSWJTpgmMDtMH+RUp
9eAJ8rUyhZ2OaElg1opGCRemX98s/o2R5JtzZvOx7so=
-----END RSA PRIVATE KEY-----
`

var cert = `-----BEGIN CERTIFICATE-----
MIIDXDCCAsWgAwIBAgIJAJqbbWPZgt0sMA0GCSqGSIb3DQEBBQUAMH0xCzAJBgNV
BAYTAlVTMQswCQYDVQQIEwJDQTEWMBQGA1UEBxMNU2FuIEZyYW5jaXNjbzEPMA0G
A1UEChMGV2ViLmdvMRcwFQYDVQQDEw5NaWNoYWVsIEhvaXNpZTEfMB0GCSqGSIb3
DQEJARYQaG9pc2llQGdtYWlsLmNvbTAeFw0xMzA0MDgxNjIzMDVaFw0xNDA0MDgx
NjIzMDVaMH0xCzAJBgNVBAYTAlVTMQswCQYDVQQIEwJDQTEWMBQGA1UEBxMNU2Fu
IEZyYW5jaXNjbzEPMA0GA1UEChMGV2ViLmdvMRcwFQYDVQQDEw5NaWNoYWVsIEhv
aXNpZTEfMB0GCSqGSIb3DQEJARYQaG9pc2llQGdtYWlsLmNvbTCBnzANBgkqhkiG
9w0BAQEFAAOBjQAwgYkCgYEA1QVMONKtM6Rf81s8AX9BhZF5p5MPmAjU4lfJe5rK
NQwr/1U2DHKrb8tINijQVe0hSwvjR7sSZDTgLSfBccdRKXiMOQNyngkTW+JaOnnA
JLgy4j2BjqoAKo4/EJoZmQOvKkyLlhmDNRU7VCt8DuMKx56hxI7N3FjYIZm8ugIC
x1UCAwEAAaOB4zCB4DAdBgNVHQ4EFgQURizcvrgUl8yhIEQvJT/1b5CzV8MwgbAG
A1UdIwSBqDCBpYAURizcvrgUl8yhIEQvJT/1b5CzV8OhgYGkfzB9MQswCQYDVQQG
EwJVUzELMAkGA1UECBMCQ0ExFjAUBgNVBAcTDVNhbiBGcmFuY2lzY28xDzANBgNV
BAoTBldlYi5nbzEXMBUGA1UEAxMOTWljaGFlbCBIb2lzaWUxHzAdBgkqhkiG9w0B
CQEWEGhvaXNpZUBnbWFpbC5jb22CCQCam21j2YLdLDAMBgNVHRMEBTADAQH/MA0G
CSqGSIb3DQEBBQUAA4GBAGBPoVCReGMO1FrsIeVrPV/N6pSK7H3PLdxm7gmmvnO9
K/LK0OKIT7UL3eus+eh0gt0/Tv/ksq4nSIzXBLPKyPggLmpC6Agf3ydNTpdLQ23J
gWrxykqyLToIiAuL+pvC3Jv8IOPIiVFsY032rOqcwSGdVUyhTsG28+7KnR6744tM
-----END CERTIFICATE-----
`

func testSetupCertificateFiles(t *testing.T) (string, string, func()) {
	certificates := make([]tls.Certificate, 1)
	var err error
	certificates[0], err = tls.X509KeyPair([]byte(cert), []byte(pkey))

	if err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}

	certFile, cleanCertFile := testTempFile(t, "cert", cert)
	pkFile, cleanPKFile := testTempFile(t, "pk", pkey)

	cleanup := func() {
		cleanCertFile()
		cleanPKFile()
	}
	return certFile, pkFile, cleanup
}

func testTempFile(t *testing.T, prefix string, content string) (string, func()) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), prefix)
	if err != nil {
		t.Fatalf("error creating temp file %v", err)
	}
	defer tmpFile.Close()

	_, err = tmpFile.WriteString(content)
	if err != nil {
		t.Fatalf("error write to temp file %v", err)
	}
	cleanFile := func() { os.Remove(tmpFile.Name()) }

	return tmpFile.Name(), cleanFile
}
